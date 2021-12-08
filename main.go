package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"unsafe"

	gohtml "html"

	"github.com/goplus/tutorial/internal"
	"github.com/russross/blackfriday/v2"
)

const (
	chNumLen = 3
)

var (
	headerTempl string
	footerTempl string
	exampleTmpl *template.Template
)

func init() {
	headerTempl = mustReadFile("templates/header.tmpl")
	footerTempl = mustReadFile("templates/footer.tmpl")
	exampleTmpl = template.New("example")
	_, err := exampleTmpl.Parse(headerTempl)
	check(err)
	_, err = exampleTmpl.Parse(footerTempl)
	check(err)
	_, err = exampleTmpl.Parse(mustReadFile("templates/example.tmpl"))
	check(err)
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

// -----------------------------------------------------------------------------

type exampleIndex struct {
	Path  string
	Name  string
	Title string
	Prev  *exampleIndex
	Next  *exampleIndex
	cache *example
}

type chapter struct {
	Title    string
	Articles []*exampleIndex
}

var (
	exampleIndexes map[string]*exampleIndex
	watcher        *internal.Watcher
)

func listTutorial(dir string) (names []string, err error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			name := fi.Name()
			if len(name) > (chNumLen+1) && name[chNumLen] == '-' {
				if _, e := strconv.Atoi(name[:1]); e == nil {
					names = append(names, name)
				}
			}
		}
	}
	return
}

func renderIndex(tutorial []string) []byte {
	indexTmpl := template.New("index")
	_, err := indexTmpl.Parse(headerTempl)
	check(err)
	_, err = indexTmpl.Parse(footerTempl)
	check(err)
	_, err = indexTmpl.Parse(mustReadFile("templates/index.tmpl"))
	check(err)

	var buf bytes.Buffer
	var indexes = make(map[string]*exampleIndex, len(tutorial))
	var chs []*chapter
	var ch *chapter
	var prev *exampleIndex
	for _, name := range tutorial {
		title := name[chNumLen+1:]
		titleEsc := strings.ReplaceAll(title, "-", " ")
		if strings.HasSuffix(name[:chNumLen], "00") {
			ch = &chapter{Title: titleEsc}
			chs = append(chs, ch)
			continue
		}
		idx := &exampleIndex{
			Path:  "/" + strings.ToLower(strings.ReplaceAll(gohtml.UnescapeString(title), "/", "-")),
			Name:  name,
			Title: titleEsc,
		}
		if ch == nil {
			ch = new(chapter)
			chs = append(chs, ch)
		}
		ch.Articles = append(ch.Articles, idx)
		indexes[idx.Path] = idx
		if prev != nil {
			prev.Next = idx
		}
		idx.Prev = prev
		prev = idx
	}
	exampleIndexes = indexes
	err = indexTmpl.Execute(&buf, chs)
	check(err)
	return buf.Bytes()
}

// -----------------------------------------------------------------------------

// Seg is a segment of an example
type Seg struct {
	Docs         []string
	Code         []string
	DocsRendered string
	CodeRendered string
}

const (
	ltNone  = -1
	ltCode  = 0 // ltCode must be 0
	ltDoc   = 1
	ltBlank = 2
)

func checkLineType(line string) (doc string, lt int) {
	doc = strings.TrimSpace(line)
	if strings.HasPrefix(doc, "//") {
		return strings.TrimPrefix(doc[2:], " "), ltDoc
	}
	if strings.HasPrefix(doc, "#") && !strings.HasPrefix(doc, "#!") {
		doc = "##" + doc
		lt = ltDoc
	} else if doc == "" {
		lt = ltBlank
	}
	return
}

func parseSegs(filecontent string) (segs []*Seg) {
	source := strings.Split(filecontent, "\n")
	lines := make([]string, len(source))
	for i, line := range source {
		// Convert tabs to spaces for uniform rendering.
		lines[i] = strings.ReplaceAll(line, "\t", "    ")
	}

	var lastSeg *Seg
	var lastSeen = ltNone
	for _, line := range lines {
		trimmed, lt := checkLineType(line)
		if lt == ltDoc || (lt == ltBlank && lastSeen == ltDoc) {
			if lt == ltDoc {
				if lastSeen == ltDoc {
					lastSeg.Docs = append(lastSeg.Docs, trimmed)
				} else {
					lastSeg = &Seg{Docs: []string{trimmed}}
					segs = append(segs, lastSeg)
				}
			}
			lastSeen = lt
		} else if lt == ltCode || lastSeen == ltCode {
			if lastSeen != ltBlank && lastSeg != nil {
				lastSeg.Code = append(lastSeg.Code, line)
			} else {
				lastSeg = &Seg{Code: []string{line}}
				segs = append(segs, lastSeg)
			}
			lastSeen = ltCode
		}
	}
	return
}

