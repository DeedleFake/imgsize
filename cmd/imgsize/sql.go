package main

import (
	"fmt"
	"log"
	"os"

	"database/sql"
	_ "github.com/lib/pq"
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
	log.Printf("Upsert(%q)", hash)
	log.Printf("\t%q", url)
	log.Printf("\t%v", width)
	log.Printf("\t%v", height)
	log.Printf("\t%q", method)

	row := db.QueryRow(`select count(1) from images where hash=$1`, hash)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	switch count {
	case 0:
		_, err = db.Exec(`insert into images values ($1, $2, $3, $4, $5)`,
			url,
			width,
			height,
			method,
			hash,
		)
	case 1:
		_, err = db.Exec(`update images set url=$1, width=$2, height=$3, method=$4 where hash=$5`,
			url,
			width,
			height,
			method,
			hash,
		)
	default:
		return fmt.Errorf("Found %v rows, but expected <= 1", count)
	}

	return err
}

func SelectImage(hash string) (url string, width, height int, method string, err error) {
	row := db.QueryRow(`select * from images where hash=$1`, hash)
	err = row.Scan(&url, &width, &height, &method, &hash)

	log.Printf("Select(%q)", hash)
	log.Printf("\t%q", url)
	log.Printf("\t%v", width)
	log.Printf("\t%v", height)
	log.Printf("\t%q", method)

	return
}
