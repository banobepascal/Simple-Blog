package main

import (
	"html/template"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

type user struct {
	UserName   string
	Email      string
	Password   string
	Repassword string
}

var tpl *template.Template
var dbUsers = map[string]user{}
var dbSessions = map[string]string{}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {

	http.HandleFunc("/userPage", userPage)
	http.HandleFunc("/", signup)
	http.Handle("/css/", http.FileServer(http.Dir("public")))
	http.Handle("/img/", http.FileServer(http.Dir("public")))
	http.Handle("/js/", http.FileServer(http.Dir("public")))
	http.Handle("/fonts/", http.FileServer(http.Dir("public")))
	http.ListenAndServe(":8080", nil)
}

func userPage(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "index.html", u)
}

func signup(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	// process of form submission
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		e := req.FormValue("email")
		p := req.FormValue("password")
		rp := req.FormValue("repassword")

		// username token
		if _, ok := dbUsers[un]; ok {
			http.Error(w, "username is invalid", http.StatusSeeOther)
			return
		}

		// create session cookie
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un

		//store user in database
		u := user{un, e, p, rp}
		dbUsers[un] = u

		// Redirect
		http.Redirect(w, req, "/userPage", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "signup.html", nil)

}
