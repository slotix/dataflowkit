package scrape

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScraper_partNames(t *testing.T) {
	s := Scraper{}
	s.Parts = []Part{
		Part{Name: "1"},
		Part{Name: "2"},
		Part{Name: "3"},
		Part{Name: "4"},
	}
	parts := s.partNames()
	assert.Equal(t, []string{"1", "2", "3", "4"}, parts)

}

func TestPayload_selectors(t *testing.T) {
	p1 := Payload{
		Fields: []Field{
			Field{Selector: "sel1"},
			Field{Selector: "sel2"},
			Field{Selector: "sel3"},
			Field{Selector: "sel4"},
		},
	}
	p2 := Payload{
		Fields: []Field{
			Field{},
			Field{},
			Field{},
			Field{},
		},
	}

	s1, err := p1.selectors()
	assert.NoError(t, err)
	assert.Equal(t, []string{"sel1","sel2","sel3","sel4"}, s1)
	s2, err := p2.selectors()
	assert.Error(t, err)
	assert.Equal(t, []string(nil), s2)

}

func TestNewTask(t *testing.T) {
	task := NewTask(Payload{})
	assert.NotEmpty(t, task.ID)
	start, err := task.startTime()
	assert.NoError(t, err)
	assert.NotNil(t, start, "task start time is not nil")
	//t.Log(start)
}
