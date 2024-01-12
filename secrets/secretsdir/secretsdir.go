package secretsdir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type secretsDir struct {
	path string
}

func Create(path string) *secretsDir {
	return &secretsDir{path: path}
}

func (s *secretsDir) GetSecret(key string) (string, error) {
	if strings.Contains(key, "/") || strings.Contains(key, string(filepath.Separator)) {
		return "", fmt.Errorf("secret key must not contain a path separator: %s", key)
	}
	f, err := os.ReadFile(filepath.Join(s.path, key))
	if err != nil {
		return "", err
	}
	return string(f), nil
}
