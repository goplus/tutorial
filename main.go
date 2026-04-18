package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"

	gohtml "html"
	htmltemplate "html/template"

	"github.com/goplus/tutorial/internal"
	"github.com/russross/blackfriday/v2"
)

const (
	chNumLen    = 3
	defaultLang = "en"
)

// supportedLangs defines the non-default languages supported by the site.
var supportedLangs = map[string]bool{"zh": true}

// translations holds locale strings keyed by language then by key.
var translations map[string]map[string]string

// titleTranslations holds title translations keyed by language then by original title.
var titleTranslations map[string]map[string]string

// loadTranslations reads all JSON files from the locales/ directory.
func loadTranslations(dir string) {
	translations = make(map[string]map[string]string)
	titleTranslations = make(map[string]map[string]string)
	fis, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Warning: cannot read locales directory %s: %v", dir, err)
		return
	}
	for _, fi := range fis {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".json") {
			continue
		}
		lang := strings.TrimSuffix(fi.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, fi.Name()))
		if err != nil {
			log.Printf("Warning: cannot read locale file %s: %v", fi.Name(), err)
			continue
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			log.Printf("Warning: cannot parse locale file %s: %v", fi.Name(), err)
			continue
		}
		m := make(map[string]string)
		for k, v := range raw {
			if k == "titles" {
				titles := make(map[string]string)
				if err := json.Unmarshal(v, &titles); err == nil {
					titleTranslations[lang] = titles
				}
				continue
			}
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				m[k] = s
			}
		}
		translations[lang] = m
	}
}

// lookup searches a nested translation map for the given language and key,
// falling back to the default language.
func lookup(dict map[string]map[string]string, lang, key string) (string, bool) {
	if m, ok := dict[lang]; ok {
		if v, ok := m[key]; ok {
			return v, true
		}
	}
	if lang != defaultLang {
		if m, ok := dict[defaultLang]; ok {
			if v, ok := m[key]; ok {
				return v, true
			}
		}
	}
	return "", false
}

// t returns the translated string for the given language and key.
func t(lang, key string) htmltemplate.HTML {
	if v, ok := lookup(translations, lang, key); ok {
		return htmltemplate.HTML(v)
	}
	return htmltemplate.HTML(key)
}

// translateTitle returns the localized title for the given language.
func translateTitle(lang, title string) string {
	if v, ok := lookup(titleTranslations, lang, title); ok {
		return v
	}
	return title
}

// extractLang parses the URL path to extract a language prefix.
// Returns (defaultLang, originalPath) for English, or (lang, cleanPath) for other languages.
func extractLang(urlPath string) (lang, cleanPath string) {
	trimmed := strings.TrimPrefix(urlPath, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) >= 1 && supportedLangs[parts[0]] {
		lang = parts[0]
		if len(parts) == 2 {
			cleanPath = "/" + parts[1]
		} else {
			cleanPath = "/"
		}
		return
	}
	return defaultLang, urlPath
}

// langPrefix returns the URL prefix for a language, e.g. "/zh" or "" for English.
func langPrefix(lang string) string {
	if lang == defaultLang {
		return ""
	}
	return "/" + lang
}

var (
	headerTempl string
	footerTempl string
	indexTmpl   *template.Template
	exampleTmpl *template.Template
)

// templateFuncMap provides shared template functions including translation.
var templateFuncMap = template.FuncMap{
	"t": func(lang, key string) htmltemplate.HTML {
		return t(lang, key)
	},
	"langPrefix":     langPrefix,
	"translateTitle": translateTitle,
}

func init() {
	loadTranslations("locales")

	headerTempl = mustReadFile("templates/header.tmpl")
	footerTempl = mustReadFile("templates/footer.tmpl")

	indexTmpl = template.New("index").Funcs(templateFuncMap)
	_, err := indexTmpl.Parse(headerTempl)
	check(err)
	_, err = indexTmpl.Parse(footerTempl)
	check(err)
	_, err = indexTmpl.Parse(mustReadFile("templates/index.tmpl"))
	check(err)

	exampleTmpl = template.New("example").Funcs(templateFuncMap)
	_, err = exampleTmpl.Parse(headerTempl)
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
	cache sync.Map // lang -> *example
}

type chapter struct {
	Title    string
	Articles []*exampleIndex
}

var (
	exampleIndexes map[string]*exampleIndex
	tutorialNames  []string
	watcher        *internal.Watcher
)

