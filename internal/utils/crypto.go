package utils

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func SHA1File(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func SHA1Compare(file1, file2 string) (bool, error) {
	hash1, err := SHA1File(file1)
	if err != nil {
		return false, err
	}

	hash2, err := SHA1File(file2)
	if err != nil {
		return false, err
	}
	return hash1 == hash2, nil
}
