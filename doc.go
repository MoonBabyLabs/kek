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
	"log"
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

type KekDoc struct {
	Id         string `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	revchain.Chain `json:"revisions"`
	Revision string `json:"rev"`
}

func (kd KekDoc) Get(id string, withRevChain bool) (KekDoc, error) {
	_, kerr := Load(DOC_DIR + id, &kd)

	if kerr != nil {
		return kd, kerr
	}

	if withRevChain {
		rchain := revchain.Chain{}
		Load(DOC_DIR + id + ".kek", &rchain)
		kd.Chain = rchain
	}

	return kd, nil
}

// New will create a kekdoc, index the field attributes, map the classes & start the revision chain
func (kd KekDoc) New(attrs map[string]interface{}) (KekDoc, error) {
	kd.Id = "dd" + xid.New().String()
	// Lets index our fields
	go func() {
		indexAttrs(kd.Id, attrs)
	}()

	ks := Kekspace{}
	_, ksLoadErr := Load(KEK_SPACE_CONFIG, &ks)
	log.Print(ks)

	if ksLoadErr != nil {
		return kd, ksLoadErr
	}

	nowTime := time.Now()
	kd.CreatedAt = nowTime
	kd.UpdatedAt = nowTime
	kd.Attributes = attrs
	blck := revchain.Block{}.New(ks, attrs, "", -1)
	kd.Revision = blck.HashString()
	Save(DOC_DIR + kd.Id, kd)
	kd.Chain = revchain.Chain{}.New(blck)
	Save(DOC_DIR + kd.Id + ".kek", kd.Chain)


	return kd, nil
}

func (kd KekDoc) Update(id string, attrs map[string]interface{}, patch bool) (KekDoc, error) {
	Load(DOC_DIR + id, &kd)
	kd.UpdatedAt = time.Now()
	chain := revchain.Chain{}
	block := revchain.Block{}
	ks := Kekspace{}
	Load (KEK_SPACE_CONFIG, &ks)
	Load(DOC_DIR + id + ".kek", &chain)

	go func() {
		kd.removeAttrs()
	}()

	if patch {
		for key, val := range attrs {
			kd.Attributes[key] = val
		}

		indexAttrs(id, kd.Attributes)
		kd.UpdatedAt = time.Now()
		Save(DOC_DIR + id, kd)
	} else {
		kd.Attributes = attrs
		indexAttrs(id, kd.Attributes)
		Save(DOC_DIR + id, kd)
	}

	data, _ := json.Marshal(attrs)
	lastBlock := chain.GetLast()
	block.New(ks, data, lastBlock.HashString(), lastBlock.Index)
	chain.AddBlock(block)
	Save(DOC_DIR + id + ".kek", chain)

	return kd, nil
}

// Delete a kekdoc and its associated indexed attributes & revision chain.
func (kd KekDoc) Delete(id string) error {
	_, err := Load(DOC_DIR + id, &kd)

	if err != nil {
		return err
	}

	delRev := make(chan error, 3)

	go func() {
		delRev <- kd.removeAttrs()
	}()
	go func() {
		delRev <- Delete(DOC_DIR + id + ".kek")
	}()
	go func() {
		delRev <- Delete(DOC_DIR + id)
	}()

	for i := 0; i < len(delRev); i++ {
		select {
		case del := <-delRev:
			if del != nil {
				return del
			}
		}
	}

	return nil
}

// removeAttrs will remove all of the indexed attributed for a Kekdoc
// @todo removeAttrs should remove deep indexes when functionality is added.
func (kd KekDoc) removeAttrs() error {
	deletedAttrs := make(chan error, len(kd.Attributes))

	for attr, _ := range kd.Attributes {
		val := kd.Attributes[attr]
		go func() {
			strVal, isStr := val.(string)

			if isStr {
				deletedAttrs <- Delete(FIELD_DIR + attr + "/" + strVal + "/" + kd.Id)
			}

			valInt, isInt := val.(int)

			if isInt {
				deletedAttrs <- Delete(FIELD_DIR + attr + "/" + strconv.Itoa(valInt) + "/" + kd.Id)
			}

			strSlice, isStrSlice := val.([]string)

			if isStrSlice {
				for _, val := range strSlice {
					go Delete(FIELD_DIR + attr + "/" + val + "/" + kd.Id)
				}
			}

			intSlice, isIntSlice := val.([]int)

			if  isIntSlice {
				for _, val := range intSlice {
					deletedAttrs <- Delete(FIELD_DIR + attr + "/" + strconv.Itoa(val) + "/" + kd.Id)
				}
			}
		}()
	}

	for range kd.Attributes {
		select {
		case deleted:= <-deletedAttrs:
			if deleted != nil {
				return deleted
			}
		}
	}

	return <- deletedAttrs
}

// indexAttrs will add all of the attribute indexes for a kekdoc.
// @todo do we want to add deep indexes for maps and arrays of maps? Seems may beyond the scope of needs.
func indexAttrs (id string, data map[string]interface{}) {
	emptyByte := make([]byte, 0)
	for field, value := range data {
		strVal, isStr := value.(string)

		if isStr {
			Save(FIELD_DIR + field + "/" + strVal + "/" + id, emptyByte)
			continue
		}

		inte, isInt := value.(int)

		if isInt {
			Save(FIELD_DIR + field + "/" + strconv.Itoa(inte) + "/" + id, emptyByte)
			continue
		}

		strSlice, isStrSlice := value.([]string)

		if isStrSlice {
			for _, val := range strSlice {
				Save(FIELD_DIR + field + "/" + val + "/" + id, emptyByte)
			}

			continue
		}

		intSlice, isIntSlice := value.([]int)

		if  isIntSlice {
			for _, val := range intSlice {
				Save(FIELD_DIR + field + "/" + strconv.Itoa(val) + "/" + id, emptyByte)
			}

			continue
		}
	}
}

// Find a kekdocument based on a DocQuery. The DocQuery must have a searchQuery as well.
func (kc KekDoc) Find(q DocQuery) ([]KekDoc, error) {
	docIds := make(map[string]int)
	limit := q.Limit

	if len(q.SearchQueries) == 0 {
		if limit < 1 {
			limit = 20
		}
		docFiles, _ := List(DOC_DIR, limit)
		for docId, _ := range docFiles {
			if !strings.Contains(docId, ".kek") {
				docIds[docId] = 1
			}
		}
	} else {
		docs := make(map[string]bool)

		for ind, queryInfo := range q.SearchQueries {
			switch queryInfo.Operator {
			case "=" :
				docs, _ = List(FIELD_DIR + queryInfo.Field + "/" + queryInfo.Value, 0)
				break
			}

			for kekId := range docs {
				if ind == 0 {
					docIds[kekId] = 1
				} else {
					if docIds[kekId] > 0 {
						docIds[kekId] = docIds[kekId] + 1
					}
				}
			}
		}
	}

	count := 0
	sortedDocs := make([]KekDoc, 0)
	queriesLength := len(q.SearchQueries)

	for docId, length := range docIds {
		kd := KekDoc{}
		if count == limit {
			break
		}

		if length < queriesLength {
			continue
		}

		Load(DOC_DIR + docId, &kd)
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