package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/georgiv/url-shortener/server/db"
	"github.com/gorilla/mux"
)

type WebWorker interface {
	Handle()
	Shutdown()
}

type urlServer struct {
	host      string
	port      int
	dbWorker  db.DbWorker
	webWorker http.Server
}

type urlAlias struct {
	Original string `json:"original"`
	Custom   string `json:"custom"`
}

func NewServer(host string, port int) (WebWorker, error) {
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
	s.webWorker = http.Server{Addr: fmt.Sprintf("%v:%v", s.host, s.port), Handler: r}

	go s.stopListener()

	err := s.webWorker.ListenAndServe()
	if err != nil {
		log.Println(fmt.Sprintf("Error while processing requests: %v", err))
		s.Shutdown()
	}
}

func (s *urlServer) Shutdown() {
	log.Println("Shutting down web server...")

	s.dbWorker.Shutdown()

	log.Println("Web server successfully shut down")
}

func (s *urlServer) getURL(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	url, err := s.dbWorker.Find("id_to_url", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Panicf("Error while retrieving data for key %v", id)
	}

	if url != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, url)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, fmt.Sprintf("Key %v does not exist", id))
	}
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

func (s *urlServer) stopListener() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	signalType := <-ch
	signal.Stop(ch)

	log.Println(fmt.Sprintf("Received signal: %v", signalType))

	s.Shutdown()

	os.Exit(0)
}
