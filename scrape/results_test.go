package scrape

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultsFirst(t *testing.T) {
	r := &Results{
		Results: [][]map[string]interface{}{
			{{"foo": 1, "bar": 2}},
		},
	}

	assert.Equal(t, r.First(), map[string]interface{}{
		"foo": 1,
		"bar": 2,
	})

	r = &Results{
		Results: [][]map[string]interface{}{{}},
	}
	assert.Nil(t, r.First())
}

func TestResultsAllBlocks(t *testing.T) {
	r := &Results{
		Results: [][]map[string]interface{}{
			{{"foo": 1, "bar": 2}},
			{{"baz": 3, "asdf": 4}},
		},
	}

	assert.Equal(t, r.AllBlocks(), []map[string]interface{}{
		{"foo": 1, "bar": 2},
		{"baz": 3, "asdf": 4},
	})
}
