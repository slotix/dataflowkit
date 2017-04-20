package helpers

import (
	"bytes"
	"crypto/md5"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "helpers: ", log.Lshortfile)
}

//stringInSlice check if specified string in the slice of strings
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// InsertStringToSlice inserts the value into the slice at the specified index,
// which must be in range.
// The slice must have room for the new element.
func InsertStringToSlice(slice []string, index int, value string) []string {
	// Grow the slice by one element.
	slice = slice[0 : len(slice)+1]
	// Use copy to move the upper part of the slice out of the way and open a hole.
	copy(slice[index+1:], slice[index:])
	// Store the new value.
	slice[index] = value
	// Return the result.
	return slice
}

func AddStringSliceToSlice(in []string, out []string) {
	for _, s := range in {
		if !StringInSlice(s, out) {
			out = append(out, s)
		}
	}
}

//func generateMD5(s string) string {
func GenerateMD5(b []byte) []byte {
	h := md5.New()
	r := bytes.NewReader(b)
	io.Copy(h, r)
	return h.Sum(nil)
}

// ReadLinesOfFile returns the lines from a file as a slice of strings
func ReadLinesOfFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Println(err.Error())
	}
	lines := strings.Split(string(content), "\n")
	return lines
}

// RegSplit splits the strings into strings using the regular expression as separator
func RegSplit(text string, reg *regexp.Regexp) []string {
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}
