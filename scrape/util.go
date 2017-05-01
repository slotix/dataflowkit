package scrape

import (
	"io"
	"strings"
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
