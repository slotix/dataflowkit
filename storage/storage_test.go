package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	var store Store
	sType := "unknownStorage"
	store = NewStore(sType)
	for _, sType := range []string{ /*"S3", "Spaces",*/ "Diskv", "Cassandra"} {
		store = NewStore(sType)
		assert.NotNil(t, store)
	}
}
