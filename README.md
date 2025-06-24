# CLSP - Command Line LSP Client

A command line tool for sending requests to Language Server Protocol (LSP) servers and displaying the results.

## Installation

```bash
go build -o clsp
```

## Usage

```bash
clsp -server <command> -method <method> [options]
```

### Flags

**Required:**
- `-server <cmd>`: LSP server command to run (e.g., gopls, clangd, pylsp)
- `-method <method>`: LSP method to call

**Options:**
- `-args <args>`: Comma-separated arguments for the LSP server
- `-params <json>`: JSON parameters for the method (default: "{}")
- `-params-file <file>`: Read parameters from JSON file instead of command line
- `-root <uri>`: Root URI for workspace initialization (default: current directory)
- `-skip-init`: Skip the initialize/initialized sequence for raw requests
- `-timeout <duration>`: Request timeout (default: 30s)
- `-format <fmt>`: Output format: pretty, json, raw (default: pretty)
- `-quiet`: Only output result data, no headers or labels
- `-verbose`: Enable verbose logging to stderr
- `-list-methods`: List common LSP methods and exit

### LSP Methods Examples with gopls

All examples use `gopls` as the LSP server. Replace file paths with your actual Go files.

#### Text Document Information

**Get hover information:**
```bash
./clsp -server gopls -method textDocument/hover \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":10,"character":5}}'
```

**Get signature help:**
```bash
./clsp -server gopls -method textDocument/signatureHelp \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":20,"character":15}}'
```

#### Code Completion

**Get code completions:**
```bash
./clsp -server gopls -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":5,"character":10},"context":{"triggerKind":1}}'
```

#### Navigation

**Go to definition:**
```bash
./clsp -server gopls -method textDocument/definition \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8}}'
```

**Go to type definition:**
```bash
./clsp -server gopls -method textDocument/typeDefinition \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8}}'
```

**Find implementations:**
```bash
./clsp -server gopls -method textDocument/implementation \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8}}'
```

**Find all references:**
```bash
./clsp -server gopls -method textDocument/references \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8},"context":{"includeDeclaration":true}}'
```

#### Document Structure

**Get document symbols:**
```bash
./clsp -server gopls -method textDocument/documentSymbol \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"}}'
```

**Highlight symbol occurrences:**
```bash
./clsp -server gopls -method textDocument/documentHighlight \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8}}'
```

#### Code Formatting

**Format entire document:**
```bash
./clsp -server gopls -method textDocument/formatting \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"options":{"tabSize":4,"insertSpaces":false}}'
```

**Format specific range:**
```bash
./clsp -server gopls -method textDocument/rangeFormatting \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"range":{"start":{"line":10,"character":0},"end":{"line":20,"character":0}},"options":{"tabSize":4,"insertSpaces":false}}'
```

#### Code Actions and Refactoring

**Get available code actions:**
```bash
./clsp -server gopls -method textDocument/codeAction \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"range":{"start":{"line":10,"character":0},"end":{"line":10,"character":50}},"context":{"diagnostics":[]}}'
```

**Rename symbol:**
```bash
./clsp -server gopls -method textDocument/rename \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8},"newName":"NewFunctionName"}'
```

**Check if symbol can be renamed:**
```bash
./clsp -server gopls -method textDocument/prepareRename \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":15,"character":8}}'
```

#### Enhanced Features

**Get code lenses:**
```bash
./clsp -server gopls -method textDocument/codeLens \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"}}'
```

**Get inlay hints:**
```bash
./clsp -server gopls -method textDocument/inlayHint \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"range":{"start":{"line":10,"character":0},"end":{"line":30,"character":0}}}'
```

#### Workspace Operations

**Search workspace symbols:**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}'
```

**Execute workspace command:**
```bash
./clsp -server gopls -method workspace/executeCommand \
  -params '{"command":"gopls.tidy","arguments":[]}'
```

#### gopls-Specific Methods

**Get gopls debug info:**
```bash
./clsp -server gopls -method gopls/debug/info -params '{}'
```

**Get GC details:**
```bash
./clsp -server gopls -method gopls/gc_details \
  -params '{"uri":"file:///path/to/file.go"}'
