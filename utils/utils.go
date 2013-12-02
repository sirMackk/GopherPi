package utils

import (
    "io/ioutil"
    "../models"
    "fmt"
    "log"
    "regexp"
    "time"
    "github.com/coopernurse/gorp"
)

type ProcFunc func(path string) error

func RescanMedia(dbmap *gorp.DbMap, processType string) {
    vids := make(chan string, 100)
    dirs := make(chan string, 100)

    regex, _ := regexp.Compile("^\\.")

    switch processType {
      //type depends on func, add regexp file extension matchers,
      //extract title, pass in user id. Maybe make this process both 
      //audio and video at the same time?
    case "video":
        processFn := func(path string) error {
            _, err := models.NewMedia(dbmap, uid, title, m_type, priv)
            if err != nil { return err }
            return nil
        }
    case "audio":
        processFn := func(path string) error {
            _, err := models.NewMedia(dbmap, uid, title, m_type, priv)
            if err != nil { return err }
            return nil
        }
    }


    go inspectDirectory(vids, dirs, regex)
    dirs <- "."
    go processFiles(dbmap, vids, processFn)

    time.Sleep(10 * 1e9)
}

func inspectDirectory(vids, dirs chan string, reg *regexp.Regexp) {
    var currentDir string
    currentDir = <- dirs
    files, err := ioutil.ReadDir(currentDir)
    if err != nil { log.Println(err) }
    for _, file := range files {
        fileName := file.Name()
        if !reg.MatchString(fileName) {
            if file.IsDir() {
                dirs <- buildPath(currentDir, fileName)
                go inspectDirectory(vids, dirs, reg)
            } else {
                vids <- buildPath(currentDir, fileName)
            }
        }
    }
}

func buildPath(dir, file string) string {
    return fmt.Sprintf("%s/%s", dir, file)
}

func processFiles(dbmap *gorp.DbMap, vids chan string, proc ProcFunc) {
    var path string
    for {
        path = <- vids
        err := proc(path)
        //_, err := models.NewMedia(dbmap, uid, title, m_type, priv)
        //if err != nil { log.Println(err) }
    }
}

