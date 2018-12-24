package mongostore

import (
	"github.com/foomo/shop/persistence"
	"github.com/pkg/errors"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var persistors = map[string]*persistence.Persistor{}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func getPersistor(url string, collection string) (*persistence.Persistor, error) {

	key := url + "." + collection

	if _, ok := persistors[key]; !ok {
		p, err := persistence.NewPersistor(url, collection)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, errors.New("failed to create mongoDB order persistor: " + err.Error())

		}
		persistors[key] = p
	}

	return persistors[key], nil
}
