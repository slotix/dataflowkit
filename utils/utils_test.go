// Dataflow kit - utils
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

package utils

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMD5(t *testing.T) {
	b := GenerateMD5([]byte("testBytes"))
	assert.Equal(t, b, []byte{0xff, 0x7, 0x73, 0x84, 0xb2, 0xc0, 0x48, 0x57, 0xe0, 0x5c, 0x13, 0xa0, 0xee, 0x1f, 0x7f, 0xf3}, "Test MD5 hash generation")
	t.Logf("%x", b)
}

func TestGenerateCRC32(t *testing.T) {
	b := GenerateCRC32([]byte("testBytes"))
	assert.Equal(t, b, []byte{0x32, 0x38, 0x30, 0x34, 0x33, 0x36, 0x31, 0x65}, "Test CRC32 hash generation")
	t.Logf("%x", b)
}

func TestRelURL(t *testing.T) {
	r, err := RelUrl("http://books.toscrape.com", "catalogue/page-2.html")
	assert.NoError(t, err)
	t.Log(r)
	r, err = RelUrl("http://books.toscrape.com/catalogue/", "page-2.html")
	assert.NoError(t, err)
	t.Log(r)
	r, err = RelUrl("http://books.toscrape.com/catalogue/page-2.html", "in-her-wake_980/index.html")
	assert.NoError(t, err)
	t.Log(r)
}

func TestRandFloat(t *testing.T) {
	// This can be used to generate random floats in
	// other ranges, for example `0.5 <= f' < 1.5`.
	rand.Seed(time.Now().Unix())
	fmt.Println((rand.Float64() * 1) + 1)
	fmt.Println(rand.Float64() + 0.5)
	// This can be used to generate random floats in
	// other ranges, for example `5.0 <= f' < 10.0`.
	fmt.Println()
	fmt.Print((rand.Float64()*5)+5, ",")
	fmt.Print((rand.Float64() * 5) + 5)
	fmt.Println()
}

func TestRandF(t *testing.T) {
	s := 500 * time.Millisecond
	inttt := int64(RandomF() * 1000)
	//tt := time.Duration(inttt*s)
	fmt.Println(s, inttt)
}

func TestRandInt(t *testing.T) {
	//initial fetch delay
	s := 500 * time.Millisecond
	//random ratio
	rand := Random(500, 1500)
	m := s * time.Duration(rand) / 1000
	fmt.Println(s, rand, m)
}

// func TestGenerateCRC32(t *testing.T){
// 	fmt.Println(string(GenerateCRC32([]byte("test test test"))))
// }
