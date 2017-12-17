package crypto

import (
	"bytes"
	"crypto/md5"
	"hash/crc32"
	"io"
	"strconv"
)

func GenerateMD5(b []byte) []byte {
	h := md5.New()
	r := bytes.NewReader(b)
	io.Copy(h, r)
	return h.Sum(nil)
}

func GenerateCRC32(b []byte) []byte {
	crc32InUint32 := crc32.ChecksumIEEE(b)
	crc32InString := strconv.FormatUint(uint64(crc32InUint32), 16)
	return []byte(crc32InString)
}
