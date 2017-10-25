package kek

import (
	"testing"
	"strconv"
)

func TestStoreSaveLoadDelete(t *testing.T) {
	st := make(map[string]bool)
	sErr := Store{}.Save("blah", map[string]bool{"yes" : true})

	if sErr != nil {
		t.Log(sErr)
		t.Fail()
	}

	lErr := Store{}.Load("blah", &st)

	if !st["yes"] {
		t.Log("unable to load content into struct")
		t.Fail()
	}

	if lErr != nil {
		t.Log(lErr)
		t.Fail()
	}

	dErr := Store{}.Delete("blah")

	if dErr != nil {
		t.Log(dErr)
		t.Fail()
	}
}

func TestStore_List(t *testing.T) {
	for i:= 0; i < 40; i++ {
		Store{}.Save("tmp/" + strconv.Itoa(i), []byte{})
	}

	items, err := Store{}.List("tmp")

	if err != nil {
		t.Log()
		t.Fail()
	}

	for i:= 0; i < 40; i++ {
		if !items[strconv.Itoa(i)] {
			t.Log("missing list item: " + strconv.Itoa(i))
			t.Fail()
		}

		Store{}.Delete("tmp/" + strconv.Itoa(i))
	}
}