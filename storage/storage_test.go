package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	var store Store
	st, err := TypeString("unknownStorage")
	assert.Error(t, err)
	store = NewStore(st)
	for _, storeType := range []string{"S3", "Spaces", "Diskv", "Redis"} {
		st, err := TypeString(storeType)
		assert.NoError(t, err)
		store = NewStore(st)
		assert.NotNil(t, store)
	}

}
