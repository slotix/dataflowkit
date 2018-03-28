// Dataflow kit - utils
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package utils of the Dataflow kit includes various functions and helpers to be used by other packages.
//
package utils

import (
	"bytes"
	"crypto/md5"
	"hash/crc32"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slotix/dataflowkit/logger"
)

var logger *logrus.Logger

func init() {
	logger = log.NewLogger()
}

// GenerateMD5 returns MD5 hash of provided byte array.
func GenerateMD5(b []byte) []byte {
	h := md5.New()
	r := bytes.NewReader(b)
	io.Copy(h, r)
	return h.Sum(nil)
}

// GenerateCRC32 returns CRC32 hash of provided byte array.
func GenerateCRC32(b []byte) []byte {
	crc32InUint32 := crc32.ChecksumIEEE(b)
	crc32InString := strconv.FormatUint(uint64(crc32InUint32), 16)
	return []byte(crc32InString)
}

// RelUrl is a helper function that aids in calculating the absolute URL from a
// base URL and relative URL.
func RelUrl(base, rel string) (string, error) {
	baseUrl, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	relUrl, err := url.Parse(rel)
	if err != nil {
		return "", err
	}

	newUrl := baseUrl.ResolveReference(relUrl)
	return newUrl.String(), nil
}

//stringInSlice check if specified string in the slice of strings
func stringInSlice(a string, list []string) bool {
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

// AddStringSliceToSlice joins two string slices.
func AddStringSliceToSlice(in []string, out []string) {
	for _, s := range in {
		if !stringInSlice(s, out) {
			out = append(out, s)
		}
	}
}

// ReadLinesOfFile returns the lines from a file as a slice of strings
func ReadLinesOfFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error(err.Error())
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

func newStringReadCloser(s string) dummyReadCloser {
	return dummyReadCloser{strings.NewReader(s)}
}

type dummyReadCloser struct {
	r io.Reader
}

func (c dummyReadCloser) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (dummyReadCloser) Close() error {
	return nil
}

var _ io.ReadCloser = &dummyReadCloser{}

//Random generates random int64 value
func Random(min, max int64) int64 {
	rand.Seed(time.Now().Unix())
	return rand.Int63n(max-min) + min
	//return rand.Intn(max - min) + min
}

//RandomF generates random Float64 between 0.5 and 1.5
func RandomF() float64 {
	rand.Seed(time.Now().Unix())
	return rand.Float64() + 0.5
}
