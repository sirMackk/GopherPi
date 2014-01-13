package main

import (
    "./models"
    "./utils"
    "fmt"
    "strconv"
    "os"
    "net/http"
    "log"
    "errors"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/coopernurse/gorp"
    "github.com/sirMackk/templates_ago"
)

//HandleFunc wrappers and settings
type HandleFunc func(w http.ResponseWriter, req *http.Request)
type AuthHandleFunc func(w http.ResponseWriter, req *http.Request, c Context)
type Context map[string]interface{}

func logPanic(function HandleFunc) HandleFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        if r := recover(); r != nil {
            http.Error(w, fmt.Sprintf("%v", r), 500)
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

func setContext(function AuthHandleFunc) HandleFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        session, _ := store.Get(req, APP_NAME)
        context := make(Context)
        if session.Values["loggedin"] == true {
            var user models.User
            err := dbmap.SelectOne(&user, "select * from users where Id = ?", session.Values["user_id"])
            context["User"] = &user
            if err != nil { panic(err) }
        }
        function(w, req, context)
    }
}

func checkAuth(function AuthHandleFunc) AuthHandleFunc {
    return func(w http.ResponseWriter, req *http.Request, c Context) {
        if _, ok := c["User"]; ok {
            function(w, req, c)
        } else {
            http.Redirect(w, req, "/login", 302)
        }
    }
}

func checkAdminAuth(function AuthHandleFunc) AuthHandleFunc {
    return func(w http.ResponseWriter, req *http.Request, c Context) {
        if c["User"].(*models.User).Admin {
            function(w, req, c)
        } else {
            http.Redirect(w, req, "/", 302)
        }
    }
}

func HandleWrapper(function AuthHandleFunc) HandleFunc {
    return logPanic(setHeaders(setContext(function)))
}

func AuthWrapper(function AuthHandleFunc) HandleFunc {
    return logPanic(setHeaders(setContext(checkAuth(function))))
}

func AdminAuthWrapper(function AuthHandleFunc) HandleFunc {
    return  AuthWrapper(checkAdminAuth(function))
}


//Actual handling functions
//static files
func StaticHandler(w http.ResponseWriter, req *http.Request) {
    muxVars := mux.Vars(req)
    assetPath := fmt.Sprintf("%s%s", STATIC_PATH, muxVars["asset"])
    if _, err := os.Stat(assetPath); os.IsNotExist(err) {
        http.Error(w, "Not found", 404)
    } else {
        http.ServeFile(w, req, assetPath)
    }
}

//about
func About(w http.ResponseWriter, req *http.Request, c Context) {
    templates["about.html"].ExecuteTemplate(w, "base", c)
    //err := templates.Execute(w, make(map[string]string));
    //if err != nil { panic(err) }
}

//media and ordinary users
func IndexMedia(w http.ResponseWriter, req *http.Request, c Context) {
    log.Println(c["User"])
    var media []models.Media
    _, err := dbmap.Select(&media, "select * from media order by Id")
    if err != nil { panic(err) }
    c["Media"] = media
    err = templates.Execute(w, c)
    if err != nil { panic(err) }
}

func IndexOwnMedia(w http.ResponseWriter, req *http.Request, c Context) {
    user_id := c["User"].(*models.User).Id
    var media []models.Media
    _, err := dbmap.Select(&media, "select * from media where user_id = ? order by Id desc", user_id)
    c["Media"] = &media
    if err != nil { panic(err) }
    err = templates.Execute(w, c)
    if err != nil { panic(err) }
}

func NewMedia(w http.ResponseWriter, req *http.Request, c Context) {
    switch req.Method {
    case "GET":
      templates.Execute(w, nil)
    case "POST":
      session, _ := store.Get(req, APP_NAME)
      media := models.NewMediaFromRequest(dbmap, req, fmt.Sprintf("%d", session.Values["user_id"]))
      http.Redirect(w, req, fmt.Sprintf("/media/%d", media.Id), 200)
    }
}


func ShowMedia(w http.ResponseWriter, req *http.Request, c Context) {
    vars := mux.Vars(req)
    var media models.Media
    err := dbmap.SelectOne(&media, "select * from media where Id = ?", vars["id"])
    if err != nil { panic(err) }
    c["Media"] = media
    if media.Private == true {
        if media.User_id == c["User"].(*models.User).Id {
            if req.Method == "DELETE" {
                _, err := dbmap.Delete(&media)
                if err != nil { panic(err) }
                http.Redirect(w, req, "/media", 301)
            } else {
                templates.Execute(w, c)
            }
        } else {
            http.Error(w, "Verbotten", 403)
        }
    } else {
        if req.Method == "DELETE" {
            _, err := dbmap.Delete(&media)
            if err != nil { panic(err) }
            log.Println(fmt.Sprintf("Deleting media %d - %s", media.Id, media.Title))
        } else {
            templates.Execute(w, c)
        }
    }
}

