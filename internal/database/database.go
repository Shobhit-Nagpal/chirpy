package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	err := db.ensureDB()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	chirp := Chirp{}
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := db.loadDB()
	if err != nil {
		return chirp, err
	}

	chirps := []Chirp{}
	for _, chirp := range dat.Chirps {
		chirps = append(chirps, chirp)
	}

	chirp.Id = len(chirps) + 1
	chirp.Body = body
	chirps = append(chirps, chirp)

	for idx := range chirps {
		if _, found := dat.Chirps[idx + 1]; found {
			continue
		} else {
			dat.Chirps[idx + 1] = chirp
		}
	}

	db.writeDB(dat)

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {

	chirps := []Chirp{}

	data, err := db.loadDB()
	if err != nil {
		return chirps, nil
	}

	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		_, err := os.Create(db.path)
		if err != nil {
			return err
		}

    err = os.WriteFile(db.path, []byte("{}"), 066)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}

func (db *DB) GetChirpById(id int) (Chirp, error) {
  dat, err := db.loadDB()
	if err != nil {
		return Chirp{}, nil
	}

  if chirp, found := dat.Chirps[id]; found {
    return chirp, nil
  } else {
    return Chirp{}, errors.New("Chirp not found")
  }
}

func (db *DB) loadDB() (DBStructure, error) {
	dbStructure := DBStructure{
    Chirps: map[int]Chirp{},
  }
	dat, err := os.ReadFile(db.path)
	if err != nil {
		return dbStructure, err
	}

	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, []byte(dat), 066)
	if err != nil {
		return err
	}

	return nil
}
