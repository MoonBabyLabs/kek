package kek

import (
	"testing"
	"github.com/MoonBabyLabs/kekcontact"
)

type KekDataStoreTester struct {
}

func (s KekDataStoreTester) Load(path string, unmarshalStruct interface{}) error {
	return nil
}

func (s KekDataStoreTester) Save(path string, cont interface{}) error {
	return nil
}

func (s KekDataStoreTester) Delete(id string) error {
	return nil
}

func (s KekDataStoreTester) List(locale string) (map[string]bool, error) {
	return map[string]bool{}, nil
}

func TestNewSpace(t *testing.T) {
	ks := Kekspace{}
	ks.Store = KekDataStoreTester{}
	owner := kekcontact.Contact{
		Name: "a",
		Email: "b",
		Phone: "3",
		Id: "aasdf",
		Address: "333",
	}
	kek, err := ks.New("a", "b", owner, []kekcontact.Contact{owner, owner, owner})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if kek.KekId == "" {
		t.Error("KekId should not return an empty string when new is called")
		t.Fail()
	}

	if kek.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
		t.Fail()
	}

	if kek.Name != "a" {
		t.Error("Name should equal 'a'")
		t.Fail()
	}

	if kek.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be populated and not zero when new is called")
		t.Fail()
	}

	if len(kek.Contributors) < 1 {
		t.Error("A new Kekspace should have at least 1 contributor")
		t.Fail()
	}

	if kek.Owner.Name != "a" {
		t.Error("Woops, an owner must be populated at least with a name")
		t.Fail()
	}
}

func TestKekspace_Load(t *testing.T) {
	ks := Kekspace{}
	ks.Store = KekDataStoreTester{}
	_, kspaceErr := ks.Load()

	if kspaceErr != nil {
		t.Error("Loaded space should not return an error for tests")
		t.Fail()
	}
}

func TestKekspace_Save(t *testing.T) {
	cont := kekcontact.Contact{}
	many := []kekcontact.Contact{cont}
	ks := Kekspace{}
	ks.Store = KekDataStoreTester{}
	ks.New("a", "b", cont, many)
	err := ks.Delete(ks.KekId)

	loaded, loadErr := ks.Load()
	t.Log(loaded)
	t.Log(err)
	t.Log(loadErr)

	if loaded.KekId != "" {
		t.Log("Kekspace should be empty after it is deleted")
		t.Fail()
	}
}

func BenchmarkKekspace_Load(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Kekspace{}.Load()
	}
}

func BenchmarkKekspace_New(b *testing.B) {
	comp := kekcontact.Company{
		Name: "sample",
		Email: "a@aaa.com",
		Phone: "333",
		Id: "333 asdf ",
		Address: "Sample address",
		City: "Sample city",
		Region: "Sample region",
		PostalCode: "3a3 lew",
		CountryCode: "US",
	}
	cont := kekcontact.Contact{
		Name: "sample",
		Email: "a@aaa.com",
		Phone: "333",
		Id: "333 asdf ",
		Address: "Sample address",
		City: "Sample city",
		Region: "Sample region",
		PostalCode: "3a3 lew",
		CountryCode: "US",
		Company: comp,
	}

	for i := 0; i < b.N; i++ {
		Kekspace{Store: KekDataStoreTester{}}.New("my name", "my description", cont, []kekcontact.Contact{cont, cont, cont})
	}
}
