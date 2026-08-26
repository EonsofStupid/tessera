package initialise

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

func Test_verifyDB(t *testing.T) {
	err := ReadStmts()
	if err != nil {
		t.Errorf("unable to read stmts: %v", err)
		t.FailNow()
	}

	type args struct {
		db       db
		database string
	}
	tests := []struct {
		name      string
		args      args
		targetErr error
	}{
		{
			name: "doesn't exist, create fails",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_database()", nil, []string{"current_database"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{false},
					}, "nomen"),
					expectExec("-- replace nomen with the name of the database\nCREATE DATABASE \"nomen\"", sql.ErrTxDone),
				),
				database: "nomen",
			},
			targetErr: sql.ErrTxDone,
		},
		{
			name: "doesn't exist, create successful",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_database()", nil, []string{"current_database"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{false},
					}, "nomen"),
					expectExec("-- replace nomen with the name of the database\nCREATE DATABASE \"nomen\"", nil),
				),
				database: "nomen",
			},
			targetErr: nil,
		},
		{
			name: "already exists in catalog, skip creation",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_database()", nil, []string{"current_database"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{true},
					}, "nomen"),
				),
				database: "nomen",
			},
			targetErr: nil,
		},
		{
			name: "catalog check fails",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_database()", nil, []string{"current_database"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", sql.ErrConnDone, []string{"exists"}, [][]driver.Value{}, "nomen"),
				),
				database: "nomen",
			},
			targetErr: sql.ErrConnDone,
		},
		{
			name: "same database as admin connection, skip creation",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_database()", nil, []string{"current_database"}, [][]driver.Value{
						{"nomen"},
					}),
				),
				database: "nomen",
			},
			targetErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyDatabase(tt.args.database)(t.Context(), tt.args.db.db); !errors.Is(err, tt.targetErr) {
				t.Errorf("verifyDB() error = %v, want: %v", err, tt.targetErr)
			}
			if err := tt.args.db.mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
