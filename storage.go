package homekit

import (
	"encoding/json"
	"fmt"

	"github.com/brutella/hc/db"
	"github.com/peterbourgon/diskv"
)

type Storage struct {
	driver *diskv.Diskv
}

func NewDatabase(dir string) (*Storage, error) {
	db := diskv.New(diskv.Options{
		BasePath:     dir,
		Transform:    func(s string) []string { return []string{} },
		CacheSizeMax: 1024 * 1024,
	})

	return &Storage{driver: db}, nil
}

func (store *Storage) EntityWithName(name string) (db.Entity, error) {
	data, err := store.driver.Read(name)
	if err != nil {
		return db.Entity{}, err
	}

	var ent db.Entity
	return ent, json.Unmarshal(data, &ent)
}

func (store *Storage) SaveEntity(entity db.Entity) error {
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	return store.driver.Write(entity.Name, data)
}

func (store *Storage) DeleteEntity(entity db.Entity) {
	_ = store.driver.Erase(entity.Name)
}

func (store *Storage) Entities() ([]db.Entity, error) {
	var entities = make([]db.Entity, 0, 2)
	cancel := make(chan struct{})
	defer close(cancel)

	keys := store.driver.Keys(cancel)
	for key := range keys {
		ent, err := store.EntityWithName(key)
		if err != nil {
			return nil, err
		}

		entities = append(entities, ent)
	}

	return entities, nil
}

func (store *Storage) Set(key, data []byte) error {
	return store.driver.Write(string(key), data)
}

func (store *Storage) Get(key []byte) ([]byte, error) {
	return store.driver.Read(string(key))
}
