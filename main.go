package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/Clever/mesos-visualizer/ecs"
)

var (
	Cluster string
)

func init() {
	Cluster = getEnv("CLUSTER")
}

func main() {
	http.HandleFunc("/resources.json", resourcesHandler)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func resourcesHandler(w http.ResponseWriter, r *http.Request) {
	c := ecs.NewClient()
	resourceGraph, err := c.GetResourceGraph()
	if err != nil {
		log.Fatal(err)
	}
	js, err := json.Marshal(resourceGraph)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(js)
}

func getEnv(envVar string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Fatalf("Must specify env variable %s", envVar)
	}
	return val
}
