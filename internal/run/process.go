package run

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TrianaLab/remake/internal/util"
)

var includeRe = regexp.MustCompile(`^\s*include\s+(.+)$`)

// Run resolves includes recursively, checks cycles, writes generated Makefile, and executes make
func Run(targets []string, file string) error {
	os.MkdirAll(".remake", 0755)
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

func processFile(src string, visited map[string]bool, outpath string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	os.MkdirAll(filepath.Dir(outpath), 0755)
	f, _ := os.Create(outpath)
	defer f.Close()
	s := bufio.NewScanner(strings.NewReader(string(data)))
	for s.Scan() {
		line := s.Text()
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
				nested := filepath.Join(".remake", "cache", filepath.Base(local)+".generated")
				if err := processFile(local, visited, nested); err != nil {
					return err
				}
				line = strings.Replace(line, ref, nested, 1)
			}
		}
		f.WriteString(line + "\n")
	}
	return nil
}
