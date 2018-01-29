package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	for _, store := range []string{"S3", "Spaces", "Diskv", "Redis", ""} {
		NewStore(store)
		assert.NotNil(t, store)
	}
}