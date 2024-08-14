package pkg

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"golang.org/x/oauth2"
)

type Storer interface {
	Store(storeToken StoreToken) error
	Retrieve() []byte
	ToToken(data []byte) (oauth2.Token, error)
}

type storer struct {
	path string
}

func NewStorer(path string) Storer {
	return &storer{
		path: path,
	}
}

func (s *storer) Store(storeToken StoreToken) error {
	storeTokenBytes, err := json.Marshal(storeToken)
	if err != nil {
		return err
	}

	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(storeTokenBytes)
	if err != nil {
		return err
	}
	return nil
}

func (s *storer) Retrieve() []byte {
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		log.Fatalf("Error: %v", err)
	}
	defer f.Close()
	b := make([]byte, 1500)
	i, err := f.Read(b)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return b[:i]
		}
		log.Fatalf("Error: %v", err)
	}

	return b[:i]
}

func (s *storer) ToToken(data []byte) (oauth2.Token, error) {
	token := oauth2.Token{}
	err := json.Unmarshal(data, &token)

	return token, err
}
