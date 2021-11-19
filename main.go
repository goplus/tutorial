package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type index struct {
	Path  string
	Name  string
	Title string
}

var (
	host = flag.String("host", "localhost:8000", "Serving host")
)

func listTutorial(dir string) (names []string, err error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			name := fi.Name()
			if len(name) > 3 && name[2] == '-' {
				if _, e := strconv.Atoi(name[:2]); e == nil {
					names = append(names, name)
				}
			}
		}
	}
	return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func mustReadFile(path string) string {
	bytes, err := os.ReadFile(path)
	check(err)
	return string(bytes)
}

func renderIndex(tutorial []string) []byte {
	indexTmpl := template.New("index")
	_, err := indexTmpl.Parse(mustReadFile("templates/footer.tmpl"))
	check(err)
	_, err = indexTmpl.Parse(mustReadFile("templates/index.tmpl"))
	check(err)

	var buf bytes.Buffer
	var examples = make([]*index, len(tutorial))
	for i, name := range tutorial {
		title := name[3:]
		examples[i] = &index{
			Path:  "/" + strings.ToLower(title),
			Name:  name,
			Title: strings.ReplaceAll(title, "-", " "),
		}
	}
	err = indexTmpl.Execute(&buf, examples)
	check(err)
	return buf.Bytes()
}

func handle(root string) func(w http.ResponseWriter, req *http.Request) {
	result := make(chan []byte, 1)
	go func() {
		names, err := listTutorial(root)
		if err != nil {
			log.Panicln(err)
		}
		result <- renderIndex(names)
	}()
	return func(w http.ResponseWriter, req *http.Request) {
		urlPath := path.Clean(req.URL.Path)
		if !path.IsAbs(urlPath) {
			urlPath = "/404.html"
		}
		if urlPath == "/" {
			text := <-result
			w.Write(text)
			return
		}
		if path.Ext(urlPath) != "" {
			http.ServeFile(w, req, "./public"+urlPath)
			return
		}
		fmt.Fprintln(w, urlPath)
	}
}

func main() {
	flag.Parse()
	fmt.Println("Serving Go+ tutorial at", *host)
	http.HandleFunc("/", handle("."))
	http.ListenAndServe(*host, nil)
}
