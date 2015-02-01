package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"os"
)

var db *sql.DB

func init() {
	tmpDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	db = tmpDB
}

func UpsertImage(hash string, url string, width int, height int, method string) error {
	row := db.QueryRow(`select 1 from images where hash=?`, hash)
	err := row.Scan()
	switch err {
	case sql.ErrNoRows:
		_, err = db.Exec(`insert into images values (?, ?, ?, ?, ?)`,
			url,
			width,
			height,
			method,
			hash,
		)
		return err
	default:
		return err
	}

	_, err = db.Exec(`update images set url=?, width=?, height=?, method=?  where hash=?`,
		url,
		width,
		height,
		method,
		hash,
	)
	return err
}

func SelectImage(hash string) (url string, width, height int, method string, err error) {
	row := db.QueryRow(`select * from images where hash=?`, hash)
	err = row.Scan(url, width, height, method)

	return
}
