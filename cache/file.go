package cache

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func (c *Cache) fileNotExists() bool {
	if _, err := os.Stat(c.file); os.IsNotExist(err) {
		return true
	}
	return false
}

// hashFile calculates a md5 hash sum of a given file
func hashFile(filename string) (hash string, err error) {
	if _, errStat := os.Stat(filename); os.IsNotExist(errStat) {
		return "", nil
	}
	file, errOpenFile := os.Open(filename)
	if errOpenFile != nil {
		err = errOpenFile
		return
	}
	defer file.Close()
	hasher := md5.New()
	if _, errHash := io.Copy(hasher, file); errHash != nil {
		err = errHash
		return
	}
	hash = hex.EncodeToString(hasher.Sum(nil))
	return
}
