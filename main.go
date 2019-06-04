package main

import (
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	uuid "github.com/satori/go.uuid"
)

type user struct {
	UserName string
	Email    string
	Password []byte
}

var tpl *template.Template
var dbUsers = map[string]user{}
var dbSessions = map[string]string{}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	bs, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	dbUsers["badboy"] = user{"badboy", "pascal@test.com", bs}
	dbUsers["ham"] = user{"ham", "ham@test.com", bs}
	dbUsers["liz"] = user{"liz", "liz@test.com", bs}

}

func main() {

	http.HandleFunc("/userPage", userPage)
	http.HandleFunc("/", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
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

		// store user in dbUsers
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		u := user{un, e, bs}
		dbUsers[un] = u

		// Redirect
		http.Redirect(w, req, "/userPage", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "signup.html", nil)

}

func login(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	// process of form submission
	if req.Method == http.MethodPost {
		//get form values
		un := req.FormValue("username")
		p := req.FormValue("password")

		//username token
		u, ok := dbUsers[un]
		if !ok {
			http.Error(w, "username and/or password mismatch", http.StatusForbidden)
			return
		}

		// Username and password matching
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		if err != nil {
			http.Error(w, "username and password invalid", http.StatusForbidden)
			return
		}

		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		http.Redirect(w, req, "/userPage", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "login.html", nil)

}

func logout(w http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	c, _ := req.Cookie("session")
	// delete cookie
	delete(dbSessions, c.Value)
	// remove cookie
	c = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
	return

}
