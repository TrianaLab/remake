package cmd

import "os"

func DetermineMakefile(defaultName string) string {
	if _, err := os.Stat("Makefile"); err == nil {
		return "Makefile"
	}
	if _, err := os.Stat("makefile"); err == nil {
		return "makefile"
	}
	return defaultName
}
