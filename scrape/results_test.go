package scrape

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultsFirst(t *testing.T) {
	r := &Results{
		Output: [][]map[string]interface{}{
			{{"foo": 1, "bar": 2}},
		},
	}

	assert.Equal(t, r.First(), map[string]interface{}{
		"foo": 1,
		"bar": 2,
	})

	r = &Results{
		Output: [][]map[string]interface{}{{}},
	}
	assert.Nil(t, r.First())
}

func TestResultsAllBlocks(t *testing.T) {
	r := &Results{
		Output: [][]map[string]interface{}{
			{{"foo": 1, "bar": 2}},
			{{"baz": 3, "asdf": 4}},
		},
	}

	assert.Equal(t, r.AllBlocks(), []map[string]interface{}{
		{"foo": 1, "bar": 2},
		{"baz": 3, "asdf": 4},
	})
}
