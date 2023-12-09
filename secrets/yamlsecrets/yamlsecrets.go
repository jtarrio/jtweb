package yamlsecrets

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type yamlSecrets struct {
	Secrets map[string]string
}

func Parse(content []byte) (*yamlSecrets, error) {
	var secrets = &yamlSecrets{}
	err := yaml.UnmarshalStrict(content, secrets)
	return secrets, err
}

func (s *yamlSecrets) GetSecret(key string) (string, error) {
	secret, ok := s.Secrets[key]
	if !ok {
		return "", fmt.Errorf("no secret found for key '%s'", key)
	}
	return secret, nil
}
