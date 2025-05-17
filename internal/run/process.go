package run

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TrianaLab/remake/internal/util"
)

var (
	includeBlockRe = regexp.MustCompile(`^\s*include:\s*$`)
	listItemRe     = regexp.MustCompile(`^\s*-\s*(.+)$`)
	includeRe      = regexp.MustCompile(`^\s*include\s+(.+)$`)
)

// Run resolves includes recursively, inlines all referenced Makefiles, writes combined Makefile, and executes make
type RunFunc func(targets []string, file string) error

func Run(targets []string, file string) error {
	if err := os.MkdirAll(".remake", 0755); err != nil {
		return err
	}
	out := filepath.Join(".remake", "Makefile.generated")
	visited := make(map[string]bool)
	if err := processFile(file, visited, out); err != nil {
		return err
	}
	cmd := exec.Command("make", append([]string{"-f", out}, targets...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// processFile reads src, inlines includes, and writes result to outpath without any include directives
func processFile(src string, visited map[string]bool, outpath string) error {
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

		// YAML-style block include
		if includeBlockRe.MatchString(line) {
			// process each list item
			for scanner.Scan() {
				next := scanner.Text()
				if m := listItemRe.FindStringSubmatch(next); m != nil {
					ref := m[1]
					if visited[ref] {
						return fmt.Errorf("cyclic include: %s", ref)
					}
					visited[ref] = true
					// fetch and process nested Makefile
					local, err := util.FetchMakefile(ref)
					if err != nil {
						return err
					}
					nestedOut := filepath.Join(".remake", "cache", filepath.Base(local)+".generated")
					if err := processFile(local, visited, nestedOut); err != nil {
						return err
					}
					// inline content
					if err := inlineFile(f, nestedOut); err != nil {
						return err
					}
				} else {
					// end of block, process this line normally
					line = next
					break
				}
			}
		}

		// single-line include
		if m := includeRe.FindStringSubmatch(line); m != nil {
			refs := strings.Fields(m[1])
			for _, ref := range refs {
				if visited[ref] {
					return fmt.Errorf("cyclic include: %s", ref)
				}
				visited[ref] = true
				local, err := util.FetchMakefile(ref)
				if err != nil {
					return err
				}
				nestedOut := filepath.Join(".remake", "cache", filepath.Base(local)+".generated")
				if err := processFile(local, visited, nestedOut); err != nil {
					return err
				}
				if err := inlineFile(f, nestedOut); err != nil {
					return err
				}
			}
			continue // skip writing the include line
		}

		// write normal line
		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// inlineFile reads the file at path and writes its content to w
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
