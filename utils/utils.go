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
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

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

// ArrayContains check if string slice contains string
func ArrayContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
