package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sdf1979/go_final_project/api"
	"github.com/sdf1979/go_final_project/db"
	"github.com/sdf1979/go_final_project/signin"
)

func getPort() int {
	port := 7540
	envPort := os.Getenv("TODO_PORT")
	if len(envPort) > 0 {
		if eport, err := strconv.ParseInt(envPort, 10, 32); err == nil {
			port = int(eport)
		}
	}
	return port
}

func getFileDb() string {
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	envDbFile := os.Getenv("TODO_DBFILE")
	if len(envDbFile) > 0 {
		dbFile = envDbFile
	}
	return dbFile
}

func main() {

	dbFile := getFileDb()
	store, err := db.Open(dbFile)
	if err != nil {
		panic(err)
	}
	api.SetStore(store)

	defer store.Close()

	port := getPort()
	webDir := "./web"

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	mux.HandleFunc("/api/nextdate", api.NextDateHandler)
	mux.HandleFunc("/api/task", signin.Auth(api.TaskHandler))
	mux.HandleFunc("/api/tasks", signin.Auth(api.TasksHandler))
	mux.HandleFunc("/api/task/done", signin.Auth(api.TaskDoneHandler))
	mux.HandleFunc("/api/signin", signin.SignInHandler)

	err = http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		panic(err)
	}

}
