package persist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	once    sync.Once
	mu      sync.Mutex
	loadErr error
	store   map[string]string
)

func GetPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	dir = filepath.Join(dir, "gdqgo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		panic(err)
	}

	return filepath.Join(dir, "persist.json")
}

func ensureLoaded() {
	once.Do(func() {
		path := GetPath()
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				store = map[string]string{}
				return
			}
			loadErr = err
			return
		}

		if err := json.Unmarshal(b, &store); err != nil {
			loadErr = err
			return
		}

		if store == nil {
			store = map[string]string{}
		}
	})
}

func Get(key string) (string, bool, error) {
	ensureLoaded()
	if loadErr != nil {
		return "", false, loadErr
	}

	mu.Lock()
	defer mu.Unlock()
	v, ok := store[key]
	return v, ok, nil
}

func GetExpected(key string) (string, error) {
	ensureLoaded()
	s, ok, err := Get(key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("key %s missing from persist.json", key)
	}
	return s, nil
}

func Set(key string, value string) error {
	ensureLoaded()
	if loadErr != nil {
		return loadErr
	}

	mu.Lock()
	store[key] = value
	b, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	path := GetPath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
