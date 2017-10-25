package kek

import (
	"time"
	"github.com/MoonBabyLabs/revchain"
	"github.com/satori/go.uuid"
	"encoding/json"
	"strconv"
	"sort"
	"strings"
	"github.com/rs/xid"
)

type SearchQuery struct {
	Operator string
	Field string
	Value string
}

type DocQuery struct {
	Id uuid.UUID
	Slug string
	WithDocs bool
	WithDocRevs bool
	Offset int
	SearchQueries []SearchQuery
	Limit int
	OrderBy string
}

type Doc struct {
	store	Storer
	Id         string `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	revchain.Chain `json:"revisions"`
	Revision string `json:"rev"`
}

func (kd Doc) Store() Storer {
	return kd.store
}

func (kd Doc) SetStore(store Storer) Doc {
	kd.store = store

	return kd
}

func (kd Doc) Get(id string, withRevChain bool) (Doc, error) {
	if kd.store == nil {
		kd.store = Store{}
	}

	kerr := kd.store.Load(DOC_DIR + id, &kd)

	if kerr != nil {
		return kd, kerr
	}

	if withRevChain {
		rchain := revchain.Chain{}
		kd.store.Load(DOC_DIR + id + ".kek", &rchain)
		kd.Chain = rchain
	}

	return kd, nil
}

// New will create a kekdoc, index the field attributes, map the classes & start the revision chain
func (kd Doc) New(attrs map[string]interface{}) (Doc, error) {
	if kd.store != nil {
		kd.store = Store{}
	}

	kd.Id = "dd" + xid.New().String()
	indDone := make(chan bool)
	// Lets index our fields
	go func() {
		indDone <- kd.indexAttrs(kd.Id, attrs)
	}()

	ks := Kekspace{}
	ksLoadErr := kd.store.Load(KEK_SPACE_CONFIG, &ks)

	if ksLoadErr != nil {
		return kd, ksLoadErr
	}

	nowTime := time.Now()
	kd.CreatedAt = nowTime
	kd.UpdatedAt = nowTime
	kd.Attributes = attrs
	blck := revchain.Block{}.New(ks, attrs, "", -1)
	kd.Revision = blck.HashString()
	kd.store.Save(DOC_DIR + kd.Id, kd)
	kd.Chain = revchain.Chain{}.New(blck)
	kd.store.Save(DOC_DIR + kd.Id + ".kek", kd.Chain)
	<- indDone

	return kd, nil
}

func (kd Doc) Update(id string, attrs map[string]interface{}, patch bool) (Doc, error) {
	if kd.store == nil {
		kd.store = Store{}
	}

	dIndDone := make(chan bool)
	kd.store.Load(DOC_DIR + id, &kd)
	kd.UpdatedAt = time.Now()
	chain := revchain.Chain{}
	block := revchain.Block{}
	ks := Kekspace{}
	kd.store.Load (KEK_SPACE_CONFIG, &ks)
	kd.store.Load(DOC_DIR + id + ".kek", &chain)

	go func() {
		success, _ := kd.removeAttrs()
		dIndDone <- success
	}()

	if patch {
		for key, val := range attrs {
			kd.Attributes[key] = val
		}

		kd.indexAttrs(id, kd.Attributes)
		kd.UpdatedAt = time.Now()
		kd.store.Save(DOC_DIR + id, kd)
	} else {
		kd.Attributes = attrs
		kd.indexAttrs(id, kd.Attributes)
		kd.store.Save(DOC_DIR + id, kd)
	}

	data, _ := json.Marshal(attrs)
	lastBlock := chain.GetLast()
	block.New(ks, data, lastBlock.HashString(), lastBlock.Index)
	chain.AddBlock(block)
	kd.store.Save(DOC_DIR + id + ".kek", chain)

	return kd, nil
}

// Delete a kekdoc and its associated indexed attributes & revision chain.
func (kd Doc) Delete(id string) error {
	if kd.store == nil {
		kd.store = Store{}
	}
	err := kd.store.Load(DOC_DIR + id, &kd)
	dAttrib := make(chan bool)

	if err != nil {
		return err
	}

	delRev := make(chan error, 3)

	go func() {
		success, _ := kd.removeAttrs()
		dAttrib <- success
	}()
	go func() {
		delRev <- kd.store.Delete(DOC_DIR + id + ".kek")
	}()
	go func() {
		delRev <- kd.store.Delete(DOC_DIR + id)
	}()

	for i := 0; i < len(delRev); i++ {
		select {
		case del := <-delRev:
			if del != nil {
				return del
			}
		}
	}

	<-dAttrib

	return nil
}

// removeAttrs will remove all of the indexed attributed for a Kekdoc
// @todo removeAttrs should remove deep indexes when functionality is added.
func (kd Doc) removeAttrs() (bool, error) {
	deletedAttrs := make(chan error, len(kd.Attributes))

	for attr := range kd.Attributes {
		val := kd.Attributes[attr]
		go func() {
			strVal, isStr := val.(string)

			if isStr {
				deletedAttrs <- kd.store.Delete(FIELD_DIR + attr + "/" + strVal + "/" + kd.Id)
			}

			valInt, isInt := val.(int)

			if isInt {
				deletedAttrs <- kd.store.Delete(FIELD_DIR + attr + "/" + strconv.Itoa(valInt) + "/" + kd.Id)
			}

			strSlice, isStrSlice := val.([]string)

			if isStrSlice {
				for _, val := range strSlice {
					go kd.store.Delete(FIELD_DIR + attr + "/" + val + "/" + kd.Id)
				}
			}

			intSlice, isIntSlice := val.([]int)

			if  isIntSlice {
				for _, val := range intSlice {
					deletedAttrs <- kd.store.Delete(FIELD_DIR + attr + "/" + strconv.Itoa(val) + "/" + kd.Id)
				}
			}
		}()
	}

	for range kd.Attributes {
		select {
		case deleted:= <-deletedAttrs:
			if deleted != nil {
				return false, deleted
			}
		}
	}

	return true, <- deletedAttrs
}

// indexAttrs will add all of the attribute indexes for a kekdoc.
// @todo do we want to add deep indexes for maps and arrays of maps? Seems may beyond the scope of needs.
func (kd Doc) indexAttrs (id string, data map[string]interface{}) bool {
	emptyByte := make([]byte, 0)
	for field, value := range data {
		strVal, isStr := value.(string)

		if isStr {
			kd.store.Save(FIELD_DIR + field + "/" + strVal + "/" + id, emptyByte)
			continue
		}

		inte, isInt := value.(int)

		if isInt {
			kd.store.Save(FIELD_DIR + field + "/" + strconv.Itoa(inte) + "/" + id, emptyByte)
			continue
		}

		strSlice, isStrSlice := value.([]string)

		if isStrSlice {
			for _, val := range strSlice {
				kd.store.Save(FIELD_DIR + field + "/" + val + "/" + id, emptyByte)
			}

			continue
		}

		intSlice, isIntSlice := value.([]int)

		if  isIntSlice {
			for _, val := range intSlice {
				kd.store.Save(FIELD_DIR + field + "/" + strconv.Itoa(val) + "/" + id, emptyByte)
			}

			continue
		}
	}

	return true
}

// Find a kekdocument based on a DocQuery. The DocQuery must have a searchQuery as well.
func (kd Doc) Find(q DocQuery) ([]Doc, error) {
	if kd.store != nil {
		kd.store = Store{}
	}
	
	docIds := make(map[string]int)
	limit := q.Limit

	if len(q.SearchQueries) == 0 {
		if limit < 1 {
			limit = 20
		}
		docFiles, _ := kd.store.List(DOC_DIR)
		for docId := range docFiles {
			if !strings.Contains(docId, ".kek") {
				docIds[docId] = 1
			}
		}
	} else {
		docs := make(map[string]bool)

		for ind, queryInfo := range q.SearchQueries {
			switch queryInfo.Operator {
			case "=" :
				docs, _ = kd.store.List(FIELD_DIR + queryInfo.Field + "/" + queryInfo.Value)
				break
			}

			for kId := range docs {
				if ind == 0 {
					docIds[kId] = 1
				} else {
					if docIds[kId] > 0 {
						docIds[kId] = docIds[kId] + 1
					}
				}
			}
		}
	}

	count := 0
	sortedDocs := make([]Doc, 0)
	queriesLength := len(q.SearchQueries)

	for docId, length := range docIds {
		kd := Doc{}
		if count == limit {
			break
		}

		if length < queriesLength {
			continue
		}

		kd.store.Load(DOC_DIR + docId, &kd)
		sortedDocs = append(sortedDocs, kd)
		count++
	}

	if len(q.OrderBy) == 0 {
		sort.Slice(sortedDocs, func(i, j int) bool {
			return sortedDocs[i].CreatedAt.Unix() < sortedDocs[j].CreatedAt.Unix()
		})

	} else {
		sort.Slice(sortedDocs, func(i, j int) bool {
			return sortedDocs[i].Attributes[q.OrderBy].(string) < sortedDocs[j].Attributes[q.OrderBy].(string)
		})
	}

	if q.Offset > 0 && len(sortedDocs) > q.Offset {
		sortedDocs = sortedDocs[q.Offset:]
	}

	if len(sortedDocs) > 0 && len(sortedDocs) > limit {
		sortedDocs = sortedDocs[:limit]
	}

	return sortedDocs, nil
}