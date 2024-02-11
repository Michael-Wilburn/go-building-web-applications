package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
)

const (
	DBHost  = "127.0.0.1"
	DBPort  = ":3306"
	DBUser  = "root"
	DBDbase = "test"
	PORT    = ":8080"
)

var database *sql.DB

type Page struct {
	Title   string
	Content string
	Date    string
}

func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)

	err := database.QueryRow("SELECT page_title,page_content,page_date FROM pages WHERE page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
	} else {
		html := `<html>
				<head><title>` + thisPage.Title + `</title></head>
				<body>
				<h1>` + thisPage.Title + `</h1>
				<div>` + thisPage.Content + `</div>
				</body>
				</html>`
		fmt.Fprintln(w, html)
	}

}

func main() {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Error reading .env")
	}
	password, exists := os.LookupEnv("MYSQL_PASSWORD")
	if !exists {
		fmt.Println("Error PASSWORD not found")
	}

	DBPass := password

	dbConn := fmt.Sprintf("%s:%s@tcp(%s)/%s", DBUser, DBPass, DBHost, DBDbase)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		log.Println("Couldn't connect!")
		log.Println(err.Error())
	}
	database = db

	routes := mux.NewRouter()
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
	http.Handle("/", routes)
	log.Fatal(http.ListenAndServe(PORT, nil))
}