func listTutorial(dir string) (names []string, err error) {
	fis, err := os.ReadDir(dir)
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

// indexData wraps the data passed to the index template.
type indexData struct {
	Lang     string
	Chapters []*chapter
}

// buildExampleIndexes builds the global exampleIndexes and chapter structure from tutorial names.
func buildExampleIndexes(tutorial []string) []*chapter {
	var indexes = make(map[string]*exampleIndex, len(tutorial))
	var chs []*chapter
	var ch *chapter
	var prev *exampleIndex
	for _, name := range tutorial {
		title := name[chNumLen+1:]
		titleEsc := gohtml.UnescapeString(strings.ReplaceAll(title, "-", " "))
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
	return chs
}

// renderIndex renders the index page for the given language using pre-built chapters.
func renderIndex(chapters []*chapter, lang string) []byte {
	data := &indexData{
		Lang:     lang,
		Chapters: chapters,
	}
	var buf bytes.Buffer
	err := indexTmpl.Execute(&buf, data)
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

// overrideDocsFromMD replaces the document segments in segs with content from a Markdown file.
// The Markdown file uses "---" separators to delimit sections that correspond to doc segments.
func overrideDocsFromMD(segs []*Seg, mdContent, mdPath string) {
	sections := strings.Split(mdContent, "\n---\n")
	docCount := 0
	for _, seg := range segs {
		if seg.Docs != nil {
			docCount++
		}
	}
	if len(sections) != docCount {
		log.Printf("Warning: %s has %d sections but .gop has %d doc segments", mdPath, len(sections), docCount)
	}
	docIdx := 0
	for _, seg := range segs {
		if seg.Docs != nil && docIdx < len(sections) {
			text := strings.TrimSpace(sections[docIdx])
			seg.Docs = []string{text}
			seg.DocsRendered = string(blackfriday.Run([]byte(text)))
			docIdx++
		}
	}
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

// exampleData wraps the example with language info for template rendering.
type exampleData struct {
	*exampleIndex
	Files []*exampleFile
	Lang  string
}

type example struct {
	*exampleIndex
	Files []*exampleFile
}

// classifySegs classify segments into headDoc, code & tailDoc
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

// replaceExt replaces the file extension with a new one.
func replaceExt(fname, newExt string) string {
	i := strings.LastIndex(fname, ".")
	if i < 0 {
		return fname + newExt
	}
	return fname[:i] + newExt
}

func parseExample(dir string, idx *exampleIndex, lang string) *example {
	fis, err := os.ReadDir(dir)
	check(err)
	ex := &example{exampleIndex: idx}
	for _, fi := range fis {
		fname := fi.Name()
		fileLang := langOf(fname)
		if fileLang != "xgo" && fileLang != "gop" { // only support XGo examples
			continue
		}
		sourcePath := filepath.Join(dir, fname)
		sourceSegs := parseAndRenderSegs(sourcePath)
		if len(sourceSegs) == 0 { // ignore file with no segs
			continue
		}

		// Try to load a language-specific .md override.
		// sourceSegs are freshly parsed so we can mutate them directly.
		if lang != defaultLang {
			mdPath := filepath.Join("locales", lang, idx.Name, replaceExt(fname, ".md"))
			if mdContent, err := os.ReadFile(mdPath); err == nil {
				overrideDocsFromMD(sourceSegs, string(mdContent), mdPath)
			}
		}

		headDoc, code, tailDoc := classifySegs(sourceSegs)
		file := &exampleFile{fileLang, headDoc, code, tailDoc}
		ex.Files = append(ex.Files, file)
	}
	return ex
}

func renderExample(e *example, lang string) []byte {
	data := &exampleData{
		exampleIndex: e.exampleIndex,
		Files:        e.Files,
		Lang:         lang,
	}
	var buf bytes.Buffer
	err := exampleTmpl.Execute(&buf, data)
	check(err)
	return buf.Bytes()
}

func handleExample(w http.ResponseWriter, req *http.Request, root, urlPath, lang string) {
	if idx, ok := exampleIndexes[urlPath]; ok {
		cached, ok := idx.cache.Load(lang)
		if !ok {
			ex := parseExample(filepath.Join(root, idx.Name), idx, lang)
			idx.cache.Store(lang, ex)
			cached = ex
		}
		data := renderExample(cached.(*example), lang)
		w.Write(data)
		return
	}
	http.ServeFile(w, req, "./public/404.html")
}

// -----------------------------------------------------------------------------

func handle(root string) func(w http.ResponseWriter, req *http.Request) {
	var indexCache sync.Map // lang -> []byte
	var chapters []*chapter
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		names, err := listTutorial(root)
		if err != nil {
			log.Panicln(err)
		}
		tutorialNames = names
		chapters = buildExampleIndexes(names)
		enIndex := renderIndex(chapters, defaultLang)
		indexCache.Store(defaultLang, enIndex)
		wg.Done()

		watcher, err = internal.NewWatcher(func(string) {
			for _, v := range exampleIndexes {
				v.cache = sync.Map{}
			}
			indexCache = sync.Map{}
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

		lang, cleanPath := extractLang(urlPath)

		if cleanPath == "/" {
			wg.Wait()
			cached, ok := indexCache.Load(lang)
			if !ok {
				text := renderIndex(chapters, lang)
				indexCache.Store(lang, text)
				cached = text
			}
			w.Write(cached.([]byte))
			return
		}
		if path.Ext(cleanPath) != "" {
			http.ServeFile(w, req, "./public"+cleanPath)
			return
		}
		handleExample(w, req, root, cleanPath, lang)
	}
}

var (
	protocol = "http"
	address  = "localhost:8000"
	host     = flag.String("host", address, "Serving host")
)

func main() {
	flag.Parse()
	fmt.Printf("Serving XGo tutorial at %s://%s\n", protocol, *host)
	http.HandleFunc("/", handle("."))
	err := http.ListenAndServe(*host, nil)
	if err != nil {
		log.Fatalf("Failed to start server at %s://%s.\nError: %s.", protocol, *host, err)
	}
}

// -----------------------------------------------------------------------------
