package img

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type ImageRepo struct {
	db *badger.DB
}

func NewImageRepo(db *badger.DB) *ImageRepo {
	return &ImageRepo{db: db}
}

func (r *ImageRepo) Create(image *Image) error {
	return r.db.Update(func(txn *badger.Txn) error {
		data, _ := json.Marshal(image)
		entry := badger.NewEntry([]byte(image.Short), data).WithTTL(time.Until(image.ExpiresAt))
		return txn.SetEntry(entry)
	})
}

func (r *ImageRepo) GetByShort(short string) (*Image, error) {
	var image Image

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(short))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &image)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}

	return &image, err
}
func (r *ImageRepo) Delete(short string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(short))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}
