package cache

import (
	"os"
)

// GetContentServerExport will return a contentserver export json
// hint: close file after reading
// hint: lock file read access
func (c *Cache) GetContentServerExport() (file *os.File, fileInfo os.FileInfo, err error) {

	if c.fileNotExists() {
		err = ErrorFileNotExists
		return
	}

	var errFileInfo error
	fileInfo, errFileInfo = os.Stat(c.file)
	if errFileInfo != nil {
		err = errFileInfo
		return
	}

	file, errOpenFile := os.Open(c.file)
	if errOpenFile != nil {
		err = errOpenFile
		return
	}
	// defer file.Close()

	return
}
