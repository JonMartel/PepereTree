package webserver

import (
	"fmt"
	"log"
	"strings"
	"time"

	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	listenAddr  = ":8443"
	certificate = "server.pem"
	keyfile     = "server.key"
)

type templateArgs struct {
	//nothing!
}

var (
	startTime     time.Time
	router        *httprouter.Router
	loginTemplate *template.Template
)

func init() {
	//Read in our templates
	loginTemplate = template.Must(template.ParseFiles("../login.html"))

	//Set up our router
	router = httprouter.New()                                      // creates a new router
	router.GET("/", rootHandler)                                   // will direct the GET / request to the Index function
	router.GET("/hello/:name", hello)                              // will redirect the GET /name to Hello, stores the name of the parameter in the a variable of httprouter.Params
	router.GET("/resources/*filepath", authCheck(resourceHandler)) // raw resources stored here!
	router.GET("/login", loginHandler)                             // provide the login template
	router.POST("/login", loginRequestHandler)                     // handle the post from the form
}

func loginHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	//TODO JMARTEL validate they aren't already logged in
	//if they are, just redirect to main. if not, display login

	page := templateArgs{}
	loginTemplate.Execute(w, page)
}

func loginRequestHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	r.ParseForm()
	/*user, pass := r.Form.Get("username"), r.Form.Get("password")
	fmt.Println("Attempt to log in as: ", user, " using password: ", pass)

	conn, _ := db.NewConnection("peperetree")
	valid, err := conn.ValidateUser(user, pass)

	if err == nil && valid {
		fmt.Println("User is valid: ", valid)

		//session := AddSession(user)
		//updateAuthCookie(w, session.GetToken())
		http.Redirect(w, r, "/resources/index.html", http.StatusFound)
	} else {
		fmt.Println("Auth rejected, redirecting")
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	*/
}

func updateAuthCookie(w http.ResponseWriter, authtoken string) {
	//session := TouchSession(authtoken)
	//if session != nil {
	//cookie := http.Cookie{Name: "authtoken", Value: session.GetToken(), Expires: session.GetExpiration(), HttpOnly: true, Path: "/"}
	//http.SetCookie(w, &cookie)
	//}
}

func authCheck(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("Auth Check")
		authcookie, err := r.Cookie("authtoken")
		if err != nil {
			fmt.Println("Error retrieving auth token", err)
		}

		//TODO JMARTEL actually validate this!
		validToken := true

		if authcookie != nil && validToken {
			fmt.Println("Validated cookie, value: ", authcookie.Value)
			//Update the cookie (to reset expire time)
			updateAuthCookie(w, authcookie.Value)
			// Delegate request to the given handle
			h(w, r, ps)
		} else {
			// Redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func resourceHandler(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	filepath := "resources" + param.ByName("filepath")

	data, err := ioutil.ReadFile(filepath)

	if err == nil {
		var contentType string = "text/plain"
		if strings.HasSuffix(filepath, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(filepath, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(filepath, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(filepath, "png") {
			contentType = "image/png"
		}

		w.Header().Add("Content Type", contentType)
		w.Write(data)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Bad times ahead - " + http.StatusText(404)))
	}
}

func hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

//Run : Starts up our webserver and begins handling requests
func Run() {
	startTime = time.Now()
	log.Fatalln(http.ListenAndServeTLS(listenAddr, certificate, keyfile, router))
}
