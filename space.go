package kek

import (
	"github.com/satori/go.uuid"
	"time"
	"github.com/MoonBabyLabs/kekcontact"
	"github.com/rs/xid"
	"errors"
)



// The current space that occupies a document. Spaces can contain many different kekdocs.
// Many users can access and contribute to a single kekspace. A version control repository is a fair comparison.
type Kekspace struct {
	Store Storer
	Contributors  []kekcontact.Contact `json:"contributors"`
	Name    string `json:"name"`
	Id      uuid.UUID `json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Owner kekcontact.Contact `json:"owner"`
	Description string `json:"description"`
	KekId string `json:"kek_id"`
}


func (ks Kekspace) Load() (Kekspace, error) {
	var err error

	if ks.Store == nil {
		err = Store{}.Load(KEK_SPACE_CONFIG, &ks)
	} else {
		err = ks.Store.Load(KEK_SPACE_CONFIG, &ks)
	}

	if err != nil {
		return ks, err
	}

	return ks, nil
}

func (ks Kekspace) New(name string, description string, owner kekcontact.Contact, contributors []kekcontact.Contact) (Kekspace, error) {
	loaded, _ := ks.Load()

	if loaded.KekId != "" {
		err := errors.New("Kekspace already exists for this user")
		return ks, err
	}

	if ks.Store == nil {
		ks.Store = Store{}
	}

	ks.Name = name
	ks.Description = description
	ks.Owner = owner
	ks.Contributors = contributors
	ks.CreatedAt = time.Now()
	ks.UpdatedAt = time.Now()
	ks.Id = uuid.NewV4()
	ks.KekId = "ss" + xid.New().String()
	saveErr := ks.Save()

	if saveErr != nil {
		return ks, saveErr
	}

	return ks, nil
}

func (ks Kekspace) Delete(kekid string) error {
	return ks.Store.Delete(KEK_SPACE_CONFIG)
}

func (ks Kekspace) Save() error {
	ks.UpdatedAt = time.Now()
	return ks.Store.Save(KEK_SPACE_CONFIG, ks)
}