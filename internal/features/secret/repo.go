package secret

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type SecretRepo struct {
	db *badger.DB
}

func NewSecretRepo(db *badger.DB) *SecretRepo {
	return &SecretRepo{db: db}
}

func (r *SecretRepo) Create(secret *Secret) error {
	return r.db.Update(func(txn *badger.Txn) error {
		data, _ := json.Marshal(secret)
		entry := badger.NewEntry([]byte(secret.Short), data).WithTTL(time.Until(secret.ExpiresAt))
		return txn.SetEntry(entry)
	})
}

func (r *SecretRepo) GetByShort(short string) (*Secret, error) {
	var secret Secret

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(short))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &secret)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}

	return &secret, err
}
func (r *SecretRepo) Delete(short string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(short))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}
