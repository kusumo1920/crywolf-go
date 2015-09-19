package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/honeybadger-io/honeybadger-go"
)

const DefaultPort = "4000"

func main() {
	APIKey := os.Getenv("HONEYBADGER_API_KEY")
	if APIKey == "" {
		fmt.Println("Error: API key is required.")
		return
	}
	honeybadger.Configure(honeybadger.Configuration{APIKey: APIKey, Root: getRoot()})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Looking to <a href=\"/fail\">/fail</a> or <a href=\"/timeout\">/timeout</a>?\n"))
	})

	http.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		honeybadger.SetContext(honeybadger.Context{"userID": "1"})
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			honeybadger.SetContext(honeybadger.Context{"herokuRequestID": requestID})
		}
		panic(fmt.Errorf("This is an error generated by the crywolf app at %s.", time.Now()))
	})

	http.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Second)
	})

	port := getPort()
	fmt.Printf("Listening on port %s.\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), errorHandler(honeybadger.Handler(nil))))
}

func getPort() (port string) {
	port = os.Getenv("PORT")
	if port == "" {
		return DefaultPort
	}
	return
}

func getRoot() string {
	_, root, _, _ := runtime.Caller(1)
	return path.Dir(root)
}

func errorHandler(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
