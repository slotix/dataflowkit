// Dataflow kit - crypto
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package crypto of the Dataflow kit contain functions for generating MD5 and CRC32 
//
package crypto

import (
	"bytes"
	"crypto/md5"
	"hash/crc32"
	"io"
	"strconv"
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
