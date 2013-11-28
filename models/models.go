package models

import (
    "log"
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

func NewUser(dbmap *gorp.DbMap, uname, pword string) (*User, error) {
    new_user := &User{Username: uname, Password: pword, Admin: false}
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


