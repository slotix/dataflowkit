package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_cassandra(t *testing.T) {
	c := newCassandra("127.0.0.1")
	//Delete all values from redis if any
	err := c.DeleteAll()
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
		{
			Type: INTERMEDIATE,
			Key:  "PayloadHash",
			Value: []byte(`"6eafe89a"
			: {"0":[0,1,2,3]}`),
			ExpTime: 100,
		}}

	rec := recs[0]
	for _, r := range recs {
		err = c.Write(r)
		assert.NoError(t, err, "Expected no error")
	}
	value, err := c.Read(Record{
		Type: rec.Type,
		Key:  rec.Key,
	})
	assert.Equal(t, testValue, value, "Expected equal")
	//read records
	for _, r := range recs {
		value, err := c.Read(r)
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
	err = c.Write(recEmptyKey)
	assert.Error(t, err, "Expected empty key error")

	//Invalid Intermediary value
	err = c.Write(Record{
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
	value, err = c.Read(Record{
		Type: INTERMEDIATE,
		Key:  "NonExistentPayload-100-100",
	})
	assert.Error(t, err, "Expected error")

	expired := c.Expired(rec)
	assert.Equal(t, expired, false, "Expected non expired value")

	for _, r := range recs {
		err = c.Delete(r)
		assert.NoError(t, err, "Expected no error")
	}
	//delete nonexistant record
	err = c.Delete(Record{
		Type: INTERMEDIATE,
		Key:  "Payload-100-100",
	})
	assert.NoError(t, err, "Expected no error. When delete a non-existent value Cassandra will never complain")

	//Add two values to storage
	err = c.Write(rec)
	err = c.Write(Record{
		Type:  COOKIES,
		Key:   "cookie1",
		Value: []byte("Cookie=Value"),
	})
	//erase all items
	err = c.DeleteAll()
	assert.NoError(t, err, "Expected no error")
	c.Close()

}
