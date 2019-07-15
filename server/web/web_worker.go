package web

import (
	"crypto/md5"
	"encoding/hex"
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

type urlPayload struct {
	Key   string `json:"key"`
	URL   string `json:"url"`
	Error string `json:"error"`
}

func NewServer(host string, port int, expiration int) (WebWorker, error) {
	dbWorker, err := db.NewDbWorker(expiration)
	if err != nil {
		return nil, err
	}

	return &urlServer{host: host, port: port, dbWorker: dbWorker}, nil
}

func (s *urlServer) Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/api/urls/{id}", s.getURL).Methods("GET")
	r.HandleFunc("/{id}", s.redirect).Methods("GET")
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
		log.Printf("Error while retrieving data for key %v: %v", id, err)
		return
	}

	if url != "" {
		w.Header().Set("location", url)
		w.WriteHeader(http.StatusPermanentRedirect)
	} else {
		keyErr := urlPayload{
			Key:   id,
			Error: fmt.Sprintf("Key %v does not exists", id),
		}

		keyErrJSON, err := json.Marshal(keyErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(keyErrJSON)
	}
}

func (s *urlServer) redirect(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	http.Redirect(w, r, fmt.Sprintf("http://%v:%v/api/urls/%v", s.host, s.port, id), http.StatusPermanentRedirect)
}

func (s *urlServer) addURL(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while reading data: %v", err)
		return
	}

	var b urlPayload
	err = json.Unmarshal(body, &b)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, fmt.Sprintf("Bad JSON format: %v", err))
		log.Printf("Bad JSON format: %v", err)
		return
	}

	if b.Key == "" {
		hasher := md5.New()
		hasher.Write([]byte(b.URL))
		hashed := hex.EncodeToString(hasher.Sum(nil))
		b.Key = hashed[len(hashed)-6:]
	}

	url, err := s.dbWorker.Find("id_to_url", b.Key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for key %v: %v", b.Key, err)
		return
	}

	if url != "" {
		keyErr := urlPayload{
			Key:   b.Key,
			URL:   url,
			Error: fmt.Sprintf("Key %v already registered for url %v", b.Key, url),
		}

		keyErrJSON, err := json.Marshal(keyErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(keyErrJSON)
		return
	}

	id, err := s.dbWorker.Find("url_to_id", b.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for url %v: %v", b.URL, err)
		return
	}

	if id != "" {
		keyErr := urlPayload{
			Key:   id,
			URL:   b.URL,
			Error: fmt.Sprintf("Url %v already registered under key %v", b.URL, id),
		}

		keyErrJSON, err := json.Marshal(keyErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(keyErrJSON)
		return
	}

	err = s.dbWorker.Register(b.Key, b.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while registering key %v for url %v: %v", b.Key, b.URL, err)
		return
	}

	w.Header().Set("location", fmt.Sprintf("http://%v:%v/%v", s.host, s.port, b.Key))
	w.WriteHeader(http.StatusCreated)
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
