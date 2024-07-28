package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinCost     int = 4  // the minimum allowable cost as passed in to GenerateFromPassword
	MaxCost     int = 31 // the maximum allowable cost as passed in to GenerateFromPassword
	DefaultCost int = 10 // the cost that will actually be set if a cost below MinCost is passed into GenerateFromPassword
)

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RefreshTokens map[int]RefreshToken `json:"refresh_tokens"`
}

type RefreshToken struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"exp"`
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

func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {
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
	chirp.AuthorId = userId
	chirps = append(chirps, chirp)

	for idx := range chirps {
		if _, found := dat.Chirps[idx+1]; found {
			continue
		} else {
			dat.Chirps[idx+1] = chirp
		}
	}

	db.writeDB(dat)

	return chirp, nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	user := User{}
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := db.loadDB()
	if err != nil {
		return user, err
	}

	users := []User{}
	for _, user := range dat.Users {
		users = append(users, user)
	}

	user.Id = len(users) + 1
	user.Email = email

	hashedPwdBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return User{}, err
	}
	user.Password = string(hashedPwdBytes)
	user.IsChirpyRed = false
	users = append(users, user)

	for idx, user := range users {
		if usr, found := dat.Users[idx+1]; found {
			if usr.Email == email {
				return User{}, errors.New("User exists")
			}
		} else {
			dat.Users[idx+1] = user
		}
	}

	db.writeDB(dat)

	return user, nil
}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dat, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, userInDb := range dat.Users {
		if userInDb.Id == id {
			userInDb.Email = email
			hashedPwdBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
			if err != nil {
				return User{}, err
			}
			userInDb.Password = string(hashedPwdBytes)
			dat.Users[id] = userInDb
			db.writeDB(dat)
			return userInDb, nil
		}
	}
	return User{}, errors.New("User doesn't exist")
}

func (db *DB) LoginUser(email, password string) (bool, User, string, error) {
	user := User{}

	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := db.loadDB()
	if err != nil {
		return false, user, "", err
	}

	for _, userInDb := range dat.Users {
		if userInDb.Email == email {
			err = bcrypt.CompareHashAndPassword([]byte(userInDb.Password), []byte(password))
			if err != nil {
				return false, user, "", nil
			}

			refreshTokenBytes := make([]byte, 32)
			_, err := rand.Read(refreshTokenBytes)
			if err != nil {
				return false, user, "", nil
			}

			refreshToken := RefreshToken{
				Token:      hex.EncodeToString(refreshTokenBytes),
				Expiration: time.Now().Add(time.Hour * 24 * 60),
			}

			dat.RefreshTokens[userInDb.Id] = refreshToken

			db.writeDB(dat)

			return true, userInDb, hex.EncodeToString(refreshTokenBytes), nil
		}
	}

	return false, user, "", errors.New("User not found")
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
		Chirps:        map[int]Chirp{},
		Users:         map[int]User{},
		RefreshTokens: map[int]RefreshToken{},
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

func (db *DB) ValidateRefreshToken(token string) (bool, int, error) {
	data, err := db.loadDB()
	if err != nil {
		return false, 0, err
	}

	if len(data.RefreshTokens) == 0 {
		return false, 0, nil
	}

	for id, refreshToken := range data.RefreshTokens {
		if refreshToken.Token == token {
			if time.Now().Sub(refreshToken.Expiration) >= 0 {
				return false, 0, nil
			}
			return true, id, nil
		}
	}

	return false, 0, errors.New("Token not found")

}

func (db *DB) RevokeToken(token string) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := db.loadDB()
	if err != nil {
		return err
	}

	for id, refreshToken := range data.RefreshTokens {
		if refreshToken.Token == token {
			delete(data.RefreshTokens, id)
			db.writeDB(data)
			return nil
		}
	}

	return errors.New("Token not found")

}

func (db *DB) DeleteChirp(chirpId, userId int) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dat, err := db.loadDB()
	if err != nil {
		return err
	}

	for _, chirp := range dat.Chirps {
		if chirp.Id == chirpId {

			if chirp.AuthorId == userId {
				delete(dat.Chirps, chirpId)
				return nil
			}

			return errors.New("Forbidden")

		}
	}

	return errors.New("Chirp does not exist")
}

func (db *DB) UpgradeUser(userId int) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dat, err := db.loadDB()
	if err != nil {
		return err
	}

	for _, user := range dat.Users {
		if user.Id == userId {

			dat.Users[userId] = User{
        Id: user.Id,
        Email: user.Email,
        Password: user.Password,
        IsChirpyRed: true,
      }

			db.writeDB(dat)
			return nil
		}
	}

	return errors.New("No user found")
}
