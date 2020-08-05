package siga

import (
	"context"

	"github.com/pkg/errors"
)

// storage is an internal interface for storing the state of open containers.
// It allows us to provide an alternative mock implementation for testing.
type storage interface {
	putStatus(ctx context.Context, session string, status status) error
	getStatus(ctx context.Context, session string, mandatory bool) (*status, error)
	removeStatus(ctx context.Context, session string) error

	putData(ctx context.Context, key string, contents []byte) error
	getData(ctx context.Context, key string) ([]byte, error)
	removeData(ctx context.Context, key string) error

	close(ctx context.Context) error
}

// newMemStorage moodustab SiGa-ga suhtlemiseks vajaliku mälustruktuuri.
func newMemStorage() storage {
	return memStorage{
		status: make(map[string]status),
		data:   make(map[string][]byte),
	}
}

// status is the state of an open container.
type status struct {
	containerID string
	filenames   []string
	signatureID string
}	

// memStorage implements storage in memory for testing.
type memStorage struct {
	status map[string]status
	data   map[string][]byte
}	

func (s memStorage) putStatus(ctx context.Context, session string, status status) error {
	s.status[session] = status
	return nil
}

func (s memStorage) getStatus(ctx context.Context, session string, mandatory bool) (*status, error) {
	status, ok := s.status[session]
	if !ok {
		if mandatory {
			return nil, errors.Errorf("memory: no open container for %s", session)
		}
		return nil, nil
	}
	return &status, nil
}

func (s memStorage) removeStatus(ctx context.Context, session string) error {
	delete(s.status, session)
	return nil
}

func (s memStorage) putData(ctx context.Context, key string, data []byte) error {
	s.data[key] = data
	return nil
}

func (s memStorage) getData(ctx context.Context, key string) ([]byte, error) {
	data, ok := s.data[key]
	if !ok {
		return nil, errors.Errorf("memory: no data for %s", key)
	}
	return data, nil
}

func (s memStorage) removeData(ctx context.Context, key string) error {
	delete(s.data, key)
	return nil
}

func (s memStorage) close(ctx context.Context) error {
	return nil
}
