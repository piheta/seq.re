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

func (r *SecretRepo) CountSecrets() (total int, err error) {
	return total, r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			if len(item.Key()) != 6 {
				continue
			}

			_ = item.Value(func(val []byte) error {
				var secret Secret
				if err := json.Unmarshal(val, &secret); err != nil || secret.Data == "" {
					return err
				}

				total++
				return nil
			})
		}
		return nil
	})
}
