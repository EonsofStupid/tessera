package initialise

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

func Test_verifyUser(t *testing.T) {
	err := ReadStmts()
	if err != nil {
		t.Errorf("unable to read stmts: %v", err)
		t.FailNow()
	}

	type args struct {
		db       db
		username string
		password string
	}
	tests := []struct {
		name      string
		args      args
		targetErr error
	}{
		{
			name: "doesn't exists, create fails",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{false},
					}, "nomen-user"),
					expectExec("-- replace nomen-user with the name of the user\nCREATE USER \"nomen-user\"", sql.ErrTxDone),
				),
				username: "nomen-user",
				password: "",
			},
			targetErr: sql.ErrTxDone,
		},
		{
			name: "correct without password",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{false},
					}, "nomen-user"),
					expectExec("-- replace nomen-user with the name of the user\nCREATE USER \"nomen-user\"", nil),
				),
				username: "nomen-user",
				password: "",
			},
			targetErr: nil,
		},
		{
			name: "correct with password",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{false},
					}, "nomen-user"),
					expectExec("-- replace nomen-user with the name of the user\nCREATE USER \"nomen-user\" WITH PASSWORD 'password'", nil),
				),
				username: "nomen-user",
				password: "password",
			},
			targetErr: nil,
		},
		{
			name: "correct with password containing percent and quote",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{false},
					}, "nomen-user"),
					expectExec("-- replace nomen-user with the name of the user\nCREATE USER \"nomen-user\" WITH PASSWORD 'p%''ass'", nil),
				),
				username: "nomen-user",
				password: "p%'ass",
			},
			targetErr: nil,
		},
		{
			name: "already exists in catalog, skip creation",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", nil, []string{"exists"}, [][]driver.Value{
						{true},
					}, "nomen-user"),
				),
				username: "nomen-user",
				password: "password",
			},
			targetErr: nil,
		},
		{
			name: "catalog check fails",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"postgres"},
					}),
					expectQuery("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", sql.ErrConnDone, []string{"exists"}, [][]driver.Value{}, "nomen-user"),
				),
				username: "nomen-user",
				password: "password",
			},
			targetErr: sql.ErrConnDone,
		},
		{
			name: "same user, skip create",
			args: args{
				db: prepareDB(t,
					expectQuery("SELECT current_user", nil, []string{"current_user"}, [][]driver.Value{
						{"nomen-user"},
					}),
				),
				username: "nomen-user",
			},
			targetErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyUser(tt.args.username, tt.args.password)(context.Background(), tt.args.db.db); !errors.Is(err, tt.targetErr) {
				t.Errorf("VerifyUser() error = %v, want: %v", err, tt.targetErr)
			}
			if err := tt.args.db.mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
