package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"code4context/internal/outline"
	"code4context/internal/parser"
	"code4context/internal/writer"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
		mcpMode     = flag.Bool("mcp", false, "Run in MCP server mode (default behavior when no directory specified)")
	)
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("code4context %s\n", Version)
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

	// Determine mode: CLI or MCP server
	if *mcpMode {
		// Explicit MCP mode requested
		runMCPMode()
	} else if len(args) > 0 {
		// CLI mode: directory specified
		runCLIMode(args, *outputFile)
	} else {
		// Default to MCP mode when no arguments provided
		runMCPMode()
	}
}

func showHelpMessage() {
	fmt.Println("code4context - Code Context Generator")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  code4context [OPTIONS] [DIRECTORY]")
	fmt.Println("  code4context [OPTIONS] --mcp")
	fmt.Println("")
	fmt.Println("MODES:")
	fmt.Println("  CLI Mode (default when directory specified):")
	fmt.Println("    Generate codebrev.md file for the specified directory")
	fmt.Println("")
	fmt.Println("  MCP Server Mode (default when no directory specified):")
	fmt.Println("    Run as Model Context Protocol server for AI agents")
	fmt.Println("")
	fmt.Println("OPTIONS:")
	fmt.Println("  --version         Show version information")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --output FILE     Output file path (CLI mode only)")
	fmt.Println("  --mcp             Force MCP server mode")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  code4context .                    # Generate codebrev.md for current directory")
	fmt.Println("  code4context /path/to/project     # Generate codebrev.md for specified directory")
	fmt.Println("  code4context --output custom.md . # Generate with custom output filename")
	fmt.Println("  code4context --mcp                # Run as MCP server")
	fmt.Println("")
	fmt.Println("MCP SERVER TOOLS:")
	fmt.Println("  - generate_code_context: Generate codebrev.md for a directory")
	fmt.Println("  - get_code_context: Get cached codebrev.md or generate if missing")
	fmt.Println("")
	fmt.Println("MCP CONFIGURATION:")
	fmt.Println("  Claude: claude mcp add code4context -- code4context")
	fmt.Println("  Cursor (mcp.json):")
	fmt.Println(`    "code4context": {`)
	fmt.Println(`      "command": "code4context",`)
	fmt.Println(`      "args": ["--mcp"]`)
	fmt.Println(`    }`)
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

func runMCPMode() {
	// Set up logging to stderr so it doesn't interfere with stdio communication
	log.SetOutput(os.Stderr)
	log.SetPrefix("[code4context-mcp] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Starting code4context MCP Server %s", Version)

	// Create a new MCP server
	s := server.NewMCPServer(
		"code4context",
		Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Add the code context generation tools
	log.Println("Registering code context tools...")
	addGenerateCodeContextTool(s)
	addGetCodeContextTool(s)
	log.Println("Tools registered successfully")

	// Start the stdio server
	log.Println("Starting stdio server...")
	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
		fmt.Printf("Server error: %v\n", err)
	}
}

// addGenerateCodeContextTool adds the tool to generate codebrev.md for a directory
func addGenerateCodeContextTool(s *server.MCPServer) {
	generateTool := mcp.NewTool("generate_code_context",
		mcp.WithDescription("Generate a codebrev.md file containing code structure outline for the specified directory"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to the directory to analyze (defaults to current directory if not specified)"),
		),
		mcp.WithString("output_file",
			mcp.Description("Output file path (defaults to 'codebrev.md' in the target directory)"),
		),
	)

	s.AddTool(generateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("GENERATE_CODE_CONTEXT tool called")

		directoryPath, err := request.RequireString("directory_path")
		if err != nil {
			log.Printf("GENERATE_CODE_CONTEXT tool error - invalid parameter 'directory_path': %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Default to current directory if empty
		if directoryPath == "" {
			directoryPath = "."
		}

		// Get output file path
		outputFile := request.GetString("output_file", "")
		if outputFile == "" {
			outputFile = filepath.Join(directoryPath, "codebrev.md")
		}

		log.Printf("Generating code context for directory: %s, output: %s", directoryPath, outputFile)

		// Check if directory exists
		if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
			log.Printf("GENERATE_CODE_CONTEXT tool error - directory does not exist: %s", directoryPath)
			return mcp.NewToolResultError(fmt.Sprintf("directory does not exist: %s", directoryPath)), nil
		}

		// Generate the code context
		err = generateCodeContext(directoryPath, outputFile)
		if err != nil {
			log.Printf("GENERATE_CODE_CONTEXT tool error - failed to generate code context: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to generate code context: %v", err)), nil
		}

		// Get file size for reporting
		fileInfo, err := os.Stat(outputFile)
		var fileSize int64
		if err == nil {
			fileSize = fileInfo.Size()
		}

		log.Printf("GENERATE_CODE_CONTEXT tool: Successfully generated code context at %s (%d bytes)", outputFile, fileSize)
		return mcp.NewToolResultText(fmt.Sprintf("Successfully generated code context outline at %s\nSize: %d bytes", outputFile, fileSize)), nil
	})
}

