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

// Render processes src (local or remote) resolving includes, writes to outpath.
func Render(src, outpath string, useCache bool) error {
	// ensure cache dir
	if err := os.MkdirAll(filepath.Dir(outpath), 0755); err != nil {
		return err
	}
	visited := make(map[string]bool)
	return processFile(src, visited, outpath, useCache)
}

// Run executes make target after rendering includes into a temp file.
func Run(src string, targets []string, useCache bool) error {
	cacheDir := config.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	gen := filepath.Join(cacheDir, "Makefile.generated")
	if err := Render(src, gen, useCache); err != nil {
		return err
	}
	cmd := exec.Command("make", append([]string{"-f", gen}, targets...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	_ = os.Remove(gen)
	return err
}

func processFile(src string, visited map[string]bool, outpath string, useCache bool) error {
	// fetch if remote
	fetcher, ferr := util.GetFetcher(src)
	if ferr == nil {
		if path, err := fetcher.Fetch(src, useCache); err != nil {
			return err
		} else if path != "" {
			src = path
		}
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(outpath), 0755); err != nil {
		return err
	}
	f, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()

		if includeBlockRe.MatchString(line) {
			for scanner.Scan() {
				next := scanner.Text()
				if m := listItemRe.FindStringSubmatch(next); m != nil {
					if err := handleInclude(m[1], visited, f, useCache); err != nil {
						return err
					}
				} else {
					line = next
					break
				}
			}
		}

		if m := includeRe.FindStringSubmatch(line); m != nil {
			for _, ref := range strings.Fields(m[1]) {
				if err := handleInclude(ref, visited, f, useCache); err != nil {
					return err
				}
			}
			continue
		}

		if len(line) > 0 && line[0] == ' ' {
			line = "\t" + strings.TrimLeft(line, " ")
		}

		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handleInclude(ref string, visited map[string]bool, w io.Writer, useCache bool) error {
	if visited[ref] {
		return fmt.Errorf("cyclic include detected: %s", ref)
	}
	visited[ref] = true

	// fetch nested
	fetcher, ferr := util.GetFetcher(ref)
	if ferr == nil {
		if path, err := fetcher.Fetch(ref, useCache); err != nil {
			return err
		} else if path != "" {
			ref = path
		}
	}

	// render nested
	cacheDir := config.GetCacheDir()
	unique := fmt.Sprintf("%x.mk", sha256.Sum256([]byte(ref)))
	nestedOut := filepath.Join(cacheDir, unique+".generated")
	if err := processFile(ref, visited, nestedOut, useCache); err != nil {
		return err
	}
	// inline
	data, err := os.ReadFile(nestedOut)
	if err != nil {
		return fmt.Errorf("inline read %s: %w", nestedOut, err)
	}
	_, err = w.Write(data)
	return err
}