```

**Run go generate:**
```bash
./clsp -server gopls -method gopls/generate \
  -params '{"uri":"file:///path/to/directory","recursive":false}'
```

**List available imports:**
```bash
./clsp -server gopls -method gopls/list_imports \
  -params '{"uri":"file:///path/to/file.go"}'
```

**Add import:**
```bash
./clsp -server gopls -method gopls/add_import \
  -params '{"uri":"file:///path/to/file.go","importPath":"fmt"}'
```

#### Output Format Examples

**JSON output:**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}' -format json
```

**Raw output (result only):**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}' -format raw
```

**Quiet mode:**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}' -quiet
```

#### Using Parameter Files

**Create and use parameter file:**
```bash
# Create parameter file
echo '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":10,"character":5}}' > hover.json

# Use parameter file
./clsp -server gopls -method textDocument/hover -params-file hover.json
```

#### Advanced Options

**Custom timeout:**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":".*"}' -timeout 60s
```

**Verbose logging:**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}' -verbose
```

**Skip initialization:**
```bash
./clsp -server gopls -method textDocument/hover \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":10,"character":5}}' \
  -skip-init
```

**Custom workspace root:**
```bash
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}' -root "file:///path/to/project"
```

## Complete LSP Methods Reference

**Text Document Methods:**
- `textDocument/hover` - Get hover information at a position
- `textDocument/completion` - Get code completions at a position
- `textDocument/signatureHelp` - Get signature help for function calls
- `textDocument/definition` - Go to definition
- `textDocument/typeDefinition` - Go to type definition
- `textDocument/implementation` - Find implementations
- `textDocument/references` - Find all references
- `textDocument/documentSymbol` - Get document symbols
- `textDocument/documentHighlight` - Highlight symbol occurrences
- `textDocument/formatting` - Format entire document
- `textDocument/rangeFormatting` - Format specific range
- `textDocument/codeAction` - Get available code actions
- `textDocument/codeLens` - Get code lenses
- `textDocument/rename` - Rename symbol
- `textDocument/prepareRename` - Check if symbol can be renamed
- `textDocument/inlayHint` - Get inlay hints

**Workspace Methods:**
- `workspace/symbol` - Search workspace symbols
- `workspace/executeCommand` - Execute workspace commands
- `workspace/didChangeConfiguration` - Notify configuration changes
- `workspace/didChangeWatchedFiles` - Notify file system changes

**gopls-Specific Methods:**
- `gopls/debug/info` - Get gopls debug information
- `gopls/gc_details` - Get garbage collection details
- `gopls/generate` - Run go generate
- `gopls/list_imports` - List available imports
- `gopls/add_import` - Add import to file
- `gopls/remove_dependency` - Remove unused dependency

Use `./clsp -list-methods` to see the built-in method list with descriptions.

## Implementation Details

### JSON-RPC Protocol

The tool implements the JSON-RPC 2.0 protocol over stdin/stdout as required by the LSP specification. It handles:

- **Content-Length Headers**: Proper LSP message framing with `Content-Length: N\r\n\r\n` headers
- **JSON-RPC Format**: Compliant request/response format with ID tracking
- **LSP Initialization**: Automatic `initialize` → `initialized` sequence (unless `-skip-init` is used)
- **Graceful Cleanup**: Sends `shutdown` → `exit` sequence before terminating the LSP server

### Output Formats

- **pretty** (default): Human-readable JSON with formatting and response metadata
- **json**: Raw JSON-RPC response including headers and error information
- **raw**: Only the `result` or `error` field content

### Error Handling

- Network and protocol errors are logged to stderr
- LSP server errors are included in the response output
- Proper timeout handling with configurable duration
- Clean process termination with signal handling

### Testing

The project includes comprehensive tests covering:
- JSON-RPC message serialization/deserialization
- Content-Length header parsing
- ID increment tracking
- LSP initialization parameter structure

Run tests with:
```bash
go test ./...
go test -v ./...  # verbose output
go test -cover ./...  # with coverage
```