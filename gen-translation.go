//go:build ignore

// gen-translation generates skeleton .md translation files for a given language.
//
// Usage:
//
//	go run gen-translation.go <lang> [tutorial-dir...]
//
// Examples:
//
//	go run gen-translation.go zh                    # generate for all tutorials
//	go run gen-translation.go zh 101-Hello-world    # generate for specific tutorial
//	go run gen-translation.go ja 101-Hello-world 102-Values
//
// The generated .md files contain the original English doc segments as placeholders,
// separated by "---". Translators replace the English text with the target language.
// If a .md file already exists, it will NOT be overwritten.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const chNumLen = 3

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run gen-translation.go <lang> [tutorial-dir...]")
		os.Exit(1)
	}
	lang := os.Args[1]
	if !regexp.MustCompile(`^[a-z]{2,10}$`).MatchString(lang) {
		fmt.Fprintln(os.Stderr, "Error: lang must be a simple language code (e.g. zh, ja, ko)")
		os.Exit(1)
	}

	var dirs []string
	if len(os.Args) > 2 {
		dirs = os.Args[2:]
	} else {
		// Find all tutorial directories
		fis, err := os.ReadDir(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading current directory:", err)
			os.Exit(1)
		}
		for _, fi := range fis {
			if fi.IsDir() {
				name := fi.Name()
				if len(name) > (chNumLen+1) && name[chNumLen] == '-' {
					if _, e := strconv.Atoi(name[:chNumLen]); e == nil {
						// Skip chapter headings (e.g. 100-, 200-)
						if !strings.HasSuffix(name[:chNumLen], "00") {
							dirs = append(dirs, name)
						}
					}
				}
			}
		}
	}

	total := 0
	skipped := 0
	for _, dir := range dirs {
		fis, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read %s: %v\n", dir, err)
			continue
		}
		langDir := filepath.Join("locales", lang, dir)
		dirCreated := false
		for _, fi := range fis {
			fname := fi.Name()
			ext := filepath.Ext(fname)
			if ext != ".gop" && ext != ".xgo" {
				continue
			}

			mdName := strings.TrimSuffix(fname, ext) + ".md"
			mdPath := filepath.Join("locales", lang, dir, mdName)

			// Skip if already exists
			if _, err := os.Stat(mdPath); err == nil {
				skipped++
				continue
			} else if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: cannot stat %s: %v\n", mdPath, err)
				continue
			}

			// Parse .gop to extract doc segments
			content, err := os.ReadFile(filepath.Join(dir, fname))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: cannot read %s: %v\n", fname, err)
				continue
			}

			docs := extractDocSegments(string(content))
			if len(docs) == 0 {
				continue
			}

			// Create lang directory once per tutorial dir
			if !dirCreated {
				if err := os.MkdirAll(langDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", langDir, err)
					break
				}
				dirCreated = true
			}

			// Write skeleton .md
			skeleton := strings.Join(docs, "\n\n---\n\n")
			if err := os.WriteFile(mdPath, []byte(skeleton+"\n"), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", mdPath, err)
				continue
			}

			fmt.Printf("Created: %s (%d segments)\n", mdPath, len(docs))
			total++
		}
	}
	fmt.Printf("\nDone: %d files created, %d skipped (already exist)\n", total, skipped)
}

// extractDocSegments parses a .gop file and returns the text of each doc segment.
func extractDocSegments(content string) []string {
	lines := strings.Split(content, "\n")
	var segments []string
	var current []string
	inDoc := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isDoc := false
		var docText string

		if strings.HasPrefix(trimmed, "//") {
			isDoc = true
			docText = strings.TrimPrefix(trimmed[2:], " ")
		} else if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!") {
			isDoc = true
			docText = "## " + strings.TrimPrefix(strings.TrimPrefix(trimmed, "#"), " ")
		}

		if isDoc {
			current = append(current, docText)
			inDoc = true
		} else {
			if inDoc && len(current) > 0 {
				segments = append(segments, strings.Join(current, "\n"))
				current = nil
			}
			inDoc = false
		}
	}
	// Flush last segment
	if len(current) > 0 {
		segments = append(segments, strings.Join(current, "\n"))
	}

	return segments
}
