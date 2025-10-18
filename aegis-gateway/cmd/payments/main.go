package main

import (
    "log"
    "os"
    "strconv"

    "github.com/example/aegis-gateway/internal/adapters/payments"
)

func main() {
    port := 9001
    if p := os.Getenv("PORT"); p != "" {
        if v, err := strconv.Atoi(p); err == nil { port = v }
    }
    srv := payments.NewServer()
    log.Printf("payments service starting on :%d", port)
    srv.Start(port)
}
