package web

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	ID    string `json:"id"`
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
	r.HandleFunc("/api/urls/{id}", s.handlePreflight).Methods("OPTIONS")
	// r.HandleFunc("/{id}", s.redirect).Methods("GET")
	// r.HandleFunc("/{id}", s.handlePreflight).Methods("OPTIONS")
	r.HandleFunc("/api/urls", s.addURL).Methods("POST")
	r.HandleFunc("/api/urls", s.handlePreflight).Methods("OPTIONS")
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

func (s *urlServer) handlePreflight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

func (s *urlServer) getURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	id := mux.Vars(r)["id"]
	_, url, err := s.dbWorker.Find("id_to_url", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for id %v: %v", id, err)
		return
	}

	if url != "" {
		w.Header().Set("location", url)
		w.WriteHeader(http.StatusPermanentRedirect)
	} else {
		idErr := urlPayload{
			ID:    id,
			Error: fmt.Sprintf("ID %v does not exists", id),
		}

		idErrJSON, err := json.Marshal(idErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(idErrJSON)
	}
}

// func (s *urlServer) redirect(w http.ResponseWriter, r *http.Request) {
// 	id := mux.Vars(r)["id"]
// 	http.Redirect(w, r, fmt.Sprintf("http://%v:%v/api/urls/%v", s.host, s.port, id), http.StatusPermanentRedirect)
// }

func (s *urlServer) addURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
	w.Header().Set("Access-Control-Expose-Headers", "Location")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while reading data: %v", err)
		return
	}

	var b urlPayload
	err = json.Unmarshal(body, &b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Bad JSON format: %v", err)
		return
	}

	_, err = url.ParseRequestURI(b.URL)
	if err != nil {
		urlErr := urlPayload{
			ID:    b.ID,
			URL:   "",
			Error: fmt.Sprintf("Invalid url: %v", b.URL),
		}

		urlErrJSON, err := json.Marshal(urlErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(urlErrJSON)
		return
	}

	id, _, err := s.dbWorker.Find("url_to_id", b.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for url %v: %v", b.URL, err)
		return
	}

	if id != "" {
		idErr := urlPayload{
			ID:    id,
			URL:   b.URL,
			Error: fmt.Sprintf("Url %v already registered under id %v", b.URL, id),
		}

		idErrJSON, err := json.Marshal(idErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(idErrJSON)
		return
	}

	if b.ID == "" {
		hasher := md5.New()
		hasher.Write([]byte(b.URL))
		hashed := hex.EncodeToString(hasher.Sum(nil))
		b.ID = hashed[len(hashed)-6:]
	}

	_, url, err := s.dbWorker.Find("id_to_url", b.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for id %v: %v", b.ID, err)
		return
	}

	if url != "" {
		idErr := urlPayload{
			ID:    b.ID,
			URL:   url,
			Error: fmt.Sprintf("ID %v already registered for url %v", b.ID, url),
		}

		idErrJSON, err := json.Marshal(idErr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Bad JSON format: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(idErrJSON)
		return
	}

	err = s.dbWorker.Register(b.ID, b.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while registering id %v for url %v: %v", b.ID, b.URL, err)
		return
	}

	w.Header().Set("location", fmt.Sprintf("/%v", b.ID))
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
