package cache

import "errors"

// ErrorNoNewExort error in case of no new export
var ErrorNoNewExort = errors.New("hash match - no new contentserver export")

// ErrorFileNotExists in case the contentserver export file does not yet exist
var ErrorFileNotExists = errors.New("contentserver export cache file not exists")
