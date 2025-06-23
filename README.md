# CLSP - Command Line LSP Client

A command line tool for sending requests to Language Server Protocol (LSP) servers and displaying the results.

## Installation

```bash
go build -o clsp
```

## Usage

```bash
clsp -server <command> -method <method> [-args <args>] [-params <json>] [-root <uri>] [-skip-init]
```

### Flags

- `-server`: LSP server command to run (required)
- `-method`: LSP method to call (required) 
- `-args`: Comma-separated arguments for the LSP server
- `-params`: JSON parameters for the method (default: "{}")
- `-root`: Root URI for workspace initialization (default: current directory)
- `-skip-init`: Skip the initialize/initialized sequence for raw requests

### Examples

#### Go Language Server (gopls)

```bash
# Get hover information
./clsp -server gopls -method textDocument/hover \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":10,"character":5}}'

# Get completions
./clsp -server gopls -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.go"},"position":{"line":5,"character":10}}'

# Search workspace symbols
./clsp -server gopls -method workspace/symbol \
  -params '{"query":"main"}'
```

#### C/C++ Language Server (clangd)

```bash
# Get hover information
./clsp -server clangd -method textDocument/hover \
  -params '{"textDocument":{"uri":"file:///path/to/file.c"},"position":{"line":10,"character":5}}'

# Get completions
./clsp -server clangd -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.c"},"position":{"line":5,"character":10}}'
```

#### Python Language Server (pylsp)

```bash
# Get completions
./clsp -server pylsp -method textDocument/completion \
  -params '{"textDocument":{"uri":"file:///path/to/file.py"},"position":{"line":5,"character":10}}'
```

## Common LSP Methods

- `initialize` - Initialize the language server (automatically called unless -skip-init is used)
- `textDocument/hover` - Get hover information at a position
- `textDocument/completion` - Get code completions at a position
- `textDocument/definition` - Go to definition
- `textDocument/references` - Find references
- `textDocument/documentSymbol` - Get document symbols
- `workspace/symbol` - Search workspace symbols
- `textDocument/formatting` - Format document
- `textDocument/codeAction` - Get available code actions

## JSON-RPC Protocol

The tool implements the JSON-RPC 2.0 protocol over stdin/stdout as required by the LSP specification. It handles:

- Content-Length headers
- JSON-RPC request/response format
- LSP initialization sequence
- Proper cleanup on exit