// Package web provides interface for managing web
// server, which exposes REST API for registering
// and accessing URL aliases.
//
// Copyright 2019 cranki. All rights reserved.
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
	"unicode"

	"github.com/georgiv/url-shortener/server/db"
	"github.com/gorilla/mux"
)

// Server exports API for starting and stopping web
// server, which exposes REST API for registering
// and accessing URL aliases.
type Server interface {
	// Exposes REST endpoints:
	//   - /api/urls/{id}: supports GET and OPTIONS methods.
	//     In case of existing id a permanent redirect (308) is
	//     sent to the client. In case of non-existing id, not
	//     found error (404) is sent to the client as JSON payload.
	//     In case of error, only status 500 is sent to the client
	//     and the error is logged
	//   - /api/urls: supports POST and OPTIONS methods. In case
	//     of successful registration, the client receives reply for
	//     created resource (201). In case of invalid payload or
	//     invalid url in the payload, a bad request (400) error
	//     is being sent. In case there is already existing entry
	//     with the same id or url, a conflict error (409) is sent
	//     to the client
	//     The incoming payload should be JSON containing id (optional)
	//     and url (required). In case of missing id, the server will
	//     generate one automatically consisting of 6 symbols
	//  All endpoints support CORS requests.
	//  The outgoing payload is JSON containing id, url and error.
	Handle()

	// Shuts down the underlying DB worker and perform all
	// necessary cleanups of resources. In case of an error,
	// it is only logged properly, but not returned.
	Shutdown()
}

// NewServer creates and returns instance satisfying the Server
// interface
// Params:
//   - host: binding host for the URL shortener service
//   - port: listening port for the URL shortener service
//   - expiration: integer representing the period in days for
//     running cleanup service which removes the entries which
//     expired
func NewServer(host string, port int, expiration int) (server Server, err error) {
	dbWorker, err := db.NewWorker(expiration)
	if err != nil {
		return
	}

	server = &web{host: host, port: port, dbWorker: dbWorker}
	return
}

type web struct {
	host      string
	port      int
	dbWorker  db.Worker
	webWorker http.Server
}

type payload struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Error string `json:"error"`
}

func (server *web) Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/api/urls/{id}", server.getURL).Methods("GET")
	r.HandleFunc("/api/urls/{id}", server.handlePreflight).Methods("OPTIONS")
	r.HandleFunc("/api/urls", server.addURL).Methods("POST")
	r.HandleFunc("/api/urls", server.handlePreflight).Methods("OPTIONS")
	server.webWorker = http.Server{Addr: fmt.Sprintf("%v:%v", server.host, server.port), Handler: r}

	go server.stopListener()

	log.Println(fmt.Sprintf("Server accepts requests on port %v", server.port))

	err := server.webWorker.ListenAndServe()
	if err != nil {
		log.Println(fmt.Sprintf("Error while processing requests: %v", err))
		server.Shutdown()
	}
}

func (server *web) Shutdown() {
	log.Println("Shutting down web server...")

	server.dbWorker.Shutdown()

	log.Println("Web server successfully shut down")
}

func (server *web) handlePreflight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}

func (server *web) getURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	id := mux.Vars(r)["id"]
	_, url, err := server.dbWorker.Find("id_to_url", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for id %v: %v", id, err)
		return
	}

	if url != "" {
		w.Header().Set("location", url)
		w.WriteHeader(http.StatusPermanentRedirect)
	} else {
		idErr := payload{
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

func (server *web) addURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
	w.Header().Set("Access-Control-Expose-Headers", "Location")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while reading data: %v", err)
		return
	}

	var b payload
	err = json.Unmarshal(body, &b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Bad JSON format: %v", err)
		return
	}

	_, err = url.ParseRequestURI(b.URL)
	if err != nil {
		urlErr := payload{
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

	if len(b.ID) != 0 && len(b.ID) != 6 {
		urlErr := payload{
			ID:    b.ID,
			URL:   b.URL,
			Error: fmt.Sprintf("Invalid ID length: %v is %v character long. It should be exactly 6 characters long", b.ID, len(b.ID)),
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

	for _, s := range b.ID {
		if !(unicode.IsLetter(s) || (s >= '0' && s <= '9') || s == '_' || s == '-') {
			urlErr := payload{
				ID:    b.ID,
				URL:   b.URL,
				Error: fmt.Sprintf("ID contains forbidden characters: %v. Allowed characters: alphanumeric characters, underscore and dash", b.ID),
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
	}

	id, _, err := server.dbWorker.Find("url_to_id", b.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for url %v: %v", b.URL, err)
		return
	}

	if id != "" {
		idErr := payload{
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

	_, url, err := server.dbWorker.Find("id_to_url", b.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while retrieving data for id %v: %v", b.ID, err)
		return
	}

	if url != "" {
		idErr := payload{
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

	err = server.dbWorker.Register(b.ID, b.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while registering id %v for url %v: %v", b.ID, b.URL, err)
		return
	}

	w.Header().Set("location", fmt.Sprintf("/%v", b.ID))
	w.WriteHeader(http.StatusCreated)
}

func (server *web) stopListener() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	signalType := <-ch
	signal.Stop(ch)

	log.Println(fmt.Sprintf("Received signal: %v", signalType))

	server.Shutdown()

	os.Exit(0)
}
