package mods

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/K4rian/kfdsl/internal/log"
	"github.com/K4rian/kfdsl/internal/utils"
)

type Author struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}

type InstallItem struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Checksum string `json:"checksum,omitempty"`
}

type Mod struct {
	Version      string        `json:"version"`
	Description  string        `json:"description"`
	Authors      []Author      `json:"authors"`
	License      string        `json:"license"`
	ProjectURL   string        `json:"project_url"`
	DownloadURL  string        `json:"download_url"`
	Checksum     string        `json:"checksum,omitempty"`
	Extract      bool          `json:"extract"`
	InstallItems []InstallItem `json:"install"`
	DependOn     []string      `json:"depend_on"`
}

var mu sync.Mutex

func ParseModsFile(filename string) (map[string]*Mod, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var items map[string]*Mod
	err = json.NewDecoder(jsonFile).Decode(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (m *Mod) installMods(dir string, deps map[string]*Mod, installed map[string]bool) error {
	for _, name := range m.DependOn {
		dep, ok := deps[name]
		if !ok {
			return fmt.Errorf("mod %s not found", name)
		}
		if err := dep.Install(dir, name, deps, installed); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mod) isDownloadRequired(dir string) bool {
	for _, item := range m.InstallItems {
		if item.Checksum == "" {
			continue
		}

		itemPath := filepath.Join(dir, item.Path, item.Name)
		if utils.FileExists(itemPath) {
			if match, err := utils.FileMatchesChecksum(itemPath, item.Checksum); err != nil || !match {
				log.Logger.Debug("Checksum mismatch, download required", "path", itemPath, "checksum", item.Checksum)
				return true
			}
		}
	}

	return false
}

func (m *Mod) download(dir, name string) (string, error) {
	if !m.isDownloadRequired(dir) {
		return "", nil
	}

	log.Logger.Debug("Downloading mod", "name", name, "url", m.DownloadURL)
	filename, err := utils.DownloadFile(m.DownloadURL, m.Checksum)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", name, err)
	}
	log.Logger.Debug("Mod download complete", "name", name)
	return filename, nil
}

func (m *Mod) installFile(dir, filename string, item InstallItem) error {
	log.Logger.Debug("Installing mod file", "name", item.Name, "dir", dir, "path", item.Path, "from", filename)
	path, err := utils.CreateDirIfNotExists(dir, item.Path)
	if err != nil {
		return err
	}
	return utils.MoveFile(filename, filepath.Join(path, item.Name), item.Checksum)
}

func (m *Mod) installFiles(dir, filename string) error {
	if len(m.InstallItems) > 1 && !m.Extract {
		return fmt.Errorf("mod contains multiple files but is not marked for extraction")
	}

	log.Logger.Debug("Installing mod files")
	if len(m.InstallItems) == 1 {
		return m.installFile(dir, filename, m.InstallItems[0])
	}
	return m.installArchive(dir, filename)
}

func (m *Mod) installArchive(dir, archive string) error {
	// Unpack item in temporary directory then move them one by one
	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	log.Logger.Debug("Extracting mod archive", "archive", archive, "to", tempDir)
	if err := utils.UnzipFile(archive, tempDir); err != nil {
		return err
	}

	for _, item := range m.InstallItems {
		if err := m.installFile(dir, filepath.Join(tempDir, item.Name), item); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mod) Install(dir string, name string, deps map[string]*Mod, installed map[string]bool) error {
	log.Logger.Debug("Installing mod", "name", name)

	mu.Lock()
	if alreadyInstalled, ok := installed[name]; ok && alreadyInstalled {
		log.Logger.Debug("Mod already installed, no actions needed", "name", name)
		return nil
	}
	mu.Unlock()

	if err := m.installMods(dir, deps, installed); err != nil {
		return err
	}

	filename, err := m.download(dir, name)
	if err != nil {
		return err
	}

	if filename != "" {
		if err := m.installFiles(dir, filename); err != nil {
			return err
		}
	}

	mu.Lock()
	installed[name] = true
	mu.Unlock()

	return nil
}
