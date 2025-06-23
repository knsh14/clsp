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

### Examples

#### Go Language Server (gopls)

```bash
# Get hover information
./clsp -server gopls -method textDocument/hover \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":10,"character":5}}'

# Get completions with JSON output
./clsp -server gopls -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":5,"character":10}}' \
  -format json

# Search workspace symbols (quiet output)
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}' -quiet

# Use parameters from file
echo '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":10,"character":5}}' > hover.json
./clsp -server gopls -method textDocument/hover -params-file hover.json
```

#### C/C++ Language Server (clangd)

```bash
# Get hover information with verbose logging
./clsp -server clangd -method textDocument/hover \
  -params '{"textDocument":{"uri":"file:///path/to/file.c"},"position":{"line":10,"character":5}}' \
  -verbose

# Get completions with timeout
./clsp -server clangd -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.c"},"position":{"line":5,"character":10}}' \
  -timeout 10s
```

#### Python Language Server (pylsp)

```bash
# Get completions (raw output format)
./clsp -server pylsp -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.py"},"position":{"line":5,"character":10}}' \
  -format raw

# List available methods
./clsp -list-methods
```

## Common LSP Methods

**Text Document Methods:**
- `textDocument/hover` - Get hover information at a position
- `textDocument/completion` - Get code completions at a position
- `textDocument/definition` - Go to definition
- `textDocument/references` - Find references
- `textDocument/documentSymbol` - Get document symbols
- `textDocument/formatting` - Format document
- `textDocument/codeAction` - Get available code actions
- `textDocument/rename` - Rename symbol

**Workspace Methods:**
- `workspace/symbol` - Search workspace symbols
- `workspace/executeCommand` - Execute workspace commands

**Diagnostic Methods:**
- `textDocument/publishDiagnostics` - Receive diagnostics (notification)

Use `./clsp -list-methods` to see the full list with descriptions.

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