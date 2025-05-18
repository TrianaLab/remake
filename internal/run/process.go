package run

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/util"
)

var (
	includeBlockRe = regexp.MustCompile(`^\s*include:\s*$`)
	listItemRe     = regexp.MustCompile(`^\s*-\s*(.+)$`)
	includeRe      = regexp.MustCompile(`^\s*include\s+(.+)$`)
)

// Run generates a combined Makefile and executes the given targets.
func Run(targets []string, file string, useCache bool) error {
	cacheDir := config.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	genFile := filepath.Join(cacheDir, "Makefile.generated")
	visited := make(map[string]bool)
	if err := processFile(file, visited, genFile, useCache); err != nil {
		return err
	}
	cmd := exec.Command("make", append([]string{"-f", genFile}, targets...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Template resolves includes and writes the result to outpath.
func Template(src, outpath string, useCache bool) (string, error) {
	if err := os.MkdirAll(filepath.Dir(outpath), 0755); err != nil {
		return "", err
	}
	visited := make(map[string]bool)
	if err := processFile(src, visited, outpath, useCache); err != nil {
		return "", err
	}
	return outpath, nil
}

func processFile(src string, visited map[string]bool, outpath string, useCache bool) error {
	// Fetch or use local
	fetcher, ferr := util.GetFetcher(src)
	if ferr == nil {
		fpath, err := fetcher.Fetch(src, useCache)
		if err != nil {
			return err
		}
		if fpath != "" {
			src = fpath
		}
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(outpath), 0755); err != nil {
		return err
	}
	outFile, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()

		// include: block
		if includeBlockRe.MatchString(line) {
			for scanner.Scan() {
				next := scanner.Text()
				if m := listItemRe.FindStringSubmatch(next); m != nil {
					if err := handleInclude(m[1], visited, outFile, useCache); err != nil {
						return err
					}
				} else {
					line = next
					break
				}
			}
		}

		// inline single include
		if m := includeRe.FindStringSubmatch(line); m != nil {
			for _, ref := range strings.Fields(m[1]) {
				if err := handleInclude(ref, visited, outFile, useCache); err != nil {
					return err
				}
			}
			continue
		}

		// normalize indentation
		if len(line) > 0 && line[0] == ' ' {
			line = "\t" + strings.TrimLeft(line, " ")
		}

		if _, err := outFile.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func handleInclude(ref string, visited map[string]bool, outFile *os.File, useCache bool) error {
	if visited[ref] {
		return fmt.Errorf("cyclic include: %s", ref)
	}
	visited[ref] = true

	// Fetch referenced Makefile
	fetcher, ferr := util.GetFetcher(ref)
	if ferr == nil {
		fpath, err := fetcher.Fetch(ref, useCache)
		if err != nil {
			return err
		}
		if fpath != "" {
			ref = fpath
		}
	}

	// Unique nested output
	cacheDir := config.GetCacheDir()
	unique := fmt.Sprintf("%x.mk", sha256.Sum256([]byte(ref)))
	nestedOut := filepath.Join(cacheDir, unique+".generated")
	if err := processFile(ref, visited, nestedOut, useCache); err != nil {
		return err
	}
	return inlineFile(outFile, nestedOut)
}

func inlineFile(w io.Writer, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("inline read %s: %w", path, err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("inline write %s: %w", path, err)
	}
	return nil
}
