package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func Index(w http.ResponseWriter, r *http.Request, reqs chan Request) {
    fmt.Fprintln(w, "Confirmation Service for Ocean Attestations!")
}

func Block(w http.ResponseWriter, r *http.Request, reqs chan Request) {
    blockid := mux.Vars(r)["blockId"]
    request := Request{Name: mux.CurrentRoute(r).GetName(), Id: blockid,}

    reqs <- request

    response := <- reqs
    fmt.Fprintln(w, response.Attested)

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}

func BestBlock(w http.ResponseWriter, r *http.Request, reqs chan Request) {
    request := Request{Name: mux.CurrentRoute(r).GetName(),}

    reqs <- request

    response := <- reqs
    fmt.Fprintln(w, response.Attested)

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}