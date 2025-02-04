package secrets

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1" // TODO: use internal ini parser
)

func Read(secretName string) (map[string]string, error) {
	secretPath := filepath.Join("/run/secrets", secretName)

	if _, err := os.Stat(secretPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("secret file '%s' does not exist", secretPath)
	}

	cfg, err := ini.Load(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret file '%s': %v", secretPath, err)
	}

	secrets := make(map[string]string)
	for _, section := range cfg.Sections() {
		for _, key := range section.Keys() {
			secrets[key.Name()] = key.String()
		}
	}
	return secrets, nil
}
