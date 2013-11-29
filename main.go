package main

import (
    "./models"
    "fmt"
    "net/http"
    "log"
    "errors"
    "crypto/sha512"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/coopernurse/gorp"
    "github.com/sirMackk/templates_ago"
)

//HandleFunc wrappers and settings
type HandleFunc func(w http.ResponseWriter, req *http.Request)

func logPanic(function HandleFunc) HandleFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        if r := recover(); r != nil {
            //Log to file too
            log.Println(r)
        }
        function(w, req)
    }
}

func setHeaders(function HandleFunc) HandleFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        function(w, req)
    }
}

func checkAuth(function HandleFunc) HandleFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        session, _ := store.Get(req, "gopi_media")
        if session.Values["loggedin"] == true {
            function(w, req)
        } else {
            http.Redirect(w, req, "/login", 403)
        }
    }
}

func HandleWrapper(function HandleFunc) HandleFunc {
    return logPanic(setHeaders(function))
}


//Actual handling functions
//media and ordinary users
func IndexMedia(w http.ResponseWriter, req *http.Request) {
    var media []models.Media
    _, err := dbmap.Select(&media, "select * from media order by Id")
    if err != nil { panic(err) }

    err = templates.Execute(w, media)
    if err != nil { panic(err) }
}

//func NewMedia(w http.ResponseWriter, req *http.Request) {
    //switch req.Method {
    //case "GET":
      //new_media := models.Media{}
      //err := templates.Execute(w, new_media)
      //if err != nil { panic(err) }
    //case "POST":
      //title := req.FormValue("title")
      //m_type := req.FormValue("type")
      //private := req.FormValue("private")
      //_, err := models.NewMedia(dbmap, uid, title, m_type, path, priv)
      //if err != nil { panic(err) }
      //http.Redirect(w, req, "/", 302)
    //}
//}

//admin
func IndexAdmin(w http.ResponseWriter, req *http.Request) {
    fmt.Fprint(w, "Admin Index")
}


//authentication
func Login(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "GET":
        templates.Execute(w, req)
    case "POST":
        username := req.FormValue("username")
        password := req.FormValue("password")
        if user, err := Authenticate(username, password); err != nil {
            session, _ := store.Get(req, "gopi_media")
            session.Values["username"] = username
            session.Values["user_id"] = user.Id
            session.Values["loggedin"] = true
            err := session.Save(req, w)
            //better err handling here yo
            if err != nil { panic(err) }
            http.Redirect(w, req, "/", 302)
        } else {
            http.Redirect(w, req, "/login", 403)
        }
    }
}

func Logout(w http.ResponseWriter, req *http.Request) {
    session, _ := store.Get(req, "gopi_media")
    delete(session.Values, "username")
    delete(session.Values, "user_id")
    http.Redirect(w, req, "/login", 302)
}

func Authenticate(username, password string) (*models.User, error) {
    var user models.User
    err := dbmap.SelectOne(&user, "select * from users where username = ?", username)
    if err != nil {
        return nil, errors.New("invalid usernam")
    }

    //unable to find user
    h := sha512.New()
    fmt.Fprint(h, password)
    log.Println(h.Sum(nil))
    log.Println(user.Password)
    if string(h.Sum(nil)) == user.Password {
        return &user, nil
    } else {
        return nil, errors.New("bad password")
    }
}

//globals
var dbmap *gorp.DbMap
var templates = templates_ago.NewTemplates()
var store = sessions.NewCookieStore([]byte("2igIIhbR8nDmkDVR5dUU56rgCEjxKPCJ"))

func setupDatabase() {
    var err error
    db, err := sql.Open("sqlite3", "./gopi_media.db")
    if err != nil { panic(err) }

    dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

    dbmap.AddTableWithName(models.User{}, "users").SetKeys(true, "Id")
    dbmap.AddTableWithName(models.Media{}, "media").SetKeys(true, "Id")

    err = dbmap.CreateTablesIfNotExists()
    if err != nil { panic(err) }
    fmt.Println("Database up")
}

func init() {
    //setup database and templates
    setupDatabase()
}

func main() {
    router := mux.NewRouter()

    //change wrappers
    router.HandleFunc("/login", logPanic(Login))
    router.HandleFunc("/logout", logPanic(Logout))

    router.HandleFunc("/", HandleWrapper(IndexMedia))
    //router.HandleFunc("/media/new", HandleWrapper(NewMedia))
    //router.HandleFunc("/media/{id}", HandleWrapper(ShowMedia))

    router.HandleFunc("/admin/", HandleWrapper(checkAuth(IndexAdmin)))
    //router.HnadleFunc("/admin/users/", HandleWrapper(IndexAdminUsers))
    //router.HandleFunc("/admin/media/", HandleWrapper(IndexAdminMedia))
    //router.HandleFunc("/admin/media/new", HandleWrapper(NewAdminMedia))

    http.Handle("/", router)
    http.ListenAndServe(":3000", nil)
}


