package utf8encoding

import (
	"bytes"
	"io"
	"io/ioutil"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

func ReaderToUtf8Encoding(rc io.ReadCloser) (r io.Reader, name string, certain bool, err error) {

	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return
	}
	e, name, certain := charset.DetermineEncoding(b, "")
	if err != nil {
		return
	}
	if name != "UTF-8" {
		r = transform.NewReader(bytes.NewReader(b), e.NewDecoder())
	}
	return
}
