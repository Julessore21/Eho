package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

//Definition des structures

type utilisateur struct {
	Name    string
	Mdp     string
	Address string
	Phone   string
	Mail    string
	Birth   string
	Id      int
}

type session struct {
	Id           int
	Token        string
	CreationDate string
}

type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string
	MaxAge     int
	Secure     bool
	HttpOnly   bool
	Raw        string
	Unparsed   []string
}

//Mise en place des routes liées au router

func RouterAddUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/form", formHandler)
	mux.HandleFunc("/login", connectionHandler)
	mux.HandleFunc("/account", AccountHandler)
	mux.HandleFunc("/user/edit", EditUserHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			url := fmt.Sprintf("%s/static/erreur404/", ExternalUrl)
			http.Redirect(w, req, url, http.StatusFound)
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})
}

func formHandler(w http.ResponseWriter, r *http.Request) {

	//Récupération des informations du formulaire

	name := r.FormValue("name")
	mdp := r.FormValue("mdp")
	address := r.FormValue("address")
	phone := r.FormValue("phone")
	mail := r.FormValue("mail")
	birth := r.FormValue("birth")

	//Création du sel

	salt := randomString(16)
	log.Printf("Le sel :%s\n", salt)

	//Création du hash du mot de passe + le sel

	sum := sha256.Sum256([]byte(salt + mdp))
	hash := base64.StdEncoding.EncodeToString(sum[:])

	//Ajout de la $méthode$sel$hash

	salthash := fmt.Sprintf("$5$%s$%s", salt, hash)

	//Renvoie les informations sur la page form

	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "mdp = %s\n", salthash)
	fmt.Fprintf(w, "Address = %s\n", address)
	fmt.Fprintf(w, "phone = %s\n", phone)
	fmt.Fprintf(w, "mail = %s\n", mail)
	fmt.Fprintf(w, "birth = %s\n", birth)

	//Création d'un nouvel utilisateur à partir des information du formulaire

	var user utilisateur = utilisateur{
		Name:    name,
		Mdp:     salthash,
		Address: address,
		Phone:   phone,
		Mail:    mail,
		Birth:   birth,
	}
	err := CreateUser(user)

	//Erreur/Réussite

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create user: %s\n", err.Error())
	} else {
		fmt.Printf("User Created !")
	}
}

//TODO : session / cookie connexion, disparition des (connexion,inscription), template pour Header / login

//Gestion de la page de connexion

func connectionHandler(w http.ResponseWriter, r *http.Request) {

	//Initiation de la seed aléatoire

	rand.Seed(time.Now().UnixNano())

	//Connexion à la base de donnée

	newdb, err := getDBConn()

	if err != nil {
		fmt.Printf("Error 404")
	}

	//Initiation de l'heure et de la date

	currentTime := time.Now()

	//Récupérations des logs

	name := r.FormValue("name")
	mdp := r.FormValue("mdp")

	//Requète sql pour récupérer l'id correspondant dans la base de donnée

	var id int
	rows, err := newdb.Query(`SELECT Id FROM utilisateur WHERE name=$1`, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query failed: %s\n", err.Error())
	}
	defer rows.Close()
	var found = false
	for rows.Next() {
		found = true
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
	}

	//Gestion erreurs

	if !found {
		fmt.Fprintf(os.Stderr, "error: can't get user from user table\n")
		return
	}

	//Récupération du user grace à l'id

	UserConnected, err := GetUserById(id)
	if err != nil {
		fmt.Printf("Error 404")
	}

	//sel + hash du mot de passe pour le comparer à celui présent dans la base de donnée

	tabmdp := strings.Split(UserConnected.Mdp, "$")
	if len(tabmdp) < 4 {
		fmt.Fprintf(os.Stderr, "error: malformed user password in database\n")
		return
	}
	saltuser := tabmdp[2]
	sum := sha256.Sum256([]byte(saltuser + mdp)) //Ash
	hash := base64.StdEncoding.EncodeToString(sum[:])
	Connectedhash := tabmdp[3]

	//Creation du token pour la session

	creationdate := currentTime.Format("2006.01.02 15:04:05")
	token := sha256.Sum256([]byte(creationdate)) //HashToken
	hashtoken := base64.StdEncoding.EncodeToString(token[:])

	var sessionutilisateur session = session{
		Id:           UserConnected.Id,
		Token:        hashtoken,
		CreationDate: creationdate,
	}

	//Comparaison mot de passe >> Creation de la session

	if Connectedhash == hash {

		//Verification d'un possible doublon de session

		var sessionexist bool
		sessionexist = SessionAlreadyExist(sessionutilisateur.Id)
		if sessionexist == false {
			err := CreateSession(sessionutilisateur)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to create session: %s\n", err.Error())
			}
		} else {
			DeleteSession(sessionutilisateur.Id)
			err := CreateSession(sessionutilisateur)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to create session: %s\n", err.Error())
			}
		}
		fmt.Printf("Vous avez réussis à vous connecter à %s !\n", name)

		//Creation / set du cookie

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "CookieSession", Value: sessionutilisateur.Token, Expires: expiration}
		http.SetCookie(w, &cookie)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		url := fmt.Sprintf("%s/account", ExternalUrl)
		http.Redirect(w, r, url, http.StatusFound)
	} else {
		fmt.Printf("Wrong Password !\n")
	}
}

//Gestion de la page account (profil)

