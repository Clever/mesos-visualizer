package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Clever/mesos-visualizer/ecs"
)

var (
	clusters map[string]string
)

func init() {
	clusters = getEnvJSON("CLUSTERS")
}

func main() {
	http.HandleFunc("/resources/", resourcesHandler)
	http.HandleFunc("/", indexHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	log.Print("Listening on port 80...")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(`
	<!DOCTYPE html>
	<head>
		<meta charset="utf-8">
		<title>ECS Visualizations</title>
	</head>
	<body>
		{{ range $name, $arn := . }}
		<h2>{{$name}}</h2>
		<ul>
			<li><a href="./static/sunburst.html?{{$name}}">Resource Utilization - Sunburst</a></li>
			<li><a href="./static/treemap.html?{{$name}}">Resource Utilization - Treemap</a></li>
		</ul>
		{{end}}
	</body>
	`)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, clusters)
	if err != nil {
		panic(err)
	}
}

func resourcesHandler(w http.ResponseWriter, req *http.Request) {
	cluster := strings.TrimPrefix(req.URL.Path, "/resources/")
	arn, ok := clusters[cluster]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(`{"error": "unknown cluster"}`))
		return
	}

	c := ecs.NewClient(arn)
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

func getEnvJSON(envVar string) map[string]string {
	data := getEnv(envVar)

	var keyval map[string]string
	err := json.Unmarshal([]byte(data), &keyval)
	if err != nil {
		log.Fatal(err.Error())
	}

	return keyval
}
