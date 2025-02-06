package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Read(secretName string) (string, error) {
	secretPath := filepath.Join("/run/secrets", secretName)
	data, err := os.ReadFile(secretPath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret '%s': %w", secretName, err)
	}
	return strings.TrimSpace(string(data)), nil
}
