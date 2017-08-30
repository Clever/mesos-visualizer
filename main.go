package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/Clever/mesos-visualizer/ecs"
)

var (
	Clusters           map[string]string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
)

func init() {
	Clusters = getEnvJSON("CLUSTERS")
	AWSAccessKeyID = getEnv("AWS_ACCESS_KEY_ID")
	AWSSecretAccessKey = getEnv("AWS_SECRET_ACCESS_KEY")

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
			<li><a href="./static/sunburst.html?{{$arn}}">Resource Utilization - Sunburst</a></li>
			<li><a href="./static/treemap.html?{{$arn}}">Resource Utilization - Treemap</a></li>
		</ul>
		{{end}}
	</body>
	`)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, Clusters)
	if err != nil {
		panic(err)
	}
}

func resourcesHandler(w http.ResponseWriter, req *http.Request) {
	cluster := strings.TrimPrefix(req.URL.Path, "/resources/")

	c := ecs.NewClient(cluster, AWSAccessKeyID, AWSSecretAccessKey)
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
		log.Fatal(err)
	}

	return keyval
}
