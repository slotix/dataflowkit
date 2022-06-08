package fetch

import (
	"bytes"
	"io"
	"io/ioutil"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

//readerToUtf8Encoding detects encoding of fetched document and convert it to utf8. It is used by Base Fetcher only.
func readerToUtf8Encoding(rc io.ReadCloser) (out io.ReadCloser, name string, certain bool, err error) {
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return
	}
	e, name, certain := charset.DetermineEncoding(b, "")
	if err != nil {
		return
	}
	if name != "utf-8" {
		out = ioutil.NopCloser(
			transform.NewReader(
				bytes.NewReader(b), e.NewDecoder()))
	} else {
		out = ioutil.NopCloser(bytes.NewReader(b))
	}
	return
}
