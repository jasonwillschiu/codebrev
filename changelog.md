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
