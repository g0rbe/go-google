package google

import (
	"math/rand"
	"sync"
)

type Credential interface {
	Token() (string, error)
}

type ApiKey struct {
	keys   []string
	next   int
	random bool
	m      *sync.Mutex
}

func NewApiKey(key string) *ApiKey {
	return &ApiKey{keys: []string{key}, random: false, m: new(sync.Mutex)}
}

func RandomApiKeys(keys ...string) *ApiKey {
	return &ApiKey{keys: keys, random: true, m: new(sync.Mutex)}
}

func RotatingApiKeys(keys ...string) *ApiKey {
	return &ApiKey{keys: keys, random: false, m: new(sync.Mutex)}
}

// Token returns the API key.
// Renturns a random API key if multiple keys added with [NewApiKeys].
func (k *ApiKey) Token() (string, error) {

	k.m.Lock()
	defer k.m.Unlock()

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
