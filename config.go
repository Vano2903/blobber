package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Secret string `yaml:"secret"`
}

var conf Config

var (
	blobsTableQuery string = `
	CREATE TABLE IF NOT EXISTS blobber.blobs (
		ID INT auto_increment NOT NULL,
		ID_user INT NOT NULL,
		content MEDIUMTEXT NULL,
		added_date DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
		PRIMARY KEY(ID)
	);
	`
	usersTableQuery string = `
	CREATE TABLE IF NOT EXISTS users (
		ID INT auto_increment NOT NULL,
		username VARCHAR(20) NOT NULL,
		password CHAR(64) NOT NULL,
		description TEXT,
		PRIMARY KEY (ID)
	);
	`
	likesTableQuery string = `
	CREATE TABLE IF NOT EXISTS likes (
		ID INT auto_increment NOT NULL,
		ID_user INT NOT NULL,
		ID_blob INT NOT NULL,
		PRIMARY KEY (ID)
	);
	`
	followsTableQuery string = `
	CREATE TABLE IF NOT EXISTS follows (
		ID INT auto_increment NOT NULL,
		ID_user_follower INT NOT NULL,
		ID_user_followed INT NOT NULL,
		PRIMARY KEY (ID)
	);
	`
)

func init() {
	//read the config.yaml, parse it and load the config struct
	secret := os.Getenv("secret")
	// log.Println(secret
	if secret == "" {
		dat, err := ioutil.ReadFile("config.yaml")
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal([]byte(dat), &conf)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	} else {
		conf.Secret = secret
	}

	db, err := connectToDB()
	if err != nil {
		log.Println("connection to db failed")
		panic(err)
	}

	//create tables if they don't exist
	_, err = db.Exec(blobsTableQuery)
	if err != nil {
		log.Fatalf("blobs table creation failed: %s", err.Error())
	}

	_, err = db.Exec(usersTableQuery)
	if err != nil {
		log.Fatalf("users table creation failed: %s", err.Error())
	}

	_, err = db.Exec(likesTableQuery)
	if err != nil {
		log.Fatalf("likes table creation failed: %s", err.Error())
	}

	_, err = db.Exec(followsTableQuery)
	if err != nil {
		log.Fatalf("follows table creation failed: %s", err.Error())
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("db close failed: %s", err.Error())
		}
	}()

	log.Println("connection with db established")
}
