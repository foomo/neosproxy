package mongostore

import (
	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/logging"
	"github.com/pkg/errors"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var globalmongostore store.Store

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// GetGlobalMongoStore will return an initialized (mongo-)store or fail
func GetGlobalMongoStore() store.Store {
	if globalmongostore == nil {
		logging.GetDefaultLogEntry().Fatal("global store has not been initialized")
	}
	return globalmongostore
}

// MustInitGlobalMongoStore will initialize a (mongo-)store or fail
func MustInitGlobalMongoStore(url string) {
	err := InitGlobalMongoStore(url)
	if err != nil {
		logging.GetDefaultLogEntry().WithError(err).Fatal("could not initialize global mongostore")
	}
}

// InitGlobalMongoStore will initialize a (mongo-)store or return an error
func InitGlobalMongoStore(url string) error {

	ms, msErr := NewMongoStore(url)
	if msErr != nil {
		return msErr
	}

	if ms == nil {
		return errors.New("could not create mongostore")
	}

	globalmongostore = ms
	return nil
}
