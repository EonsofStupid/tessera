package view

import (
	"github.com/jinzhu/gorm"

	"github.com/EonsofStupid/tessera/internal/database"
	"github.com/EonsofStupid/tessera/internal/query"
)

type View struct {
	Db     *gorm.DB
	client *database.DB
	Query  *query.Queries
}

func StartView(sqlClient *database.DB, queries *query.Queries) (*View, error) {
	gorm, err := gorm.Open("postgres", sqlClient.DB)
	if err != nil {
		return nil, err
	}
	return &View{
		Db:     gorm,
		Query:  queries,
		client: sqlClient,
	}, nil
}

func (v *View) Health() (err error) {
	return v.Db.DB().Ping()
}
