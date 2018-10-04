package storage

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type mongodb struct {
	session *mgo.Session
}

func newMongo(host string) *mongodb {
	s, err := mgo.Dial(host)
	if err != nil {
		log.Fatal(err)
	}
	return &mongodb{session: s}
}

//Reads value from mongodb by specified key
func (m mongodb) Read(rec Record) (value []byte, err error) {
	collection := m.session.DB("dfk").C(rec.Type)
	item := make(map[string]interface{})
	err = collection.Find(bson.M{"uid": rec.Key}).One(item)
	if err != nil {
		return
	}
	if rec.Type == INTERMEDIATE {
		delete(item, "_id")
		delete(item, "uid")
		return json.Marshal(item)
	}
	val, ok := item[rec.Type].(string)
	if !ok {
		return nil, fmt.Errorf("Failed to convert value to byte array")
	}
	return []byte(val), err
}

//Writes specified pair key value to storage.
//expTime value sets TTL for Redis storage.
//expTime set Metadata Expires value for S3Storage
func (m mongodb) Write(rec Record) error {
	value := map[string]interface{}{}
	switch rec.Type {
	case INTERMEDIATE:
		err := json.Unmarshal(rec.Value, &value)
		if err != nil {
			return err
		} /*
			case COOKIES:
				var cookie []interface{}
				err := json.Unmarshal(rec.Value, &cookie)
				if err != nil {
					return err
				}
				value[rec.Type] = cookie
			case CACHE:
				value[rec.Type] = string(rec.Value) */
	default:
		value[rec.Type] = string(rec.Value)
	}
	value["uid"] = rec.Key
	collection := m.session.DB("dfk").C(rec.Type)
	_, err := collection.Upsert(
		bson.M{"uid": rec.Key},
		value)
	if err != nil {
		return err
	}
	return nil
}

func (m mongodb) IsExists(rec Record) bool {
	collection := m.session.DB("dfk").C(rec.Type)
	ssss := make(map[string]interface{})
	err := collection.Find(bson.M{"uid": rec.Key}).One(ssss)
	return err == nil
}

//Is key expired ? It checks if parse results storage item is expired. Set up  Expiration as "ITEM_EXPIRE_IN" environment variable.
//html pages cache stores this info in sResponse.Expires . It is not used for fetch endpoint.
func (m mongodb) Expired(rec Record) bool {
	return false
}

//Delete deletes specified item from the store
func (m mongodb) Delete(rec Record) error {
	collection := m.session.DB("dfk").C(rec.Type)
	err := collection.Remove(bson.M{"uid": rec.Key})
	if err != nil {
		return err
	}
	return nil
}

//DeleteAll erases all items from the store
func (m mongodb) DeleteAll() error {
	return m.session.DB("dfk").DropDatabase()
}

// Close storage connection
func (m mongodb) Close() {
	m.session.Close()
}
