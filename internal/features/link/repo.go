package link

import (
	"encoding/json"
	"errors"
	badger "github.com/dgraph-io/badger/v4"
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
		return txn.Set([]byte(link.Short), data)
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
