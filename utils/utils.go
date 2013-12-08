package utils

import (
    "io/ioutil"
    "../models"
    "os"
    "fmt"
    "log"
    "sync"
    "regexp"
    "github.com/coopernurse/gorp"
    "strconv"
)

type FileMatchingFunc func(filename string) string


func ScanMediaDir(dbmap *gorp.DbMap, directory, uid, priv string) {
    vids := make(chan string, 100)
    dirs := make(chan string, 100)
    wg := new(sync.WaitGroup)
    user_id := toUi64(uid)
    priv_setting, err := strconv.ParseBool(priv)
    if err != nil {
        priv_setting = true
        log.Println("Error converting to boolean in utils ScanMediaDir")
    }

    regex, _ := regexp.Compile("^\\.")

    wg.Add(1)
    go inspectDirectory(vids, dirs, regex, wg)
    //dirs <- "."
    dirs <- directory
    go processFiles(dbmap, vids, user_id, priv_setting)
    wg.Wait()
    //time.Sleep(10 * 1e9)
}

func PruneMedia(dbmap *gorp.DbMap) error {
    //used when using batch processing ie. over 10k records
    //ids := countRecords(dbmap)
    if err := simplePrune(dbmap); err != nil {
        return err
    } else {
        return nil
    }
}

func inspectDirectory(vids, dirs chan string, reg *regexp.Regexp, wg *sync.WaitGroup) {
    var currentDir string
    defer wg.Done()
    currentDir = <- dirs

    files, err := ioutil.ReadDir(currentDir)
    if err != nil { log.Println(err) }
    for _, file := range files {
        fileName := file.Name()
        if !reg.MatchString(fileName) {
            if file.IsDir() {
                dirs <- buildPath(currentDir, fileName)
                wg.Add(1)
                go inspectDirectory(vids, dirs, reg, wg)
            } else {
                vids <- buildPath(currentDir, fileName)
            }
        }
    }
}

func buildPath(dir, file string) string {
    return fmt.Sprintf("%s%s", dir, file)
}

func processFiles(dbmap *gorp.DbMap, vids chan string, uid uint64, priv bool) {
    var path string
    fileMatcher := determineFileType()
    fileTitler := determineFileName()
    for {
        path = <- vids
        m_type := fileMatcher(path)
        if m_type == "video" || m_type == "audio" {
          title := fileTitler(path)
          _, err := models.NewMedia(dbmap, uid, title, m_type, path, priv)
          if err != nil { log.Println(err) }
        }
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
    extractName, _ := regexp.Compile("[\\)\\(\\w\\s-\\.]+\\.[m4flcwogkvp3ai]{2,4}$")
    return func(filename string) string {
        name := extractName.FindString(filename)
        if len(name) == 0 {
              return "unknown"
        } else {
              return name
        }
    }
}

func toUi64(integer string) uint64 {
    val, err := strconv.ParseUint(integer, 10, 64)
    if err != nil {
        log.Println("Error in toUi64")
        return 0
    }
    return val
}

func countRecords(dbmap *gorp.DbMap) []uint64 {
    var ids []uint64
    _, err := dbmap.Select(&ids, "select Id from media")
    //take care of errs
    if err != nil { panic(err) }
    return ids
}

func simplePrune(dbmap *gorp.DbMap) error {
    var media []models.Media
    _, err := dbmap.Select(&media, "select * from media")
    if err != nil { panic(err) }
    for _, mediaItem := range media {
        if _, err := os.Stat(mediaItem.Path); os.IsNotExist(err) {
            //check error here
            _, err := dbmap.Delete(&mediaItem)
            if err != nil {
                log.Println(err)
                return err
            }
        }
    }
    return nil
}
