package main

import (
	"log"
	"os"

	"code4context/internal/outline"
	"code4context/internal/parser"
	"code4context/internal/writer"
)

func main() {
	root := "." // start in current dir; override with arg[1] if you like
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	// Create new outline
	out := outline.New()

	// Process all files
	err := parser.ProcessFiles(root, out)
	if err != nil {
		log.Fatal(err)
	}

	// Remove duplicates
	out.RemoveDuplicates()

	// Write output
	err = writer.WriteOutlineToFile(out)
	if err != nil {
		log.Fatal(err)
	}
}
