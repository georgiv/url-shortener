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
	host     string
	port     int
	dbWorker db.DbWorker
}

type urlAlias struct {
	Original string `json:"original"`
	Custom   string `json:"custom"`
}

func NewServer(host string, port int) (Server, error) {
	dbWorker, err := db.NewDbWorker()
	if err != nil {
		return nil, err
	}

	return &urlServer{host: host, port: port, dbWorker: dbWorker}, nil
}

func (s *urlServer) Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/api/urls/{id}", s.getURL).Methods("GET")
	r.HandleFunc("/api/urls", s.addURL).Methods("POST")
	http.ListenAndServe(fmt.Sprintf("%v:%v", s.host, s.port), r)
}

func (s *urlServer) getURL(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get URL %v", mux.Vars(r)["id"])
}

func (s *urlServer) addURL(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var b urlAlias
	json.Unmarshal(body, &b)
	fmt.Println(b)
	//w.Header().Set("Location", "http://localost:8888/dummy")
	fmt.Fprintf(w, "Add URL")
}
