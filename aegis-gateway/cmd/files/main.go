package main

import (
    "log"
    "os"
    "strconv"

    "github.com/example/aegis-gateway/internal/adapters/files"
)

func main() {
    port := 9002
    if p := os.Getenv("PORT"); p != "" {
        if v, err := strconv.Atoi(p); err == nil { port = v }
    }
    srv := files.NewServer()
    log.Printf("files service starting on :%d", port)
    srv.Start(port)
}
