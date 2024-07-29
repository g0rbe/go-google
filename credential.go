package google

type Credential interface {
	Token() (string, error)
}

type ApiKey struct {
	key string
}

func NewApiKey(key string) *ApiKey {
	return &ApiKey{key: string(key)}
}

func (k *ApiKey) Token() (string, error) {
	if k == nil {
		return "", nil
	}
	return k.key, nil
}
