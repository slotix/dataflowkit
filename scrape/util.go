package scrape

import (
	"io"
	"math/rand"
	"strings"
	"time"
)

func newStringReadCloser(s string) dummyReadCloser {
	return dummyReadCloser{strings.NewReader(s)}
}

type dummyReadCloser struct {
	r io.Reader
}

func (c dummyReadCloser) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (s dummyReadCloser) Close() error {
	return nil
}

var _ io.ReadCloser = &dummyReadCloser{}

func Random(min, max int64) int64 {
    rand.Seed(time.Now().Unix())
    return rand.Int63n(max - min) + min
	//return rand.Intn(max - min) + min
}

//RandomF generates random Float64 between 0.5 and 1.5
func  RandomF() float64{
	rand.Seed(time.Now().Unix())
	return rand.Float64()+0.5
}
