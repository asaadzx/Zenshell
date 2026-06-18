package shell

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// PATH cache — scanned once at startup
var (
	pathCache   []string
	pathCacheMu sync.Once
)

func buildPathCache() {
	dirs := filepath.SplitList(os.Getenv("PATH"))
	seen := make(map[string]bool)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !seen[name] {
				seen[name] = true
				pathCache = append(pathCache, name)
			}
		}
	}
	sort.Strings(pathCache)
}

func (s *Shell) Do(line []rune, pos int) ([][]rune, int) {
	pathCacheMu.Do(buildPathCache)

	input := string(line[:pos])
	words := strings.Fields(input)
	var prefix string
	if len(words) > 0 && !strings.HasSuffix(input, " ") {
		prefix = words[len(words)-1]
	}

	isFirstWord := len(words) == 0 || (len(words) == 1 && !strings.HasSuffix(input, " "))
	var candidates []string

	if isFirstWord {
		// Command completion from PATH cache
		for _, cmd := range pathCache {
			if strings.HasPrefix(cmd, prefix) {
				candidates = append(candidates, cmd)
			}
		}
	}

	// File completion (also for first-word to handle ./foo)
	fileDir := filepath.Dir(prefix)
	if fileDir == "." {
		fileDir = ""
	}
	filePrefix := filepath.Base(prefix)

	// Expand ~ in file paths
	searchDir := fileDir
	if strings.HasPrefix(searchDir, "~") {
		searchDir = s.home + searchDir[1:]
	}
	if searchDir == "" {
		searchDir = "."
	}

	entries, err := os.ReadDir(searchDir)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if !strings.HasPrefix(name, filePrefix) {
				continue
			}
			full := name
			if fileDir != "" {
				full = fileDir + "/" + name
			}
			if e.IsDir() {
				full += "/"
			}
			candidates = append(candidates, full)
		}
	}

	if len(candidates) == 0 {
		return nil, 0
	}

	sort.Strings(candidates)
	completions := make([][]rune, len(candidates))
	for i, c := range candidates {
		completions[i] = []rune(c)
	}
	return completions, len([]rune(prefix))
}
