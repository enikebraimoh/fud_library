package main

import (
	"encoding/json"
	"fmt"
	"fud_library/auth"
	"fud_library/book"
	"fud_library/utils"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func Router(server *socketio.Server) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/", VersionHandler)
	r.HandleFunc("/posts", book.CreatePost).Methods("POST")
	r.HandleFunc("/posts", book.GetPosts).Methods("GET")
	r.HandleFunc("/sendotp", auth.SendOTP).Methods("POST")
	r.HandleFunc("/verifyotp", auth.VerifyOTP).Methods("POST")

	return r
}

func main() {

	// Socket  events
	var Server = socketio.NewServer(nil)

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	fmt.Println("Environment variables successfully loaded. Starting application...")

	if err = utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		fmt.Println("Could not connect to MongoDB")
	}

	// get PORT from environment variables
	port, _ := os.LookupEnv("PORT")
	if port == "" {
		port = "8000"
	}

	r := Router(Server)

	c := cors.AllowAll()

	h := RequestDurationMiddleware(r)

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, c.Handler(h)),
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//nolint:errcheck //CODEI8: ignore error check
	go Server.Serve()

	fmt.Println("Socket Served")

	defer Server.Close()

	fmt.Println("Fud Book API - running on port ", port)
	//nolint:gocritic //CODEI8: please provide soln -> lint throw exitAfterDefer: log.Fatal will exit, and `defer Server.Close()` will not run
	log.Fatal(srv.ListenAndServe())

}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Fud Book API - Version 0.01\n")
}

func RequestDurationMiddleware(h http.Handler) http.Handler {
	const durationLimit = 10

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)

		postToSlack := func() {
			m := make(map[string]interface{})
			m["timeTaken"] = duration.Seconds()

			if duration.Seconds() < durationLimit {
				return
			}

			scheme := "http"

			if r.TLS != nil {
				scheme += "s"
			}

			m["endpoint"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)
			m["timeTaken"] = duration.Seconds()

			b, _ := json.Marshal(m)
			resp, err := http.Post("https://companyfiles.zuri.chat/api/v1/slack/message", "application/json", strings.NewReader(string(b)))

			if err != nil {
				return
			}

			if resp.StatusCode != 200 {
				fmt.Printf("got error %d", resp.StatusCode)
			}

			defer resp.Body.Close()
		}

		if strings.Contains(r.Host, "api.zuri.chat") {
			go postToSlack()
		}
	})
}
