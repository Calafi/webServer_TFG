package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "sync/atomic"
    "strings"
)

type parameters struct {
    Body string `json:"body"`
}

type validR struct {
    Cleaned_body string `json:"cleaned_body"`
}


type apiConfig struct {
    fileserverHits atomic.Int32
}

func badWordReplacement(text string) string {
    divided := strings.Split(text, " ")
    for i, s := range divided{
	if strings.ToLower(s) == "kerfuffle" || strings.ToLower(s) == "sharbert" || strings.ToLower(s) == "fornax" {
	    divided[i]="****"
	}
    }
    return strings.Join(divided, " ")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    cfg.fileserverHits.Add(1)
    next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
        fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", hits)
    }

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        cfg.fileserverHits.Store(0)
    }


func (cfg *apiConfig) handlerValidate(w http.ResponseWriter, r *http.Request){
    params := parameters{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&params)
    if err != nil {
	respondWithError(w, 400, "error decoding request JSON")
    }
    if len(params.Body) > 140 {
        respondWithError(w, 400, "Chirp is too long")
    }else{
	cleanBody := badWordReplacement(params.Body)
	resp := validR{Cleaned_body: cleanBody}
	respondWithJSON(w, 200, resp)
    }
}


func respondWithError(w http.ResponseWriter, code int, msg string){
    type errorR struct {
        Error string `json:"error"`
    }
    ret := errorR{Error: msg}
    dat, err := json.Marshal(ret)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
    dat, err := json.Marshal(payload)
    if err != nil { 
        respondWithError(w, 400, "Error marshalling JSON")
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    w.Write(dat)

}

func main(){
    var apiCfg apiConfig
    mux := http.NewServeMux()
    fileServer := http.FileServer(http.Dir("."))
    handler := http.StripPrefix("/app", fileServer)

    mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

    mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
    })

    mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

    mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

    mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidate)

     server := &http.Server{
        Addr: ":8080",
        Handler: mux,
    }

    err := server.ListenAndServe()
    if err != nil {
        fmt.Printf("ERROR: %v\n", err)
    }
}

