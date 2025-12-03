package link

import (
	"encoding/json"
	"errors"
	badger "github.com/dgraph-io/badger/v4"
	"time"
)

type LinkRepo struct {
	db *badger.DB
}

func NewLinkRepo(db *badger.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) Create(link *Link) error {
	return r.db.Update(func(txn *badger.Txn) error {
		data, _ := json.Marshal(link)
		entry := badger.NewEntry([]byte(link.Short), data).WithTTL(time.Until(link.ExpiresAt))
		return txn.SetEntry(entry)
	})
}

func (r *LinkRepo) GetByShort(short string) (*Link, error) {
	var link Link

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(short))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &link)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}

	return &link, err
}

func (r *LinkRepo) Delete(short string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(short))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}
