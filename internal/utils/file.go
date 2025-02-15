package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(srcPath, destPath string) error {
	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return destinationFile.Sync()
}

func MoveFile(srcPath, dstPath string) error {
	if err := CopyFile(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := os.Remove(srcPath); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

func CopyAndReplaceFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create/open destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	err = os.Chmod(destPath, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	return nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func DownloadFile(url string) (string, error) {
	filename := filepath.Join(os.TempDir(), url[strings.LastIndex(url, "/")+1:])

	out, err := os.Create(filename)
	if err != nil {
		return filename, err
	}
	defer out.Close()

	response, err := http.Get(url)
	if err != nil {
		return filename, err
	}
	defer response.Body.Close()

	_, err = io.Copy(out, response.Body)

	return filename, err
}

func UnzipFile(source, destination string) error {
	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(destination, os.ModeDir)

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(destination, f.Name)

		// Check for ZipSlip (directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	if err := os.Remove(source); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}
