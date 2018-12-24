package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// downloadNeosContentServerExport ...
func downloadNeosContentServerExport(filename string, neosURL string) error {
	file, errFile := os.Create(filename)
	if errFile != nil {
		return errFile
	}
	defer file.Close()

	response, errGetExport := http.Get(neosURL)
	if errGetExport != nil {
		return errGetExport
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintln("unexpected status code from site contentserver export", response.StatusCode, response.Status))
	}
	defer response.Body.Close()

	// obj for tmp. json decoding
	var obj interface{}

	// decode JSON stream
	decoder := json.NewDecoder(response.Body)
	errDecoding := decoder.Decode(&obj)
	if errDecoding != nil {
		return errDecoding
	}

	// encode JSON stream
	encoder := json.NewEncoder(file)
	errEncoding := encoder.Encode(obj)
	if errEncoding != nil {
		return errEncoding
	}

	return nil
}
