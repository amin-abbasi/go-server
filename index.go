package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, this is your Go server!")
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Server is running on http://localhost:4000")
    http.ListenAndServe(":4000", nil)
}
