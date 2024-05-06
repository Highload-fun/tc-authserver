package users

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

type User struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Country   string `json:"country"`
	IsAdmin   bool   `json:"is_admin,omitempty"`
	IsBlocked bool   `json:"-"`
}

type Users struct {
	mtx     sync.RWMutex
	byLogin map[string]*User
}

var (
	ErrExists         = errors.New("user exists")
	ErrNotFound       = errors.New("not found")
	ErrAlreadyBlocked = errors.New("already blocked")
)

func Load(filename string) *Users {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	res := &Users{
		byLogin: map[string]*User{},
	}

	jd := json.NewDecoder(f)
	for {
		var u User
		if err := jd.Decode(&u); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}

		res.byLogin[u.Login] = &u
		if u.Country == "" {
			log.Fatalf("%v", u)
		}
	}

	return res
}

func (u *Users) Get(login string) *User {
	u.mtx.RLock()
	defer u.mtx.RUnlock()

	return u.byLogin[login]
}

func (u *Users) Register(user User) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	if _, exists := u.byLogin[user.Login]; exists {
		return ErrExists
	}

	u.byLogin[user.Login] = &user

	return nil
}

func (u *Users) Edit(login string, password, name, phone, country *string) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	user, exists := u.byLogin[login]
	if !exists {
		return ErrNotFound
	}

	if password != nil {
		user.Password = *password
	}
	if name != nil {
		user.Name = *name
	}
	if phone != nil {
		user.Phone = *phone
	}
	if country != nil {
		user.Country = *country
	}

	return nil
}

func (u *Users) Block(login string) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	user := u.byLogin[login]
	if user == nil {
		return ErrNotFound
	}

	if user.IsBlocked {
		return ErrAlreadyBlocked
	}

	user.IsBlocked = true

	return nil
}

func (u *Users) Unblock(login string) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	user := u.byLogin[login]
	if user == nil || !user.IsBlocked {
		return ErrNotFound
	}

	user.IsBlocked = false

	return nil
}
