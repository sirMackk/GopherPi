package main

import (
    "./models"
    "./utils"
    "fmt"
    "strconv"
    "net/http"
    "log"
    "errors"
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
            http.Error(w, fmt.Sprintf("%v", r), 500)
            log.Println(r)
        }
        function(w, req)
    }
}

func setHeaders(function HandleFunc) HandleFunc {
    //setup to handle json here? yeah...
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
            http.Redirect(w, req, "/login", 403) //302
        }
    }
}

func HandleWrapper(function HandleFunc) HandleFunc {
    return logPanic(setHeaders(function))
}

func AuthWrapper(function HandleFunc) HandleFunc {
    return logPanic(checkAuth(setHeaders(function)))
}


//Actual handling functions
//media and ordinary users
func IndexMedia(w http.ResponseWriter, req *http.Request) {
    //search fn here
    var media []models.Media
    _, err := dbmap.Select(&media, "select * from media order by Id")
    if err != nil { panic(err) }

    err = templates.Execute(w, media)
    if err != nil { panic(err) }
}

func IndexOwnMedia(w http.ResponseWriter, req *http.Request) {
    session, _ := store.Get(req, "gopi_media")
    user_id := session.Values["user_id"]
    var media []models.Media
    _, err := dbmap.Select(&media, "select * from media where user_id = ? order by Id desc", user_id)
    if err != nil { panic(err) }
    templates["indexmedia.html"].ExecuteTemplate(w, "base", media)
}

func NewMedia(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "GET":
      templates.Execute(w, nil)
    case "POST":
      session, _ := store.Get(req, "gopi_media")
      media := models.NewMediaFromRequest(dbmap, req, fmt.Sprintf("%s", session.Values["user_id"]))
      http.Redirect(w, req, fmt.Sprintf("/media/%d", media.Id), 200)
    }
}


func ShowMedia(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    session, _ := store.Get(req, "gopi_media")
    var media models.Media
    err := dbmap.SelectOne(&media, "select * from media where Id = ?", vars["id"])
    if err != nil { panic(err) }
    if media.Private == true {
        if media.User_id == session.Values["user_id"] {
            if req.Method == "DELETE" {
                _, err := dbmap.Delete(&media)
                if err != nil { panic(err) }
                http.Redirect(w, req, "/media", 301)
            } else {
                templates.Execute(w, media)
            }
        } else {
            http.Error(w, "Verbotten", 403)
        }
    } else {
        if req.Method == "DELETE" {
            _, err := dbmap.Delete(&media)
            if err != nil { panic(err) }
            http.Redirect(w, req, "/media", 301)
        } else {
        templates.Execute(w, media)
        }
    }
}

func EditMedia(w http.ResponseWriter, req *http.Request) {
    session, _ := store.Get(req, "gopi_media")
    vars := mux.Vars(req)

    var user models.User
    _, err := dbmap.Select(&user, "select * from users where Id = ?", session.Values["user_id"])
    if err != nil { panic(err) }
    var media models.Media
    _, err = dbmap.Select(&media, "select * from media where Id = ?", vars["id"])
    if err != nil { panic(err) }

    switch req.Method {
    case "GET":
        if user.Admin == true || user.Id == media.User_id {
            templates["newmedia.html"].ExecuteTemplate(w, "base", media)
        } else {
            http.Error(w, "Verbotten!", 403)
        }
    case "POST":
        if user.Admin == true || user.Id == media.User_id {
            req.ParseForm()
            media.Title = req.Form["title"][0]
            private, err := strconv.ParseBool(req.Form["private"][0])
            media.Private = private
            _, err = dbmap.Update(&media)
            if err != nil { panic(err) }
            http.Redirect(w, req, fmt.Sprintf("/media/%d", media.Id), 301)
        } else {
            http.Error(w, "Verbotten!", 403)
        }
    }
}

//admin-funcs
//users
func IndexAdmin(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "GET":
        templates.Execute(w, nil)
    case "POST":
        req.ParseForm()
        switch req.Form["action"][0] {
            case "scan":
                //user_id - which user will get the files
                //directory - absolute path to media dir
                //priv setting - should media be private or public?
                uid := req.FormValue("user_id")
                directory := req.FormValue("directory")
                priv_setting := req.FormValue("priv_setting")
                utils.ScanMediaDir(dbmap, directory, uid, priv_setting)
            case "changemediadir":
                mediaDir = req.Form["directory"][0]

        }
    }

}


func IndexAdminUsers(w http.ResponseWriter, req *http.Request) {
    var users []models.User
    _, err := dbmap.Select(&users, "select * from users order by Id")
    if err != nil { panic(err) }
    err = templates.Execute(w, users)
    if err != nil { panic(err) }
}

func NewAdminUsers(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "GET":
      templates.Execute(w, nil)
    case "POST":
      //sum validations
      user := models.NewUserFromRequest(dbmap, req)
      http.Redirect(w, req, fmt.Sprintf("/admin/users/%d", user.Id), 301)
    }
}

