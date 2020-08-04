package siga

import (
	"context"

	"github.com/pkg/errors"

	"stash.ria.ee/vis3/vis3-common/pkg/ignite"
	"stash.ria.ee/vis3/vis3-common/pkg/log"
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

// igniteStorage implements storage using Apache Ignite as the storage service.
type igniteStorage struct {
	client      *ignite.Client
	statusCache string
	dataCache   string
}

func newIgniteStorage(ctx context.Context, conf Conf) (storage, error) {
	client, err := ignite.New(ctx, conf.Ignite.Conf)
	if err != nil {
		return nil, errors.WithMessage(err, "ignite")
	}
	storage := igniteStorage{
		client:      client,
		statusCache: conf.Ignite.StatusCache,
		dataCache:   conf.Ignite.DataCache,
	}
	if err := storage.client.CheckCacheNames(context.Background(),
		storage.statusCache, storage.dataCache); err != nil {

		storage.close(ctx)
		return nil, errors.WithMessage(err, "ignite caches")
	}
	return storage, nil
}

func (s igniteStorage) putStatus(ctx context.Context, session string, status status) error {
	log.Debug().
		WithString("session", session).
		WithString("containerID", status.containerID).
		WithString("signatureID", status.signatureID).
		Log(ctx, "request")
	return errors.WithMessage(
		s.client.CachePut(ctx, s.statusCache, session, status.toComplex()),
		"ignite")
}

func (s igniteStorage) getStatus(ctx context.Context, session string, mandatory bool) (*status, error) {
	log.Debug().WithString("session", session).Log(ctx, "request")
	value, err := s.client.CacheGet(ctx, s.statusCache, session)
	if err != nil {
		return nil, errors.WithMessage(err, "ignite")
	}
	if value == nil {
		if mandatory {
			err = errors.Errorf("ignite: no open container for %s", session)
		}
		return nil, err
	}

	status, ok := fromComplex(value)
	if !ok {
		return nil, errors.Errorf("ignite: unexpected status value type, cache corrupted?")
	}
	return &status, nil
}

func (s igniteStorage) removeStatus(ctx context.Context, session string) error {
	log.Debug().WithString("session", session).Log(ctx, "request")
	_, err := s.client.CacheRemoveKey(ctx, s.statusCache, session)
	return errors.WithMessage(err, "ignite")
}

func (s igniteStorage) putData(ctx context.Context, key string, data []byte) error {
	log.Debug().WithString("key", key).Log(ctx, "request")
	return errors.WithMessage(
		s.client.CachePut(ctx, s.dataCache, key, data),
		"ignite")
}

func (s igniteStorage) getData(ctx context.Context, key string) ([]byte, error) {
	log.Debug().WithString("key", key).Log(ctx, "request")
	value, err := s.client.CacheGet(ctx, s.dataCache, key)
	if err != nil {
		return nil, errors.WithMessage(err, "ignite")
	}
	if value == nil {
		return nil, errors.Errorf("ignite: no data for %s", key)
	}
	data, ok := value.([]byte)
	if !ok {
		return nil, errors.Errorf("ignite: unexpected data value type, cache corrupted?")
	}
	return data, nil
}

func (s igniteStorage) removeData(ctx context.Context, key string) error {
	log.Debug().WithString("key", key).Log(ctx, "request")
	_, err := s.client.CacheRemoveKey(ctx, s.dataCache, key)
	return errors.WithMessage(err, "ignite")
}

func (s igniteStorage) close(ctx context.Context) error {
	return errors.WithMessage(s.client.Close(ctx), "ignite")
}

// status is the state of an open container.
type status struct {
	containerID string
	filenames   []string
	signatureID string
}

func fromComplex(value interface{}) (status, bool) {
	complex, ok := ignite.FromComplexObject(value)
	if !ok || !complex.Is("vis3-siga-container") {
		return status{}, false
	}
	return status{
		containerID: complex.GetOrDefault("containerID", "").(string),
		filenames:   complex.GetOrDefault("filenames", []string{}).([]string),
		signatureID: complex.GetOrDefault("signatureID", "").(string),
	}, true
}

func (s status) toComplex() interface{} {
	complex := ignite.NewComplexObject("vis3-siga-container")
	complex.Set("containerID", s.containerID)
	complex.Set("filenames", s.filenames)
	complex.Set("signatureID", s.signatureID)
	return complex.Value()
}
