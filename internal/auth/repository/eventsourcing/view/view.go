package view

import (
	"github.com/jinzhu/gorm"

	"github.com/EonsofStupid/tessera/internal/crypto"
	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/query"
)

type View struct {
	Db           *gorm.DB
	client       *database.DB
	keyAlgorithm crypto.AuthAlgorithm
	query        *query.Queries
	es           *eventstore.Eventstore
}

func StartView(sqlClient *database.DB, keyAlgorithm crypto.AuthAlgorithm, queries *query.Queries, es *eventstore.Eventstore) (*View, error) {
	gorm, err := gorm.Open("postgres", sqlClient.DB)
	if err != nil {
		return nil, err
	}
	return &View{
		Db:           gorm,
		client:       sqlClient,
		keyAlgorithm: keyAlgorithm,
		query:        queries,
		es:           es,
	}, nil
}

func (v *View) Health() (err error) {
	return v.Db.DB().Ping()
}
