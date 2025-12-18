package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jasonwillschiu/codebrev/internal/outline"
	"github.com/jasonwillschiu/codebrev/internal/parser"
	"github.com/jasonwillschiu/codebrev/internal/writer"
)

// Build information (set via ldflags during build)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Command line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		outputFile  = flag.String("output", "", "Output file path (defaults to 'codebrev.md' in target directory)")
	)
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("codebrev %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	// Handle help flag
	if *showHelp {
		showHelpMessage()
		return
	}

	// Get remaining arguments (directory path)
	args := flag.Args()

	// Run CLI mode
	runCLIMode(args, *outputFile)
}

func showHelpMessage() {
	fmt.Println("codebrev - Code Context Generator")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  codebrev [OPTIONS] [DIRECTORY]")
	fmt.Println("")
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Generate a codebrev.md file containing code structure outline for the specified directory.")
	fmt.Println("  If no directory is specified, defaults to current directory.")
	fmt.Println("")
	fmt.Println("OPTIONS:")
	fmt.Println("  --version         Show version information")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --output FILE     Output file path (defaults to 'codebrev.md' in target directory)")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  codebrev .                    # Generate codebrev.md for current directory")
	fmt.Println("  codebrev /path/to/project     # Generate codebrev.md for specified directory")
	fmt.Println("  codebrev --output custom.md . # Generate with custom output filename")
}

func runCLIMode(args []string, outputFile string) {
	// Default to current directory if no directory specified
	directoryPath := "."
	if len(args) > 0 {
		directoryPath = args[0]
	}

	// Set default output file if not specified
	if outputFile == "" {
		outputFile = filepath.Join(directoryPath, "codebrev.md")
	}

	fmt.Printf("Generating code context for: %s\n", directoryPath)
	fmt.Printf("Output file: %s\n", outputFile)

	// Check if directory exists
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory does not exist: %s\n", directoryPath)
		os.Exit(1)
	}

	// Generate the code context
	err := generateCodeContext(directoryPath, outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code context: %v\n", err)
		os.Exit(1)
	}

	// Get file size for reporting
	fileInfo, err := os.Stat(outputFile)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	fmt.Printf("Successfully generated code context outline\n")
	fmt.Printf("File: %s (%d bytes)\n", outputFile, fileSize)
}

// generateCodeContext generates the code context outline using the existing parser and writer
func generateCodeContext(directoryPath, outputFile string) error {
	// Create new outline
	out := outline.New()

	// Process all files in the directory
	err := parser.ProcessFiles(directoryPath, out)
	if err != nil {
		return fmt.Errorf("failed to process files: %v", err)
	}

	// Remove duplicates
	out.RemoveDuplicates()

	// Write output to the specified file
	err = writer.WriteOutlineToFileWithPath(out, outputFile)
	if err != nil {
		return fmt.Errorf("failed to write outline: %v", err)
	}

	return nil
}
