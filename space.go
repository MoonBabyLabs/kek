package kek

import (
	"github.com/satori/go.uuid"
	"time"
	"github.com/MoonBabyLabs/kekcontact"
	"github.com/rs/xid"
)

// The current space that occupies a document. Spaces can contain many different kekdocs.
// Many users can access and contribute to a single kekspace. A version control repository is a fair comparison.
type Kekspace struct {
	Contributors  []kekcontact.Contact `json:"contributors"`
	Name    string `json:"name"`
	Id      uuid.UUID `json:"id"`
	CreatedAt time.Time
	Owner kekcontact.Contact `json:"owner"`
	KekId string `json:"kek_id"`
}

func (ks Kekspace) Load() (Kekspace, error) {
	_, err := Load(KEK_SPACE_CONFIG, &ks)

	if err != nil {
		return ks, err
	}

	return ks, nil
}

func (ks Kekspace) New() (Kekspace, error) {
	ks.CreatedAt = time.Now()
	ks.Id = uuid.NewV4()
	ks.KekId = "ss" + xid.New().String()
	saveErr := Save(KEK_SPACE_CONFIG, ks)

	if saveErr != nil {
		return ks, saveErr
	}

	return ks, nil
}