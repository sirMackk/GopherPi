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

type FileMatchingFunc func(filename string) string


func RescanMedia(dbmap *gorp.DbMap, uid uint64, priv bool) {
    vids := make(chan string, 100)
    dirs := make(chan string, 100)

    regex, _ := regexp.Compile("^\\.")

    go inspectDirectory(vids, dirs, regex)
    dirs <- "."
    go processFiles(dbmap, vids, uid, priv)

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

func processFiles(dbmap *gorp.DbMap, vids chan string, uid uint64, priv bool) {
    var path string
    fileMatcher := determineFileType()
    fileTitler := determineFileName()
    for {
        path = <- vids
        m_type := fileMatcher(path)
        title := fileTitler(path)
        _, err := models.NewMedia(dbmap, uid, title, m_type, path, priv)
        if err != nil { log.Println(err) }
    }
}

func determineFileType() FileMatchingFunc {
    matchVideo, _ := regexp.Compile("\\.(avi|mp4|mkv|flv)$")
    matchAudio, _ := regexp.Compile("\\.(mp3|ogg|aac|flac|m4a|wav|wma)")
    return func(filename string) string {
        if matchVideo.MatchString(filename) {
            return "video"
        } else if matchAudio.MatchString(filename) {
            return "audio"
        } else {
            return "unknown"
        }
    }
}

func determineFileName() FileMatchingFunc {
    extractName, _ := regexp.Compile("/[\\S\\s]+\\.[avimp4kfl3ogcw]{2,4}$")
    return func(filename string) string {
        name := extractName.FindString(filename)
        if len(name) == 0 {
            return "unknown"
        } else {
            return name
        }
    }
}

