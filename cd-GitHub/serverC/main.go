package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type key int

const (
	requestIDKey key    = 0
	host         string = "0.0.0.0"
	port         string = ":15000"
)

var (
	listenAddr string = host + port
	healthy    int32
)

var integers []int

type ScheduleType int

const (
	direct ScheduleType = iota + 1
	response
)

func (s ScheduleType) String() string {
	switch s {
	case direct:
		return "direct"
	case response:
		return "response"
	default:
		return "invalid"
	}
}

// Value struct
type Value struct {
	Timestamp   string `json:"timestamp"`
	ServiceName string `json:"serviceName"`
	Value       int    `json:"value"`
}

type GlobalVarManager struct {
	mu     sync.RWMutex // protects the fields below
	values []Value
}

func NewGlobalVarManager() *GlobalVarManager {
	return &GlobalVarManager{
		values: make([]Value, 0),
	}
}

// postCall handles the /post route
func (sm *GlobalVarManager) postCall(w http.ResponseWriter, r *http.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		value := Value{}
		err = json.Unmarshal(body, &value)
		if err != nil {
			http.Error(w, "JSON unmarshal error", http.StatusInternalServerError)
		}
		fmt.Printf("received value %v\n", value)
		value.Value = value.Value + 100
		t := time.Now()

		value.Timestamp = t.Format(time.RFC3339)
		sm.values = append(sm.values, value)

		intVar, _ := strconv.Atoi(string(body[:]))

		integers = append(integers, intVar+100)

		fmt.Fprint(w, "POST done")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// getCall handles the /get route
func (sm *GlobalVarManager) getCall(w http.ResponseWriter, r *http.Request) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	jsonVal, err := json.Marshal(sm.values)
	if err != nil {
		http.Error(w, "Error converting results to json",
			http.StatusInternalServerError)
	}

	_, err = w.Write(jsonVal)
	if err != nil {
		http.Error(w, "Error sending response body", http.StatusInternalServerError)
	}
}

func main() {
	flag.StringVar(&listenAddr, "listen-addr", port, "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	logger.Println("Server is starting...")

	gm := NewGlobalVarManager()

	router := http.NewServeMux()
	router.Handle("/", index())
	router.HandleFunc("/post", gm.postCall)
	router.HandleFunc("/get", gm.getCall)

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         host + port,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

// index handles the / route
func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "This is Server C!")
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
