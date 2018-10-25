package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {

	for _, sType := range []string{ /*"S3", "Spaces",*/ "Diskv", "Cassandra", "mongodb"} {
		store := NewStore(sType)
		assert.NotNil(t, store)
	}
}

func TestInvalidStore(t *testing.T) {
	sType := "unknownStorage"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	store := NewStore(sType)
	assert.NotNil(t, store)
}
