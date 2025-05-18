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

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/util"
)

var includeBlockRe = regexp.MustCompile(`^\s*include:\s*$`)
var listItemRe = regexp.MustCompile(`^\s*-\s*(.+)$`)
var includeRe = regexp.MustCompile(`^\s*include\s+(.+)$`)

func Run(targets []string, file string) error {
	cacheDir := config.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	out := filepath.Join(cacheDir, "Makefile.generated")
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
		if includeBlockRe.MatchString(line) {
			for scanner.Scan() {
				next := scanner.Text()
				if m := listItemRe.FindStringSubmatch(next); m != nil {
					ref := m[1]
					if visited[ref] {
						return fmt.Errorf("cyclic include: %s", ref)
					}
					visited[ref] = true
					local, err := util.FetchMakefile(ref)
					if err != nil {
						return err
					}
					cacheDir := config.GetCacheDir()
					nestedOut := filepath.Join(cacheDir, filepath.Base(local)+".generated")
					if err := processFile(local, visited, nestedOut); err != nil {
						return err
					}
					if err := inlineFile(f, nestedOut); err != nil {
						return err
					}
				} else {
					line = next
					break
				}
			}
		}
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
				cacheDir := config.GetCacheDir()
				nestedOut := filepath.Join(cacheDir, filepath.Base(local)+".generated")
				if err := processFile(local, visited, nestedOut); err != nil {
					return err
				}
				if err := inlineFile(f, nestedOut); err != nil {
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
