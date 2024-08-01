package google

import "math/rand"

type Credential interface {
	Token() (string, error)
}

type ApiKey struct {
	keys   []string
	next   int
	random bool
}

func NewApiKey(key string) *ApiKey {
	return &ApiKey{keys: []string{key}, random: false}
}

func RandomApiKeys(keys ...string) *ApiKey {
	return &ApiKey{keys: keys, random: true}
}

func RotatingApiKeys(keys ...string) *ApiKey {
	return &ApiKey{keys: keys, random: false}
}

// Token returns the API key.
// Renturns a random API key if multiple keys added with [NewApiKeys].
func (k *ApiKey) Token() (string, error) {

	if len(k.keys) == 0 {
		return "", nil
	}

	if len(k.keys) == 1 {
		return k.keys[0], nil
	}

	if k.random {
		k.next = rand.Intn(len(k.keys))
	} else {
		k.next += 1
		if k.next > len(k.keys)-1 {
			k.next = 0
		}
	}

	return k.keys[k.next], nil
}
