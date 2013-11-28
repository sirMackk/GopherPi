package main

import (
    "./models"
    "fmt"
    "net/http"
    "log"
    "github.com/gorilla/mux"
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

//globals
var dbmap *gorp.DbMap
var templates = templates_ago.NewTemplates()

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

    router.HandleFunc("/", HandleWrapper(IndexMedia))
    //router.HandleFunc("/media/new", HandleWrapper(NewMedia))
    //router.HandleFunc("/media/{id}", HandleWrapper(ShowMedia))

    router.HandleFunc("/admin/", HandleWrapper(IndexAdmin))
    router.HnadleFunc("/admin/users/", HandleWrapper(IndexAdminUsers))
    router.HandleFunc("/admin/media/", HandleWrapper(IndexAdminMedia))
    router.HandleFunc("/admin/media/new", HandleWrapper(NewAdminMedia))

    http.Handle("/", router)
    http.ListenAndServe(":3000", nil)
}


