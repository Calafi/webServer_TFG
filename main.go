package main

import (
    "net/http"
    "fmt"
    "sync/atomic"
)

type apiConfig struct {
    fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    cfg.fileserverHits.Add(1)
    next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
        fmt.Fprintf(w, "Hits: %d", hits)
    }

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        cfg.fileserverHits.Store(0)
    }


func main(){
    var apiCfg apiConfig
    mux := http.NewServeMux()
    fileServer := http.FileServer(http.Dir("."))
    handler := http.StripPrefix("/app", fileServer)

    mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
    })

    mux.HandleFunc("/metrics", apiCfg.handlerMetrics)

    mux.HandleFunc("/reset", apiCfg.handlerReset)

     server := &http.Server{
        Addr: ":8080",
        Handler: mux,
    }

    err := server.ListenAndServe()
    if err != nil {
        fmt.Printf("ERROR: %v\n", err)
    }
}