func ShowAdminUsers(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    id := vars["id"]
    switch req.Method {
    case "GET":
      var user models.User
      err := dbmap.SelectOne(&user, "select * from users where Id = ?", id)
      if err != nil { panic(err) }
      templates.Execute(w, user)
    case "DELETE":
        fmt.Println("deleting")
        _, err := dbmap.Exec("delete from users where Id = ?", id)
        if err != nil { panic(err) }
        //http.Redirect(w, req, "/admin/users", 301)
    }
}

func EditAdminUsers(w http.ResponseWriter, req *http.Request) {
    //put put or post switch here
    vars := mux.Vars(req)
    id := vars["id"]
    var user models.User
    err := dbmap.SelectOne(&user, "select * from users where Id = ?", id)
    if err != nil { panic(err) }
    templates["newadminusers.html"].ExecuteTemplate(w, "base", user)
}

//authentication
func Login(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "GET":
        //med := new(models.User)
        templates.Execute(w, nil)
    case "POST":
        username := req.FormValue("username")
        password := req.FormValue("password")
        if user, err := Authenticate(username, password); err == nil {
            session, _ := store.Get(req, "gopi_media")
            session.Values["username"] = username
            session.Values["user_id"] = user.Id
            session.Values["loggedin"] = true
            err := session.Save(req, w)
            //better err handling here yo
            if err != nil { panic(err) }
            http.Redirect(w, req, "/", 302)
        } else {
            log.Println(err)
            http.Redirect(w, req, "/login", 403)
        }
    }
}

func Logout(w http.ResponseWriter, req *http.Request) {
    session, _ := store.Get(req, "gopi_media")
    delete(session.Values, "username")
    delete(session.Values, "user_id")
    delete(session.Values, "loggedin")
    session.Save(req, w)
    http.Redirect(w, req, "/login", 302)
}

func Authenticate(username, password string) (*models.User, error) {
    var user models.User
    err := dbmap.SelectOne(&user, "select * from users where username = ?", username)
    if err != nil {
        return nil, errors.New("invalid username")
    }

    //unable to find user
    pwd := models.HashPwd(password)
    if pwd == user.Password {
        return &user, nil
    } else {
        return nil, errors.New("bad password")
    }
}


//globals
var dbmap *gorp.DbMap
var templates = templates_ago.NewTemplates()
var store = sessions.NewCookieStore([]byte("2igIIhbR8nDmkDVR5dUU56rgCEjxKPCJ"))
var mediaDir = "users/"

func setupDatabase() {
    var err error
    db, err := sql.Open("sqlite3", "./gopi_media.db")
    if err != nil { panic(err) }

    dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

    dbmap.AddTableWithName(models.User{}, "users").SetKeys(true, "Id")
    dbmap.AddTableWithName(models.Media{}, "media").SetKeys(true, "Id")

    err = dbmap.CreateTablesIfNotExists()

    if err != nil { panic(err) }
    users, err := dbmap.SelectInt("select count(*) from users")
    if err != nil { panic(err) }
    if users == 0 {
        pwd := models.HashPwd("password")
        _, err := models.NewUser(dbmap, "admin", pwd)
        if err != nil {
            fmt.Println("Problem with creating user")
            panic(err)
        }
    }


    fmt.Println("Database up")
}

func init() {
    //setup database and templates
    setupDatabase()
    templates_ago.LoadTemplates("templates/", templates)
}

func main() {
    defer dbmap.Db.Close()
    router := mux.NewRouter()

    //change wrappers
    router.HandleFunc("/login", logPanic(Login))
    router.HandleFunc("/logout", logPanic(Logout))

    router.HandleFunc("/", HandleWrapper(IndexMedia))
    router.HandleFunc("/media", AuthWrapper(IndexOwnMedia))
    router.HandleFunc("/media/new", AuthWrapper(NewMedia))
    router.HandleFunc("/media/{id}", AuthWrapper(ShowMedia))
    router.HandleFunc("/media/{id}/edit", AuthWrapper(EditMedia))

    router.HandleFunc("/admin", AuthWrapper(IndexAdmin))
    router.HandleFunc("/admin/users", AuthWrapper(IndexAdminUsers))
    router.HandleFunc("/admin/users/new", AuthWrapper(NewAdminUsers))
    router.HandleFunc("/admin/users/{id}", AuthWrapper(ShowAdminUsers))
    router.HandleFunc("/admin/users/{id}/edit", AuthWrapper(EditAdminUsers))
    //router.HandleFunc("/admin/media/", HandleWrapper(IndexAdminMedia))
    //router.HandleFunc("/admin/media/new", HandleWrapper(NewAdminMedia))

    router.PathPrefix("/users/").Handler(http.StripPrefix("/users/", http.FileServer(http.Dir(mediaDir))))
    fmt.Println("routes set, about to handle")
    http.Handle("/", router)
    err := http.ListenAndServe(":3000", nil)
    if err != nil { panic(err) }
}