func AccountHandler(w http.ResponseWriter, r *http.Request) {

	//Recuperation du cookie >> user >> nom.user

	cookie, _ := r.Cookie("CookieSession")

	sessioncookie, err := GetSessionByToken(cookie.Value)

	usersession, err := GetUserById(sessioncookie.Id)

	name := usersession.Name

	//Connection db + requete pour recup l'utilisateur lié

	newdb, err := getDBConn()

	rows, err := newdb.Query(`SELECT address, phone, mail, birth FROM utilisateur WHERE Name=$1`, name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var address, phone, mail, birth string
	for rows.Next() {
		if err := rows.Scan(&address, &phone, &mail, &birth); err != nil {
			log.Fatal(err)
		}
	}

	//Remplissage d'une instance de user avec les informations recup dans la db

	var user utilisateur = utilisateur{
		Name:    name,
		Address: address,
		Phone:   phone,
		Mail:    mail,
		Birth:   birth,
	}

	//Lancement du template avec les informations de user et la page account

	t := template.Must(template.New("account.html").ParseFiles("tmpl/account.html"))
	err = t.Execute(w, user)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to execute template account.html: %s\n", err.Error())
	}
}

//Mofification du user dans la page account

func EditUserHandler(w http.ResponseWriter, r *http.Request) {

	//Formulaire pour modif son profil

	newname := r.FormValue("name")
	newaddress := r.FormValue("address")
	newphone := r.FormValue("phone")
	newmail := r.FormValue("mail")
	newbirth := r.FormValue("birth")

	newdb, err := getDBConn()
	if err != nil {
		return
	}

	//Recup du user dans la db

	cookie, _ := r.Cookie("CookieSession")

	sessioncookie, err := GetSessionByToken(cookie.Value)

	//Requete pour update l'utilidsateur avec les nouvelles infos

	stmt, e := newdb.Prepare("UPDATE utilisateur SET name = $1, address = $2, phone = $3,  mail = $4, birth = $5 WHERE id=$6;")
	ErrorCheck(e)
	res, e := stmt.Exec(newname, newaddress, newphone, newmail, newbirth, sessioncookie.Id)
	ErrorCheck(e)
	a, e := res.RowsAffected()
	ErrorCheck(e)
	if a != 1 {
		fmt.Printf("Erreur de modification")
	}
	log.Print(a)
	http.Redirect(w, r, "https:/blogotin.fr/account", http.StatusFound)
}

//Création d'un utilisateur

func CreateUser(user utilisateur) error {

	newdb, err := getDBConn()
	if err != nil {
		return err
	}

	_, err = newdb.Exec(`INSERT INTO utilisateur(name, mdp, address, phone, mail, birth)
	VALUES($1, $2, $3, $4, $5, $6)`, user.Name, user.Mdp, user.Address, user.Phone, user.Mail, user.Birth)

	if err != nil {
		return err
	}

	return nil
}

//Création d'une session

func CreateSession(sessionutilisateur session) error {
	newdb, err := getDBConn()
	if err != nil {
		return err
	}
	_, err = newdb.Exec(`INSERT INTO session(Id, Token, CreationDate) VALUES($1, $2, $3)`, sessionutilisateur.Id, sessionutilisateur.Token, sessionutilisateur.CreationDate)
	if err != nil {
		return err
	}
	return nil
}

//Obtenir l'utilisateur grace au nom

func GetUserByName(name string) (utilisateur, error) {

	newdb, err := getDBConn()

	rows, err := newdb.Query(`SELECT id, mdp, address, phone, mail, birth FROM utilisateur WHERE name=$1`, name)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var id int
	var mdp, address, phone, mail, birth string
	for rows.Next() {
		if err := rows.Scan(&id, &mdp, &address, &phone, &mail, &birth); err != nil {
			log.Fatal(err)
		}
	}

	var finduserbyname utilisateur = utilisateur{
		Name:    name,
		Id:      id,
		Mdp:     mdp,
		Address: address,
		Phone:   phone,
		Mail:    mail,
		Birth:   birth,
	}

	return finduserbyname, nil
}

//ou à l'id

func GetUserById(Id int) (utilisateur, error) {

	newdb, err := getDBConn()

	rows, err := newdb.Query(`SELECT name, mdp, address, phone, mail, birth FROM utilisateur WHERE Id=$1`, Id)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var Name string
	var mdp, address, phone, mail, birth string

	for rows.Next() {
		if err := rows.Scan(&Name, &mdp, &address, &phone, &mail, &birth); err != nil {
			log.Fatal(err)
		}
	}

	var finduserbyid utilisateur = utilisateur{
		Name:    Name,
		Id:      Id,
		Mdp:     mdp,
		Address: address,
		Phone:   phone,
		Mail:    mail,
		Birth:   birth,
	}

	return finduserbyid, nil
}

//Obtenir la session grace au token

func GetSessionByToken(token string) (session, error) {
	newdb, err := getDBConn()
	var id int
	var creationdate string
	rows, err := newdb.Query(`SELECT id, creationdate FROM session  WHERE token=$1`, token)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&id, &creationdate); err != nil {
			log.Fatal(err)
		}
	}
	var findsessionbytoken session = session{
		Id:           id,
		Token:        token,
		CreationDate: creationdate,
	}
	return findsessionbytoken, nil
}

//Verifie si une session existe sur une id en particulier

func SessionAlreadyExist(Id int) bool {
	newdb, err := getDBConn()
	var alreadyexist bool
	rows, err := newdb.Query(`SELECT Token FROM session WHERE id = $1`, Id)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var token string

	for rows.Next() {
		if err := rows.Scan(&token); err != nil {
			log.Fatal(err)
		}
	}

	if token == "" {
		alreadyexist = false
	} else {
		alreadyexist = true
	}

	return alreadyexist
}

//Fonction pour supprimer une session

func DeleteSession(Id int) {
	newdb, err := getDBConn()
	_, err = newdb.Exec(`DELETE FROM session  WHERE id=$1`, Id)
	if err != nil {
		log.Fatal(err)
	}
}
