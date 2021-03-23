package storage

import (
	"github.com/dgraph-io/badger/v3"
)

// Backend wraps the two basic methods to interact with a key-value store.
type Backend interface {
	Get(key string) (string, error)
	Put(key, value string) error
}

// BadgerBackend is a backend implemented using BadgerDB.
type BadgerBackend struct {
	db *badger.DB
}

// NewBadgerBackend returns an initialized backend, opening (or creating, if
// not present) a badgerDB from the specified path.
func NewBadgerBackend(path string) (*BadgerBackend, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &BadgerBackend{db}, nil
}

// Get takes a key as a string and returns the associated value saved in
// the storage, if present.
// If something goes wrong, it returns a non-nil error.
func (b *BadgerBackend) Get(key string) (string, error) {
	var url string

	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		if err := item.Value(func(value []byte) error {
			url = string(value)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return url, nil
}

// Put takes a key-value pair in input and saves them in the underlying
// storage. It returns an error if something goes wrong.
func (b *BadgerBackend) Put(key, value string) error {
	if err := b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(key), []byte(value)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Close closes the underlying badgerDB.
func (b *BadgerBackend) Close() error {
	return b.db.Close()
}
