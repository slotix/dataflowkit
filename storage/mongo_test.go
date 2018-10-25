package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mongo(t *testing.T) {
	m := newMongo("127.0.0.1")
	//Delete all values from redis if any
	err := m.DeleteAll()
	assert.NoError(t, err, "Expected no error")
	//write records
	testValue := []byte("testValue")

	recs := []Record{{
		Type:    CACHE,
		Key:     "testKey",
		Value:   testValue,
		ExpTime: 100,
	}, {
		Type:    COOKIES,
		Key:     "testKey",
		Value:   testValue,
		ExpTime: 100,
	}, {
		Type:    INTERMEDIATE,
		Key:     "PayloadHash-0-0",
		Value:   []byte(`{"selector1_text":"Selector1_value"}`),
		ExpTime: 100,
	},
	// {
	// 	Type: INTERMEDIATE,
	// 	Key:  "PayloadHash",
	// 	Value: []byte(`"6eafe89a":{"0":[0,1,2,3]}`),
	// 	ExpTime: 100,
	// }
	}

	rec := recs[0]
	for _, r := range recs {
		err = m.Write(r)
		assert.NoError(t, err, "Expected no error")
	}
	isExists := m.IsExists(rec)
	assert.Equal(t, true, isExists, "Is rec exists in db")
	value, _ := m.Read(Record{
		Type: rec.Type,
		Key:  rec.Key,
	})
	assert.Equal(t, testValue, value, "Expected equal")
	//read records
	for _, r := range recs {
		value, err := m.Read(r)
		assert.NoError(t, err, "Expected no error")
		assert.NotNil(t, value, "Expected NotNil")
	}

	//Write record with Empty key
	recEmptyKey := Record{
		Type:    CACHE,
		Key:     "",
		Value:   testValue,
		ExpTime: 0,
	}
	err = m.Write(recEmptyKey)
	assert.NoError(t, err, "Expected no error")

	//Invalid Intermediary value
	err = m.Write(Record{
		Type:    INTERMEDIATE,
		Key:     "PayloadHash-0-0",
		Value:   []byte(`InvalidJSON`),
		ExpTime: 100,
	})
	assert.Error(t, err, "Invalid Intermediary JSON value")

	//Write invalid value to Intermediary
	// q := fmt.Sprintf(writeIntermediateQuery, 100)
	// err = c.session.Query(q, "Payload", 100, 100, "InvalidValue").Exec()
	// assert.NoError(t, err)

	// value, err = c.Read(Record{
	// 	Type: INTERMEDIATE,
	// 	Key:  "Payload-100-100",
	// })
	// assert.Error(t, err, "Read Intermediate error")

	// Read NonExistent key
	_, err = m.Read(Record{
		Type: INTERMEDIATE,
		Key:  "NonExistentPayload-100-100",
	})
	assert.Error(t, err, "Expected error")

	expired := m.Expired(rec)
	assert.Equal(t, expired, false, "Expected non expired value")

	for _, r := range recs {
		err = m.Delete(r)
		assert.NoError(t, err, "Expected no error")
	}
	//delete nonexistant record
	err = m.Delete(Record{
		Type: INTERMEDIATE,
		Key:  "Payload-100-100",
	})
	assert.Error(t, err, "Not found error")

	//Add two values to storage
	m.Write(rec)
	m.Write(Record{
		Type:  COOKIES,
		Key:   "cookie1",
		Value: []byte("Cookie=Value"),
	})
	//erase all items
	err = m.DeleteAll()
	assert.NoError(t, err, "Expected no error")
	m.Close()
}