// addGetCodeContextTool adds the tool to get cached codebrev.md or generate if missing
func addGetCodeContextTool(s *server.MCPServer) {
	getTool := mcp.NewTool("get_code_context",
		mcp.WithDescription("Get cached codebrev.md file content, or generate it if it doesn't exist (cache pattern)"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to the directory to get context for (defaults to current directory if not specified)"),
		),
		mcp.WithString("cache_file",
			mcp.Description("Path to the cache file (defaults to 'codebrev.md' in the target directory)"),
		),
		mcp.WithBoolean("force_regenerate",
			mcp.Description("Force regeneration even if cache file exists (defaults to false)"),
		),
	)

	s.AddTool(getTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("GET_CODE_CONTEXT tool called")

		directoryPath, err := request.RequireString("directory_path")
		if err != nil {
			log.Printf("GET_CODE_CONTEXT tool error - invalid parameter 'directory_path': %v", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Default to current directory if empty
		if directoryPath == "" {
			directoryPath = "."
		}

		// Get cache file path
		cacheFile := request.GetString("cache_file", "")
		if cacheFile == "" {
			cacheFile = filepath.Join(directoryPath, "codebrev.md")
		}

		// Get force regenerate flag
		forceRegenerate := request.GetBool("force_regenerate", false)

		log.Printf("Getting code context for directory: %s, cache: %s, force: %v", directoryPath, cacheFile, forceRegenerate)

		// Check if directory exists
		if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
			log.Printf("GET_CODE_CONTEXT tool error - directory does not exist: %s", directoryPath)
			return mcp.NewToolResultError(fmt.Sprintf("directory does not exist: %s", directoryPath)), nil
		}

		// Check if cache file exists and we're not forcing regeneration
		var needsGeneration bool
		if forceRegenerate {
			needsGeneration = true
			log.Printf("Force regeneration requested")
		} else {
			if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
				needsGeneration = true
				log.Printf("Cache file does not exist, will generate: %s", cacheFile)
			} else {
				log.Printf("Cache file exists: %s", cacheFile)
			}
		}

		// Generate if needed
		if needsGeneration {
			log.Printf("Generating code context...")
			err = generateCodeContext(directoryPath, cacheFile)
			if err != nil {
				log.Printf("GET_CODE_CONTEXT tool error - failed to generate code context: %v", err)
				return mcp.NewToolResultError(fmt.Sprintf("failed to generate code context: %v", err)), nil
			}
			log.Printf("Code context generated successfully")
		}

		// Read the cache file
		content, err := os.ReadFile(cacheFile)
		if err != nil {
			log.Printf("GET_CODE_CONTEXT tool error - failed to read cache file: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to read cache file: %v", err)), nil
		}

		status := "cached"
		if needsGeneration {
			status = "generated"
		}

		log.Printf("GET_CODE_CONTEXT tool: Successfully returned code context (%s, %d bytes)", status, len(content))
		return mcp.NewToolResultText(fmt.Sprintf("# Code Context (%s)\n\n%s", status, string(content))), nil
	})
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