func EditMedia(w http.ResponseWriter, req *http.Request, c Context) {
    user := c["User"].(*models.User)
    vars := mux.Vars(req)

    var media models.Media
    err := dbmap.SelectOne(&media, "select * from media where Id = ?", vars["id"])
    if err != nil { panic(err) }
    c["Media"] = media

    switch req.Method {
    case "GET":
        if user.Admin == true || user.Id == media.User_id {
            templates["newmedia.html"].ExecuteTemplate(w, "base", c)
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
func IndexAdmin(w http.ResponseWriter, req *http.Request, c Context) {
    switch req.Method {
    case "GET":
        templates.Execute(w, c)
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
                log.Println(fmt.Sprintf("Scanning %s to add media to user id: %s", directory, uid))
            case "prune":
                log.Println("Pruning existing media")
                utils.PruneMedia(dbmap)
            http.Redirect(w, req, "/home", 301)
        }
    }

}


func IndexAdminUsers(w http.ResponseWriter, req *http.Request, c Context) {
    var users []models.User
    _, err := dbmap.Select(&users, "select * from users order by Id")
    if err != nil { panic(err) }
    c["Users"] = users
    err = templates.Execute(w, c)
    if err != nil { panic(err) }
}

func NewAdminUsers(w http.ResponseWriter, req *http.Request, c Context) {
    switch req.Method {
    case "GET":
      templates.Execute(w, c)
    case "POST":
      //sum validations
      user := models.NewUserFromRequest(dbmap, req)
      http.Redirect(w, req, fmt.Sprintf("/admin/users/%d", user.Id), 301)
    }
}

func ShowAdminUsers(w http.ResponseWriter, req *http.Request, c Context) {
    vars := mux.Vars(req)
    id := vars["id"]
    switch req.Method {
    case "GET":
      var user models.User
      err := dbmap.SelectOne(&user, "select * from users where Id = ?", id)
      if err != nil { panic(err) }
      c["ShowUser"] = user
      var media []models.Media
      _, err = dbmap.Select(&media, "select * from media where user_id = ? order by Id desc", id)
      if err != nil { panic(err) }
      c["Media"] = media
      stats := map[string]int{
        "TotalMedia": len(media),
      }
      c["Stats"] = stats
      templates.Execute(w, c)
    case "DELETE":
        fmt.Println("deleting")
        _, err := dbmap.Exec("delete from users where Id = ?", id)
        if err != nil { panic(err) }
        log.Println(fmt.Sprintf("Deleting user %s", id))
    }
}

func EditAdminUsers(w http.ResponseWriter, req *http.Request, c Context) {
    vars := mux.Vars(req)
    id := vars["id"]
    var user models.User
    err := dbmap.SelectOne(&user, "select * from users where Id = ?", id)
    if err != nil { panic(err) }
    c["Edit"] = user
    switch req.Method {
    case "GET":
      templates["newadminusers.html"].ExecuteTemplate(w, "base", c)
    case "POST":
      user.Username = req.FormValue("username")
      user.Password = models.HashPwd(req.FormValue("password"))
      user.Admin = models.ParseCheckBox(req.FormValue("is-admin"))
      _, err := dbmap.Update(&user)
      if err != nil { panic(err) }
      http.Redirect(w, req, fmt.Sprintf("/admin/users/%s", id), 302)
    }
}


//media serving
func ServeMedia(w http.ResponseWriter, req *http.Request) {
    vars := mux.Vars(req)
    req.ParseForm()
    isDownload := req.FormValue("download")
    obj, err := dbmap.Get(models.Media{}, vars["id"])
    if err != nil { panic(err) }
    if obj != nil {
        if isDownload == "true" {
          w.Header().Set("Content-Type", "application/octet-stream")
          w.Header().Set("Content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", obj.(*models.Media).Title))
        }
        log.Println(fmt.Sprintf("Serving media id %s", vars["id"]))
        http.ServeFile(w, req, obj.(*models.Media).Path)
    } else {
        http.Error(w, "File not found", 404)
    }
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
            session, _ := store.Get(req, APP_NAME)
            session.Values["username"] = username
            session.Values["user_id"] = user.Id
            session.Values["loggedin"] = true
            err := session.Save(req, w)
            //better err handling here yo
            if err != nil { panic(err) }
            log.Println(fmt.Sprintf("User %d - %s logged in", user.Id, username))
            http.Redirect(w, req, "/", 302)
        } else {
            log.Println(err)
            http.Redirect(w, req, "/login", 403)
        }
    }
}

