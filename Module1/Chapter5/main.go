package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"html/template"
	"log"
	"net/http"
	"os"
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
	Id         string
	Title      string
	RawContent string
	Content    template.HTML
	Date       string
	Comments   []Comment
	GUID       string
}

type JSONResponse struct {
	Id    int64
	Added bool
}

type Comment struct {
	Id          int
	Name        string
	Email       string
	CommentText string
}

func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT id, page_title,page_content,page_date,page_guid FROM pages WHERE page_guid=?", pageGUID).
		Scan(&thisPage.Id, &thisPage.Title, &thisPage.RawContent, &thisPage.Date, &thisPage.GUID)
	thisPage.Content = template.HTML(thisPage.RawContent)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
	} else {
		comments, err := database.Query("SELECT id, comment_name,comment_email, comment_text FROM comments WHERE page_id=?",
			thisPage.Id)
		if err != nil {
			log.Println(err)
		}
		for comments.Next() {
			var comment Comment
			comments.Scan(&comment.Id, &comment.Name, &comment.Email, &comment.CommentText)
			thisPage.Comments = append(thisPage.Comments, comment)
		}
		t, _ := template.ParseFiles("templates/blog.html")
		t.Execute(w, thisPage)
	}

}

func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 301)
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Pages = []Page{}
	pages, err := database.Query("SELECT page_title,page_content,page_date,page_guid FROM pages ORDER BY ? DESC", "page_date")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
	defer pages.Close()

	for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.RawContent,
			&thisPage.Date, &thisPage.GUID)
		thisPage.Content = template.HTML(thisPage.RawContent)
		Pages = append(Pages, thisPage)
	}

	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
}

func (p Page) TruncatedText() string {
	chars := 0
	for i, _ := range p.RawContent {
		chars++
		if chars > 120 {
			return p.RawContent[:i] + ` ...`
		}
	}
	return p.RawContent
}

func APIPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT page_title,page_content,page_date,page_guid FROM pages WHERE page_guid=?", pageGUID).
		Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date, &thisPage.GUID)
	thisPage.Content = template.HTML(thisPage.RawContent)
	fmt.Println(err)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println(err)
		return
	}
	APIOutput, err := json.Marshal(thisPage)
	fmt.Println(string(APIOutput))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(thisPage)
}

func APICommentPost(w http.ResponseWriter, r *http.Request) {
	var commentAdded bool
	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
	}
	pageId := r.FormValue("pageId")
	fmt.Println(pageId)
	comment_guid := r.FormValue("guid")
	name := r.FormValue("name")
	email := r.FormValue("email")
	comments := r.FormValue("comments")
	res, err := database.Exec("INSERT INTO comments SET page_id=?,comment_guid=?, comment_name=?, comment_email=?, comment_text=?",
		pageId, comment_guid, name, email, comments)
	if err != nil {
		log.Println(err.Error())
	}
	id, err := res.LastInsertId()
	if err != nil {
		commentAdded = false
	} else {
		commentAdded = true
	}

	var resp JSONResponse
	resp.Id = id
	resp.Added = commentAdded

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func APICommentPut(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
	}
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println(id)
	guid := r.FormValue("guid")
	name := r.FormValue("name")
	email := r.FormValue("email")
	comments := r.FormValue("comments")
	res, err := database.Exec("UPDATE comments SET comment_name=?,comment_email=?, comment_text=? WHERE id=?",
		name, email, comments, id)
	fmt.Println(res)
	if err != nil {
		log.Println(err.Error())
	}
	http.Redirect(w, r, "/page/"+guid, http.StatusSeeOther)
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

	// certificates, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	// tlsConf := tls.Config{Certificates: []tls.Certificate{certificates}}
	// tls.Listen("tcp", PORT, &tlsConf)

	routes := mux.NewRouter()
	routes.HandleFunc("/api/pages", APIPage).Methods("GET").Schemes("http")
	routes.HandleFunc("/api/pages/{guid:[0-9a-zA\\-]+}", APIPage).Methods("GET").Schemes("http")
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
	routes.HandleFunc("/home", ServeIndex)
	routes.HandleFunc("/api/comments", APICommentPost).Methods("POST")
	routes.HandleFunc("/api/comments/{id:[\\w\\d\\-]+}", APICommentPut).Methods("PUT")
	http.Handle("/", routes)
	log.Fatal(http.ListenAndServe(PORT, nil))

}
