package token

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"golang.org/x/oauth2"
)

type FileStorage struct {
	path string
	lock sync.RWMutex
}

func NewFileStorage(path string) (*FileStorage, error) {
	return &FileStorage{path: path}, nil
}

func (f *FileStorage) LoadToken() (*oauth2.Token, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	content, err := ioutil.ReadFile(f.path)
	switch {
	case os.IsNotExist(err):
		return nil, nil
	case err != nil:
		return nil, err
	}
	var t oauth2.Token
	if err := json.Unmarshal(content, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (f *FileStorage) SaveToken(t *oauth2.Token) error {
	content, err := json.Marshal(t)
	if err != nil {
		return err
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	return ioutil.WriteFile(f.path, content, 0600)
}
