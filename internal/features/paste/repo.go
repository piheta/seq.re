package paste

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type PasteRepo struct {
	db *badger.DB
}

func NewPasteRepo(db *badger.DB) *PasteRepo {
	return &PasteRepo{db: db}
}

func (r *PasteRepo) Create(paste *Paste) error {
	return r.db.Update(func(txn *badger.Txn) error {
		data, _ := json.Marshal(paste)
		entry := badger.NewEntry([]byte(paste.Short), data).WithTTL(time.Until(paste.ExpiresAt))
		return txn.SetEntry(entry)
	})
}

func (r *PasteRepo) GetByShort(short string) (*Paste, error) {
	var paste Paste

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(short))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &paste)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}

	return &paste, err
}
func (r *PasteRepo) Delete(short string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(short))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}