// Join concatenates the elements of its first argument to create a single string. The separator
// string sep is placed between elements in the resulting string.
func stringJoin(elems []string, sep string) []byte {
	switch len(elems) {
	case 0:
		return nil
	case 1:
		return []byte(elems[0])
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b bytes.Buffer
	b.Grow(n)
	b.WriteString(elems[0])
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(s)
	}
	return b.Bytes()
}

func markdown(docs []string) []byte {
	text := stringJoin(docs, "\n")
	return blackfriday.Run(text)
}

func parseAndRenderSegs(sourcePath string) []*Seg {
	filecontent := mustReadFile(sourcePath)
	segs := parseSegs(filecontent)
	for _, seg := range segs {
		if seg.Docs != nil {
			seg.DocsRendered = string(markdown(seg.Docs))
		}
		if seg.Code != nil {
			code := strings.Join(seg.Code, "\n")
			seg.CodeRendered = code
		}
	}
	return segs
}

// -----------------------------------------------------------------------------

type exampleFile struct {
	// Lang is the language of code file, like `gop` / `go`
	Lang string
	// HeadDoc is document content before code content
	HeadDoc []*Seg
	// Code is all code segments of example file
	Code []*Seg
	// TailDoc is document content after code content
	TailDoc []*Seg
}

type example struct {
	*exampleIndex
	Files []*exampleFile
}

func classifySegs(segs []*Seg) (headDoc, code, tailDoc []*Seg) {
	hasSeenCode := false
	for _, seg := range segs {
		if seg.Code != nil {
			code = append(code, seg)
			hasSeenCode = true
			continue
		}
		if hasSeenCode {
			// move all doc segs among code segs to tail,
			// so that single code block will be rendered as result
			tailDoc = append(tailDoc, seg)
		} else {
			headDoc = append(headDoc, seg)
		}
	}
	return
}

func langOf(fname string) string {
	i := strings.LastIndex(fname, ".")
	if i < 0 {
		return ""
	}
	return fname[i+1:]
}

func parseExample(dir string, idx *exampleIndex) *example {
	fis, err := ioutil.ReadDir(dir)
	check(err)
	example := &example{exampleIndex: idx}
	for _, fi := range fis {
		fname := fi.Name()
		lang := langOf(fname)
		sourcePath := filepath.Join(dir, fname)
		sourceSegs := parseAndRenderSegs(sourcePath)
		if len(sourceSegs) != 0 { // ignore file with no segs
			headDoc, code, tailDoc := classifySegs(sourceSegs)
			file := &exampleFile{lang, headDoc, code, tailDoc}
			example.Files = append(example.Files, file)
		}
	}
	return example
}

func renderExample(e *example) []byte {
	var buf bytes.Buffer
	err := exampleTmpl.Execute(&buf, e)
	check(err)
	return buf.Bytes()
}

func handleExample(w http.ResponseWriter, req *http.Request, root, path string) {
	if idx, ok := exampleIndexes[path]; ok {
		cache := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&idx.cache)))
		if cache == nil {
			cache = unsafe.Pointer(parseExample(filepath.Join(root, idx.Name), idx))
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&idx.cache)), cache)
		}
		data := renderExample((*example)(cache))
		w.Write(data)
		return
	}
	http.ServeFile(w, req, "./public/404.html")
}

// -----------------------------------------------------------------------------

func handle(root string) func(w http.ResponseWriter, req *http.Request) {
	var text []byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		names, err := listTutorial(root)
		if err != nil {
			log.Panicln(err)
		}
		text = renderIndex(names)
		wg.Done()

		watcher, err = internal.NewWatcher(func(string) {
			for _, v := range exampleIndexes {
				atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&v.cache)), nil)
			}
		})
		if err != nil {
			log.Panicln(err)
		}
		if err := watcher.Watch(root); err != nil {
			log.Panicln(err)
		}
	}()

	return func(w http.ResponseWriter, req *http.Request) {
		urlPath := path.Clean(req.URL.Path)
		if !path.IsAbs(urlPath) {
			urlPath = "/404.html"
		}
		if urlPath == "/" {
			wg.Wait()
			w.Write(text)
			return
		}
		if path.Ext(urlPath) != "" {
			http.ServeFile(w, req, "./public"+urlPath)
			return
		}
		handleExample(w, req, root, urlPath)
	}
}

var (
	host = flag.String("host", "localhost:8000", "Serving host")
)

func main() {
	flag.Parse()
	fmt.Println("Serving Go+ tutorial at", *host)
	http.HandleFunc("/", handle("."))
	http.ListenAndServe(*host, nil)
}

// -----------------------------------------------------------------------------
