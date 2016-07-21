package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

// Retrieve SQL configuration from API gateway stage variables.
func GetSQLConfigFromAPIGateway(in GWInput) *mysql.Config {
	return &mysql.Config{
		User:              in.StageVar["DBUser"],
		Passwd:            in.StageVar["DBPass"],
		Net:               "tcp",
		Addr:              fmt.Sprintf("%s:%s", in.StageVar["DBHost"], in.StageVar["DBPort"]),
		DBName:            in.StageVar["DBName"],
		Collation:         "utf8_general_ci",
		Timeout:           30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		AllowOldPasswords: true,
	}
}

// Open a new database connection to run the operations, and then close the connection.
func DoSQL(config *mysql.Config, fun func(*sqlx.DB) error) error {
	db, err := sqlx.Open("mysql", config.FormatDSN())
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return fun(db)
}
