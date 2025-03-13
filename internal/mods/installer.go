package mods

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Enabled      bool          `json:"enabled,omitempty"`
}

type installError struct {
	name string
	err  error
}

var mu sync.Mutex

func (m *Mod) isDownloadRequired(dir string) bool {
	for _, item := range m.InstallItems {
		if item.Checksum == "" {
			continue
		}

		itemPath := filepath.Join(dir, item.Path, item.Name)
		log.Logger.Debug("Checking mod file", "path", itemPath, "checksum", item.Checksum)

		if !utils.FileExists(itemPath) {
			return true
		}

		if match, err := utils.FileMatchesChecksum(itemPath, item.Checksum); err != nil || !match {
			log.Logger.Debug("Checksum mismatch, download required", "path", itemPath, "checksum", item.Checksum)
			return true
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

func (m *Mod) install(dir string, name string) error {
	if !m.Enabled {
		log.Logger.Debug("Skipping installation of mod, it is disabled", "name", name)
		return nil
	}

	if !m.isDownloadRequired(dir) {
		log.Logger.Debug("Skipping installation of mod, it is already installed", "name", name)
		return nil
	}

	log.Logger.Debug("Installing mod", "name", name)

	filename, err := m.download(dir, name)
	if err != nil {
		return err
	}

	if filename != "" {
		if err := m.installFiles(dir, filename); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mod) resolveDependencies(mods map[string]*Mod) []string {
	if m.DependOn == nil {
		return nil
	}

	deps := make([]string, 0)
	for _, name := range m.DependOn {
		if dep, ok := mods[name]; ok {
			// Enable mod for installation if it is a dependency
			dep.Enabled = true
			deps = append(deps, dep.resolveDependencies(mods)...)
		} else {
			log.Logger.Warn("Dependency not found", "name", name)
		}
	}
	return deps
}

func resolveModsToInstall(mods map[string]*Mod) []string {
	m := make([]string, 0)
	for name, mod := range mods {
		m = append(m, name)
		m = append(m, mod.resolveDependencies(mods)...)
	}
	return utils.RemoveDuplicates(m)
}

func InstallMods(wg *sync.WaitGroup, dir string, mods map[string]*Mod, installed map[string]bool) error {
	toInstall := resolveModsToInstall(mods)
	log.Logger.Debug("Mods to install", "mods", strings.Join(toInstall, " / "))

	installations := make(chan string, len(toInstall))
	errors := make(chan installError, len(toInstall))

	for _, name := range toInstall {
		mod := mods[name]

		wg.Add(1)
		go func(name string, mod *Mod) {
			defer wg.Done()

			err := mod.install(dir, name)
			if err != nil {
				errors <- installError{name, err}
			} else {
				installations <- name
			}
		}(name, mod)
	}

	go func() {
		for installedMod := range installations {
			mu.Lock()
			installed[installedMod] = true
			mu.Unlock()
		}
	}()

	go func() {
		for err := range errors {
			log.Logger.Error("Failed to install mod", "name", err.name, "error", err.err)
		}
	}()

	wg.Wait()
	close(installations)
	close(errors)

	return nil
}

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
