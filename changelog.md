# 0.4.9 - Update: cicd.js prettier install instructions
- added install instructions for claude code, opencode and cursor
- enhanced code4context-readme.txt when installed
- binary reuse is sym link instead of copy now

# 0.4.8 - Update: cicd.js binary reuse 5
- hasGoFileChanges() function - Checks git status --porcelain for any .go file modifications
- Early exit in build logic - If no .go files changed, skip building entirely
- skipBuild parameter - Pass this flag to uploadToR2() to force reuse path
- Smart reuse logic - When skipBuild=true, always use S3 copy from previous version

# 0.4.7 - Update: cicd.js binary reuse 4
- Single storage location: Only /releases/v{version}/ - eliminated /binaries/ folder entirely
- Content-based hashing: Calculate hash of all Go source files (not per-platform)
- Smart build skipping: If content hash matches previous version, skip building entirely
- Fast S3 copies: When reusing, copy binaries directly in R2 (no local download/upload)

# 0.4.6 - Update: cicd.js binary reuse, again
- modified buildCrossPlatform(), checking content hash before building

# 0.4.5 - Update: cicd.js binary reuse
- now looks at all *.go files to see if binary will be reused
- modified names of mcp servers to code4context and docs4context

# 0.4.4 - Update: cicd.js binary reuse
- previously binary reuse wasn't implemented

# 0.4.3 - Fix: install.sh
- removed "-e" prefix
- binary reuse if it didn't change

# 0.4.2 - Update: install.sh interactive TTY
- uses TTY redirection to make piped installation interactive
- cicd.js fix for --help error

# 0.4.1 - Add: install script and cicd.js
- cicd.js to automate git operations, build and deploy
- added install.sh for curl installation command
- testing storing binaries in cloudflare r2
- locally awscli is a dependency for r2

# 0.4.0 - Update: refactor to mcp server
- uses mark3labs/mcp-go now
- todo: install and update for the mcp server
- with installation, add a README and add to gitignore, see if i can use an interactive installer

# 0.3.1 - Fix: only function files now have output
- previously if a file only had a function, it would not output anything
- also updated README and AGENTS.md

# 0.3.0 - Update: refactor for better schema, removed treesitter
- focus on imports, functions and types
- where functions and types have input/output and schema where possible
- remove variables and constants as that's not useful
- remove treesitter, it didn't add value for the complexity
- the goal is that the output file gives an llm-useful-summary of the codebase, where an LLM can view the summary and not miss files to edit and understand the existing functions and types

# 0.2.0 - Update: refactor to use treesitter algorithm for most
- .go, .js, .ts and .tsx now parsed with treesitter
- .astro still uses custom regex parsing
- a working state, still no mermaid diagrams

# 0.1.3 - Update: refactor to use /internal folder
- packages are: gitignore, outline, parser and writer
- tested on go files, .astro, .js and .tsx files

# 0.1.2 - Add: .astro and better .tsx file support
- improved outputs for .tsx files
- .astro files now show something useful

# 0.1.1 - Update: ignore multiple .gitignore files
- previously wasn't respecting multiple .gitignore files

# 0.1.0 - Add: initial commit
- creates a codebrev.md file which has a summary of the codebase
- walks all files and folders for .go and .js files
- reasonable output but can be improved
