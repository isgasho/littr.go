package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

type SSHKey struct {
	Id      string `json:"id"`
	Private []byte `json:"-"`
	Public  []byte `json:"publicKeyPem"`
}

type AccountMetadata struct {
	Password []byte `json:"-"`
	Provider string `json:"provider,omitempty"`
	Salt     []byte `json:"-"`
	Key      SSHKey `json:"-"`
}

type account struct {
	Id        int64     `orm:id,"auto"`
	Key       Key       `orm:key`
	Email     []byte    `orm:email`
	Handle    string    `orm:handle`
	Score     int64     `orm:score`
	CreatedAt time.Time `orm:created_at`
	UpdatedAt time.Time `orm:updated_at`
	Flags     int8      `orm:flags`
	Metadata  []byte    `orm:metadata`
	Votes     map[Key]vote
}

type Account struct {
	Email     string           `json:"-"`
	Hash      string           `json:"hash"`
	Score     int64            `json:"score"`
	Handle    string           `json:"handle"`
	CreatedAt time.Time        `json:"-"`
	UpdatedAt time.Time        `json:"-"`
	Flags     int8             `json:"-"`
	Metadata  *AccountMetadata `json:"metadata,omitempty"`
	Votes     map[string]Vote  `json:"votes,omitempty"`
}

func (a Account) HasMetadata() bool {
	return a.Metadata != nil
}

func loadAccountFromModel(a account) Account {
	acct := Account{
		Hash:      a.Hash(),
		Flags:     a.Flags,
		UpdatedAt: a.UpdatedAt,
		Handle:    a.Handle,
		Score:     int64(float64(a.Score) / ScoreMultiplier),
		CreatedAt: a.CreatedAt,
		Email:     string(a.Email),
	}
	if a.Metadata != nil {
		err := json.Unmarshal(a.Metadata, &acct.Metadata)
		if err != nil {
			log.WithFields(log.Fields{}).Error(errors.NewErrWithCause(err, "unable to unmarshal account metadata"))
		}
	}

	return acct
}

func loadAccount(db *sql.DB, handle string) (Account, error) {
	a, err := loadAccountByHandle(db, handle)
	if err != nil {
		return Account{}, errors.Errorf("user %q not found", handle)
	}
	return loadAccountFromModel(a), nil
}

type Deletable interface {
	Deleted() bool
	Delete()
	UnDelete()
}

func (a *Account) VotedOn(i Item) *Vote {
	for key, v := range a.Votes {
		if key == i.Hash {
			return &v
		}
	}
	return nil
}

func (a account) Hash() string {
	return a.Hash8()
}
func (a account) Hash8() string {
	return string(a.Key[0:8])
}
func (a account) Hash16() string {
	return string(a.Key[0:16])
}
func (a account) Hash32() string {
	return string(a.Key[0:32])
}
func (a account) Hash64() string {
	return a.Key.String()
}

func (a Account) GetLink() string {
	return fmt.Sprintf("/~%s", a.Handle)
}

func GenKey(handle string) Key {
	data := []byte(handle)
	//now := a.UpdatedAt
	//if now.IsZero() {
	//	now = time.Now()
	//}
	k := Key{}
	k.FromString(fmt.Sprintf("%x", sha256.Sum256(data)))
	return k
}

func (a *Account) IsLogged() bool {
	return a != nil && (!a.CreatedAt.IsZero())
}

func loadAccountByHandle(db *sql.DB, handle string) (account, error) {
	a := account{}
	selAcct := `select "id", "key", "handle", "email", "score", "created_at", "updated_at", "metadata", "flags" from "accounts" where "handle" = $1`
	rows, err := db.Query(selAcct, handle)
	if err != nil {
		return a, err
	}
	var aKey []byte
	for rows.Next() {
		err = rows.Scan(&a.Id, &aKey, &a.Handle, &a.Email, &a.Score, &a.CreatedAt, &a.UpdatedAt, &a.Metadata, &a.Flags)
		if err != nil {
			return a, err
		}
		a.Key.FromBytes(aKey)
	}

	if err != nil {
		log.WithFields(log.Fields{}).Error(err)
	}

	return a, nil
}
