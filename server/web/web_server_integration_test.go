package web_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/georgiv/url-shortener/server/web"
	"github.com/georgiv/url-shortener/testdata"
)

func TestNewServer(t *testing.T) {
	test := func() {
		_, err := web.NewServer("localhost", 8888, 7)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	testdata.Execute(t, test)
}

func TestNewServerMissingConfig(t *testing.T) {
	test := func() {
		err := os.Chdir("./res")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		server, err := web.NewServer("localhost", 8888, 7)

		if server != nil {
			t.Errorf("Expected nil, received %v", server)
		}

		if err == nil {
			t.Errorf("Expected error, received nil")
		}
	}

	testdata.Execute(t, test)
}

func TestHandleGet(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		id := "cranki"
		url := "https://google.com"

		testdata.AddEntry(t, id, url, 604800)

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			client := http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			resp, err := client.Get("http://localhost:8888/api/urls/cranki")
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 308 {
				t.Errorf("Expected status code 308, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandleGetNonExistingEntry(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			client := http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			resp, err := client.Get("http://localhost:8888/api/urls/cranki")
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 404 {
				t.Errorf("Expected status code 404, received: %v", status)
			}
		}()

		server.Handle()
	}

	testdata.Execute(t, test)
}

func TestHandlePost(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"id":  "cranki",
				"url": "http://testurl.com",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 201 {
				t.Errorf("Expected status code 201, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostNoId(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"url": "http://testurl.com",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 201 {
				t.Errorf("Expected status code 201, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostNoUrl(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"id": "cranki",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 400 {
				t.Errorf("Expected status code 400, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostBadUrl(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"url": "testurl",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 400 {
				t.Errorf("Expected status code 400, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostNoBody(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				nil)
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 500 {
				t.Errorf("Expected status code 500, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostBadJson(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer([]byte("bad json")))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 500 {
				t.Errorf("Expected status code 500, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostBadIdSize(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"id":  "crank",
				"url": "http://testurl.com",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 400 {
				t.Errorf("Expected status code 400, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostIdWithForbiddenCharacters(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"id":  "crank ",
				"url": "http://testurl.com",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))
			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 400 {
				t.Errorf("Expected status code 400, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostConflictId(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"id":  "cranki",
				"url": "http://testurl.com",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))

			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 201 {
				t.Errorf("Expected status code 201, received: %v", status)
			}

			jsonBody, err = json.Marshal(map[string]string{
				"id":  "cranki",
				"url": "http://anothertesturl.com",
			})

			resp, err = http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))

			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status = resp.StatusCode
			if status != 409 {
				t.Errorf("Expected status code 409, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestHandlePostConflictUrl(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(3 * time.Second)

			jsonBody, err := json.Marshal(map[string]string{
				"id":  "cranki",
				"url": "http://testurl.com",
			})

			resp, err := http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))

			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status := resp.StatusCode
			if status != 201 {
				t.Errorf("Expected status code 201, received: %v", status)
			}

			jsonBody, err = json.Marshal(map[string]string{
				"id":  "tester",
				"url": "http://testurl.com",
			})

			resp, err = http.Post("http://localhost:8888/api/urls",
				"application/json",
				bytes.NewBuffer(jsonBody))

			if resp == nil {
				t.Errorf("Expected response, received nil")
			}
			if err != nil {
				t.Errorf("Unexpected nil, received: %v", err)
			}

			status = resp.StatusCode
			if status != 409 {
				t.Errorf("Expected status code 409, received: %v", status)
			}
		}()

		server.Handle()

		testdata.DeleteAll(t)
	}

	testdata.Execute(t, test)
}

func TestShutdown(t *testing.T) {
	test := func() {
		server, err := web.NewServer("localhost", 8888, 7)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		go func() {
			time.Sleep(5 * time.Second)

			server.Shutdown()
		}()

		go func() {
			time.Sleep(10 * time.Second)

			resp, err := http.Get("http://localhost:8888/api/urls/cranki")
			if resp != nil {
				t.Errorf("Expected nil, received: %v", resp)
			}
			if err == nil {
				t.Errorf("Unexpected error, received nil")
			}
		}()

		server.Handle()
	}

	testdata.Execute(t, test)
}
