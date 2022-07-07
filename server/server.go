package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"

	_ "github.com/lib/pq"
)

var dbConn *sql.DB
var client http.Client

//Fonction d'initiation du cookie

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	client = http.Client{
		Jar: jar,
	}
}

//Connexion à la base de donnée

func getDBConn() (*sql.DB, error) {

	if dbConn != nil {
		return dbConn, nil
	}

	connStr := "user=julessore dbname=newdb sslmode=disable"
	newdb, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	dbConn = newdb
	return dbConn, nil
}

//Fonction main qui lance le server lié aux fichiers /etc/nginx/nginx.conf et /etc/named.conf

func main() {
	serv := http.DefaultServeMux
	RouterAddRoutes(serv)

	fmt.Printf("Starting server at port 8080\n")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	if dbConn != nil {
		dbConn.Close()
	}

}

//Création d'une chaine de charactères aléatoires de (len) nombres de charactères.

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

//Fonction d'affichage d'erreur

func ErrorCheck(err error) {
	if err != nil {
		panic(err.Error())
	}
}
