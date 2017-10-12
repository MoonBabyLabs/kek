package service

import (
	"time"
	"github.com/MoonBabyLabs/revchain"
	"github.com/satori/go.uuid"
	"encoding/json"
	"strconv"
	"errors"
	"sort"
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
	Id         string
	Attributes map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
	revchain.Chain
	Related    []KekDoc
	Revision string
}

func (kd KekDoc) Get(id string, withRevChain bool) (KekDoc, error) {
	_, kerr := Load(DOC_DIR + id, &kd)

	if kerr != nil {
		return kd, kerr
	}

	if withRevChain {
		rchain := revchain.Chain{}
		Load(DOC_DIR + id + ".rev", &rchain)
		kd.Chain = rchain
	}

	return kd, nil
}

// New will create a kekdoc, index the field attributes, map the classes & start the revision chain
func (kd KekDoc) New(attrs map[string]interface{}) (KekDoc, error) {
	ks := Kekspace{}
	_, ksLoadErr := Load(KEK_SPACE_CONFIG, ks)

	if ksLoadErr != nil {
		return kd, ksLoadErr
	}

	nowTime := time.Now()
	cuid := uuid.NewV5(ks.Id, nowTime.String())
	kd.Id = strconv.FormatInt(nowTime.Unix(), 10) + "." + cuid.String()
	kd.CreatedAt = nowTime
	kd.UpdatedAt = nowTime
	kd.Attributes = attrs
	js, _ := json.Marshal(attrs)
	blck := revchain.Block{}.New(ks, js, "", 0)
	kd.Revision = blck.HashString()
	Save(DOC_DIR + kd.Id, kd)
	kd.Chain = revchain.Chain{}.New(blck)
	Save(DOC_DIR + kd.Id + ".rev", kd.Chain)

	// Lets index our fields
	indexAttrs(kd.Id, attrs)

	return kd, nil
}

func (kd KekDoc) Update(id string, attrs map[string]interface{}, patch bool) (KekDoc, error) {
	Load(DOC_DIR + id, &kd)
	chain := revchain.Chain{}
	block := revchain.Block{}
	ks := Kekspace{}
	Load (KEK_SPACE_CONFIG, &ks)
	Load(DOC_DIR + id + ".rev", &chain)
	kd.removeAttrs()

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
	block.New(ks, data, lastBlock.HashString(), lastBlock.Index + 1)
	chain.AddBlock(block)
	Save(DOC_DIR + id + ".rev", chain)

	return kd, nil
}

// Delete a kekdoc and its associated indexed attributes & revision chain.
func (kd KekDoc) Delete(id string) error {
	_, err := Load(DOC_DIR + id, &kd)

	if err != nil {
		return err
	}

	kd.removeAttrs()
	Delete (DOC_DIR + id + ".rev")
	return Delete(DOC_DIR + id)
}

// removeAttrs will remove all of the indexed attributed for a Kekdoc
// @todo removeAttrs should remove deep indexes when functionality is added.
func (kd KekDoc) removeAttrs() {
	for attr, val := range kd.Attributes {
		strVal, isStr := val.(string)

		if isStr {
			Delete(FIELD_DIR + attr + "/" + strVal + "/" + kd.Id)
		}

		valInt, isInt := val.(int)

		if isInt {
			Delete(FIELD_DIR + attr + "/" + strconv.Itoa(valInt) + "/" + kd.Id)
		}

		strSlice, isStrSlice := val.([]string)

		if isStrSlice {
			for _, val := range strSlice {
				Delete(FIELD_DIR + attr + "/" + val + "/" + kd.Id)
			}

			continue
		}

		intSlice, isIntSlice := val.([]int)

		if  isIntSlice {
			for _, val := range intSlice {
				Delete(FIELD_DIR + attr + "/" + strconv.Itoa(val) + "/" + kd.Id)
			}

			continue
		}
	}
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
	finalDocs := make([]KekDoc, 0)
	docIds := make(map[string]int)
	limit := q.Limit

	if len(q.SearchQueries) == 0 {
		return finalDocs, errors.New("Need to have searchQueries in order to find docs")
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