package web

import (
	"github.com/georgiv/url-shortener/server/db"
	"github.com/gorilla/mux"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Server interface {
	Handle()	
}

type urlServer struct {
	host string
	port int
	db db.DbWorker
}

type urlAlias struct {
	original string
	custom string
}

func NewServer(host string, port int) Server {
	return &urlServer{host: host, port: port}
}

func (s *urlServer) Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/api/urls/{id}", s.getUrl).Methods("GET")
	r.HandleFunc("/api/urls", s.addUrl).Methods("POST")
	http.ListenAndServe(fmt.Sprintf("%v:%v", s.host, s.port), r)
}

func (s *urlServer) getUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get URL %v", mux.Vars(r)["id"])
}

func (s *urlServer) addUrl(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        panic(err)
	}

	var b urlAlias
	json.Unmarshal(body, &b)
	fmt.Println(b)
	w.Header().Set("Location", "http://localost:8888/dummy")
	fmt.Fprintf(w, "Add URL")
}
