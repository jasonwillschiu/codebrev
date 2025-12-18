# code4context Installation

This directory contains the code4context binary, a tool for generating 
structured code summaries for LLM consumption.

## Important Notes

- The code4context binary (~10MB) is typically added to .gitignore files
  to keep it out of version control repositories
- This helps maintain clean repositories and faster clone/push operations
- If you're using a different version control system, consider excluding
  this binary using your VCS's ignore mechanisms:
  - Git: Add 'code4context' to .gitignore
  - Mercurial: Add 'code4context' to .hgignore  
  - Subversion: Use svn:ignore property

## Usage

Run './code4context' in any directory to generate a codebrev.md file
containing a structured overview of your codebase.

For more information, visit: https://github.com/jasonwillschiu/code4context-com
