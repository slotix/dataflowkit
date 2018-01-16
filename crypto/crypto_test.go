// Dataflow kit - crypto
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

package crypto

import (
	"testing"

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
