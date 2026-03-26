package main

import (
    "net/http"
    "fmt"
)

func main(){
    mux := http.NewServeMux()
    fileServer := http.FileServer(http.Dir("."))
    mux.Handle("/app/", http.StripPrefix("/app", fileServer))
    server := &http.Server{
        Addr: ":8080",
	Handler: mux,
    }

    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
    })

    err := server.ListenAndServe()
    if err != nil {
        fmt.Printf("ERROR: %v\n", err)
    }
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
    // your code here
}