func Logout(w http.ResponseWriter, req *http.Request) {
    session, _ := store.Get(req, APP_NAME)
    username := session.Values["username"]
    user_id := session.Values["user_id"]
    delete(session.Values, "username")
    delete(session.Values, "user_id")
    delete(session.Values, "loggedin")
    session.Options = &sessions.Options{MaxAge: -1}
    err := session.Save(req, w)
    if err != nil { panic(err) }
    log.Println(fmt.Sprintf("User %d - %s logged in", user_id, username))
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
var logFile *os.File
var dbmap *gorp.DbMap
var templates = templates_ago.NewTemplates()
var store *sessions.CookieStore
var templateDir string
var dbName string
var logFileName string
var port string
var mediaDir = "users/"
const STATIC_PATH = "static/"
const APP_NAME = "gopher_pi"

func initConfig() {
    var config map[string]string
    defaults := map[string]string{
      "Port": "3000",
      "Templates": "templates/",
      "DbName": "gopher_pi.db",
      "LogFile": "log.txt",
      "CookieSecret": "secret",
    }
    file, err := os.Open("config.json")
    defer func() {
      if file != nil { file.Close() }
    }()
    if err != nil {
      fmt.Println("Error opening configuration file, using defaults")
      parseConfig(defaults)
      return
    }
    configRead := make([]byte, 4096)
    count, err := file.Read(configRead)
    if err != nil {
      fmt.Println("Error reading configuration file, using defaults")
      parseConfig(defaults)
      return
    }
    err = json.Unmarshal(configRead[:count], &config)
    if err != nil { fmt.Println(err) }
    parseConfig(config)
}

func parseConfig(config map[string]string) {
    store = sessions.NewCookieStore([]byte(config["CookieSecret"]))
    templateDir = config["Templates"]
    dbName = config["DbName"]
    logFileName = config["LogFile"]
    port = fmt.Sprintf(":%s", config["Port"])
}

func setupLogging() {
    logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil { panic(err) }
    log.SetOutput(logFile)
    log.Println("Setting up loggin and initiating server")
}

func setupDatabase() {
    var err error
    db, err := sql.Open("sqlite3", dbName)
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
        _, err := models.NewUser(dbmap, "admin", pwd, true)
        if err != nil {
            fmt.Println("Problem with creating user")
            panic(err)
        }
    }
    log.Println("Database up")
}

func init() {
    //setup database and templates and logging
    initConfig()
    setupLogging()
    setupDatabase()
    templates_ago.LoadTemplates(templateDir, templates)
}

func main() {
    defer dbmap.Db.Close()
    defer logFile.Close()
    router := mux.NewRouter()

    router.HandleFunc("/login", logPanic(Login))
    router.HandleFunc("/logout", logPanic(Logout))

    router.HandleFunc("/", AuthWrapper(IndexMedia))
    router.HandleFunc("/media", AuthWrapper(IndexOwnMedia))
    router.HandleFunc("/media/new", AuthWrapper(NewMedia))
    router.HandleFunc("/media/{id}", AuthWrapper(ShowMedia))
    router.HandleFunc("/media/{id}/edit", AuthWrapper(EditMedia))

    router.HandleFunc("/admin", AdminAuthWrapper(IndexAdmin))
    router.HandleFunc("/admin/users", AdminAuthWrapper(IndexAdminUsers))
    router.HandleFunc("/admin/users/new", AdminAuthWrapper(NewAdminUsers))
    router.HandleFunc("/admin/users/{id}", AdminAuthWrapper(ShowAdminUsers))
    router.HandleFunc("/admin/users/{id}/edit", AdminAuthWrapper(EditAdminUsers))
    //router.HandleFunc("/admin/media/", HandleWrapper(IndexAdminMedia))
    //router.HandleFunc("/admin/media/new", HandleWrapper(NewAdminMedia))

    router.HandleFunc("/about", HandleWrapper(About))

    router.HandleFunc("/serve/{id}", logPanic(ServeMedia))
    //handle all assets below static too
    router.HandleFunc("/static/{asset:[a-zA-Z0-9\\./-]+}", logPanic(StaticHandler))

    log.Println("routes set, about to handle")
    http.Handle("/", router)
    err := http.ListenAndServe(port, nil)
    if err != nil { panic(err) }
}

