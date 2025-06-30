package writer

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"code4context/internal/outline"
)

// WriteOutlineToFile writes the outline to codebrev.md
func WriteOutlineToFile(out *outline.Outline) error {
	file, err := os.Create("codebrev.md")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintln(writer, "# Code Structure Outline")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "This file provides an overview of available functions, types, and variables per file for LLM context.")
	fmt.Fprintln(writer, "")

	// Sort file paths for consistent output
	var filePaths []string
	for path := range out.Files {
		filePaths = append(filePaths, path)
	}
	sort.Strings(filePaths)

	// Write file-by-file breakdown
	for _, path := range filePaths {
		fileInfo := out.Files[path]
		fmt.Fprintf(writer, "## %s\n", path)
		fmt.Fprintln(writer, "")

		// Functions available in this file
		if len(fileInfo.Functions) > 0 {
			// Sort functions by name
			sort.Slice(fileInfo.Functions, func(i, j int) bool {
				return fileInfo.Functions[i].Name < fileInfo.Functions[j].Name
			})
			fmt.Fprintln(writer, "### Functions")
			for _, f := range fileInfo.Functions {
				params := strings.Join(f.Params, ", ")
				if f.ReturnType != "" {
					fmt.Fprintf(writer, "- %s(%s) -> %s\n", f.Name, params, f.ReturnType)
				} else {
					fmt.Fprintf(writer, "- %s(%s)\n", f.Name, params)
				}
			}
			fmt.Fprintln(writer, "")
		}

		// Types/Structs/Classes available in this file
		if len(fileInfo.Types) > 0 {
			sort.Strings(fileInfo.Types)
			fmt.Fprintln(writer, "### Types")
			for _, t := range fileInfo.Types {
				fmt.Fprintf(writer, "- %s", t)
				if ti, exists := out.Types[t]; exists {
					if len(ti.Methods) > 0 {
						fmt.Fprintf(writer, " (methods: %s)", strings.Join(ti.Methods, ", "))
					}
					if len(ti.Fields) > 0 {
						fmt.Fprintf(writer, " (fields: %s)", strings.Join(ti.Fields, ", "))
					}
				}
				fmt.Fprintln(writer, "")
			}
			fmt.Fprintln(writer, "")
		}

		// Variables available in this file
		if len(fileInfo.Vars) > 0 {
			sort.Strings(fileInfo.Vars)
			fmt.Fprintln(writer, "### Variables")
			for _, v := range fileInfo.Vars {
				fmt.Fprintf(writer, "- %s\n", v)
			}
			fmt.Fprintln(writer, "")
		}

		fmt.Fprintln(writer, "---")
		fmt.Fprintln(writer, "")
	}

	fmt.Println("Code outline written to codebrev.md")
	return nil
}
