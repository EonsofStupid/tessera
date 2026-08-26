package database

//go:generate mockgen -typed -package dbmock -destination ./dbmock/database.mock.go github.com/shippinAI/nomen/backend/v3/storage/database Pool,Connection,Row,Rows,Transaction
