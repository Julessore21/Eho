package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	_ "github.com/lib/pq"
	"html/template"
	"time"
)

type Post struct {
	Title string
	Textpost string
	Id int
	CreationTime string
}

type BlogData struct {
	Title string
	Posts []Post
}

func RouterAddBlogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/post", PostHandler)
	mux.HandleFunc("/blog", BlogHandler)
	mux.HandleFunc("/acceuil", HomeHandler)

}


//Gestion des erreurs de serveur

func helloHandler(w http.ResponseWriter, r *http.Request) {
	
	
	if r.URL.Path != "/hello" {
		http.Error(w, "404 is not supported.", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	
	fmt.Fprintf(w, "Hello!")
}

//Gestionnaire de l'acceuil

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := GetPosts()
	if err != nil{
		fmt.Printf("Impossible de recuperer le post")
	}

	var tmplData BlogData = BlogData{
		Title: "titre de la page",
		Posts: posts,
	}
	
	t := template.Must(template.New("acceuil.html").ParseFiles("tmpl/acceuil.html"))	
	err = t.ExecuteTemplate(w, "acceuil", tmplData)
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to execute template acceuil.html: %s\n", err.Error())
	}
	
}

//Fonction de configuration de la page de création de post

func PostHandler(w http.ResponseWriter, r *http.Request) {

	cookie, _ := r.Cookie("CookieSession")
	sessioncookie, err := GetSessionByToken(cookie.Value)

	if err != nil{
		fmt.Printf("Erreur de récupération du token")
	}
	
	usersession, err := GetUserById(sessioncookie.Id)
	
	if err != nil{
		fmt.Printf("Erreur de récupération du user")
	}
	
	currentTime := time.Now()
	
	id := usersession.Id
	title := r.FormValue("title")
	textpost := r.FormValue("textpost")
	creationdate := currentTime.Format("2006.01.02 15:04:05")

	var userpost Post = Post{
		Id: id,
		Title: title,
		Textpost: textpost,
		CreationTime: creationdate,
	}
	er := CreatePost(userpost)
	
	if er != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create user: %s\n", er.Error())
	} else {
		fmt.Printf("Post Created !")
		url := fmt.Sprintf("%s/blog", ExternalUrl)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

//Gestionnaire du Blog

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	
	cookie, _ := r.Cookie("CookieSession")
	
	sessioncookie, err := GetSessionByToken(cookie.Value)

	if err != nil{
		fmt.Printf("Erreur de récupération du token")
	}
	
	usersession, err := GetUserById(sessioncookie.Id)
	
	if err != nil{
		fmt.Printf("Erreur de récupération du user")
	}
	
	userpost, err := GetPostById(usersession.Id)
	if err != nil{
		fmt.Printf("Impossible de recuperer le post")
	}

	var tmplData BlogData = BlogData{
		Title: "titre de la page",
		Posts: userpost,
	}	
	t := template.Must(template.New("blog.html").ParseFiles("tmpl/blog.html"))	
	err = t.ExecuteTemplate(w, "blog", tmplData)
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to execute template blog.html: %s\n", err.Error())
	}
	
}


//Creation d'un post

func CreatePost(userpost Post) error{
	newdb, err := getDBConn()
	if err != nil {
		return err
	}
	_, err = newdb.Exec(`INSERT INTO post (Id, Title, Textpost, CreationTime) VALUES($1, $2, $3, $4)`, userpost.Id, userpost.Title, userpost.Textpost, userpost.CreationTime)
	if err != nil {
		return err
	}
	return nil
}


//Obtenir un post grace à l'id

func GetPostById(Id int) ([]Post,error) {
	
	newdb, err := getDBConn()
	
	rows, err := newdb.Query(`SELECT Title, Textpost, CreationTime FROM Post WHERE Id=$1`, Id)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	
	defer rows.Close()
	var title, textpost, creationtime string

	var posts []Post
	for rows.Next() {
		if err := rows.Scan(&title, &textpost, &creationtime); err != nil {
			log.Fatal(err)
			return nil, err
		}
		var findpostbyid Post = Post{
			Title: title,
			Textpost: textpost,
			CreationTime: creationtime,	
		}	
		posts = append(posts, findpostbyid)
	}
	return posts, nil
}

func GetPosts() ([]Post,error) {
	
	newdb, err := getDBConn()
	
	rows, err := newdb.Query(`SELECT Title, Textpost, CreationTime FROM Post`)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}	
	defer rows.Close()
	var title, textpost, creationtime string

	var posts []Post
	for rows.Next() {
		if err := rows.Scan(&title, &textpost, &creationtime); err != nil {
			log.Fatal(err)
			return nil, err
		}
		var findpostbyid Post = Post{
			Title: title,
			Textpost: textpost,
			CreationTime: creationtime,	
		}
		posts = append(posts, findpostbyid)
	}
	return posts, nil
}
