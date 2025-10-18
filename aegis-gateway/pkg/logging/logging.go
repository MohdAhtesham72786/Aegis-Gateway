package logging

import (
    "encoding/json"
    "log"
    "os"
    "sync"

    "gopkg.in/natefinch/lumberjack.v2"
)

var (
    mu sync.Mutex
    rot *lumberjack.Logger
)

func init() {
    // ensure logs dir
    _ = os.MkdirAll("./logs", 0o755)
    rot = &lumberjack.Logger{
        Filename:   "./logs/aegis.log",
        MaxSize:    10, // megabytes
        MaxBackups: 5,
        MaxAge:     28, //days
        Compress:   true,
    }
}

// LogJSON writes the provided map as JSON to stdout and append to rotating file.
func LogJSON(m map[string]interface{}) {
    mu.Lock()
    defer mu.Unlock()
    b, err := json.Marshal(m)
    if err != nil {
        log.Printf("log marshal error: %v", err)
        return
    }
    s := string(b)
    // stdout
    log.Println(s)
    if rot != nil {
        if _, err := rot.Write(append(b, '\n')); err != nil {
            log.Printf("write log file error: %v", err)
        }
    }
}
