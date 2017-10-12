package service

import (
	"github.com/satori/go.uuid"
	"github.com/revel/modules/csrf/app"
	"time"
	"strconv"
	"os"
	"go/build"
	"github.com/MoonBabyLabs/kekcontact"
)

// The current space that occupies a document. Spaces can contain many different kek content items and class indexes.
// Many users can access and contribute to a single kekspace. A version control repository is a fair comparison.
type Kekspace struct {
	Contributors  []kekcontact.Contact
	Name    string
	Id      uuid.UUID
	CreatedAt time.Time
	Owner kekcontact.Contact
}

type KekspaceConfig struct {
	Owners []Contact `json:"contact"`
	Name string `json:"name"`
}

func (ks Kekspace) Load() (Kekspace, error) {
	_, err := Load(KEK_SPACE_CONFIG, ks)

	if err != nil {
		return ks, err
	}

	return ks, nil
}

func (ks Kekspace) New(config KekspaceConfig) (Kekspace, error) {
	defaultPath := build.Default.GOPATH
	os.Mkdir(build.Default.GOPATH + "/.kek", 0755)
	os.Mkdir(defaultPath + "/.kek/d", 0755)
	os.Mkdir(defaultPath + "/.kek/f", 0755)
	ks.Contributors = make([]kekcontact.Contact, len(config.Owners))

	if len(config.Name) == 0 {
		name, csrfError := csrf.RandomString(8)

		if csrfError != nil {
			return ks, csrfError
		}

		ks.CreatedAt = time.Now()
		ti := strconv.FormatInt(ks.CreatedAt.Unix(), 10)
		ks.Name = name + ti
	}

	saveErr := Save(KEK_SPACE_CONFIG, ks)

	if saveErr != nil {
		return ks, saveErr
	}

	return ks, nil
}