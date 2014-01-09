package models

import (
    "log"
    "fmt"
    "net/http"
    "crypto/sha512"
    "os"
    "io"
    "strconv"
    "github.com/coopernurse/gorp"
)

type User struct {
    Id uint64
    Username, Password string
    Admin bool
}

type Media struct {
    Id, User_id uint64
    Title, Type, Path string
    Private bool
}

func NewUser(dbmap *gorp.DbMap, uname, pword string, admin bool) (*User, error) {
    new_user := &User{Username: uname, Password: pword, Admin: admin}
    err := dbmap.Insert(new_user)
    if err != nil {
        log.Println("Error creating user")
        return nil, err
    }
    return new_user, nil
}

func NewMedia(dbmap *gorp.DbMap, uid uint64, title, m_type, path string, priv bool) (*Media, error) {
    new_media := &Media{Title: title, Type: m_type, Path: path,
                        Private: priv, User_id: uid}
    err := dbmap.Insert(new_media)
    if err != nil {
        log.Println("Error creating media")
        return nil, err
    }
    return new_media, nil
}

func NewMediaFromRequest(dbmap *gorp.DbMap, req *http.Request, user_id string) *Media {
      log.Println(user_id)
      file, header, err := req.FormFile("file")
      if err != nil { panic(err) }
      f_path := fmt.Sprintf("users/%s/video/%s", user_id, header.Filename)
      f, err := os.Create(f_path)
      if err != nil { panic(err) }
      _, err = io.Copy(f, file)
      if err != nil { panic(err) }
      f.Close()

      title := req.FormValue("title")
      private := false
      m_type := "video"
      uid, err := strconv.ParseUint(fmt.Sprintf("%d", user_id), 10, 64)
      media, err := NewMedia(dbmap, uid, title, m_type, f_path, private)
      if err != nil { panic(err) }
      return media
}

func NewUserFromRequest(dbmap *gorp.DbMap, req *http.Request) *User {
      uname := req.FormValue("username")
      pword := HashPwd(req.FormValue("password"))
      log.Println(req.FormValue("is-admin"))
      admin := ParseCheckBox(req.FormValue("is-admin"))
      user, err := NewUser(dbmap, uname, pword, admin)
      if err != nil { panic(err) }
      //create user dir. tidy this up
      err = os.Mkdir(fmt.Sprintf("users/%d/", user.Id), 0755)
      if err != nil { panic(err) }
      err = os.Mkdir(fmt.Sprintf("users/%d/video/", user.Id), 0755)
      if err != nil { panic(err) }
      err = os.Mkdir(fmt.Sprintf("users/%d/audio/", user.Id), 0755)
      if err != nil { panic(err) }
      return user
}

func HashPwd(password string) string {
    h := sha512.New()
    fmt.Fprint(h, password)
    return fmt.Sprintf("%x", h.Sum(nil))
}

func ParseCheckBox(box string) bool {
  switch box {
  case "on":
    return true
  default:
    return false
  }
}
