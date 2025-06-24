package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      *int   `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type InitializeParams struct {
	ProcessID    int            `json:"processId"`
	RootURI      string         `json:"rootUri"`
	Capabilities map[string]any `json:"capabilities"`
}

type LSPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	reader *bufio.Reader
	id     int
	logger *slog.Logger
}

func NewLSPClient(ctx context.Context, command string, args []string, logger *slog.Logger) (*LSPClient, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &LSPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		reader: bufio.NewReader(stdout),
		id:     1,
		logger: logger,
	}, nil
}

func (c *LSPClient) SendRequest(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	id := c.id
	c.id++
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
		Params:  params,
	}

	c.logger.Debug("Sending LSP request", "method", method, "id", *request.ID)

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	content := string(requestBytes)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))

	if _, err := c.stdin.Write([]byte(header + content)); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	return c.ReadResponse(ctx, *request.ID)
}

func (c *LSPClient) SendNotification(method string, params any) error {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	c.logger.Debug("Sending LSP notification", "method", method)

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	content := string(requestBytes)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))

	if _, err := c.stdin.Write([]byte(header + content)); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

func (c *LSPClient) ReadResponse(ctx context.Context, expectedID int) (*JSONRPCResponse, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Read one message
		var contentLength int
		for {
			line, err := c.reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("failed to read header line: %w", err)
			}

			line = strings.TrimSpace(line)
			if line == "" {
				break
			}

			if s, ok := strings.CutPrefix(line, "Content-Length:"); ok {
				lengthStr := strings.TrimSpace(s)
				contentLength, err = strconv.Atoi(lengthStr)
				if err != nil {
					return nil, fmt.Errorf("invalid Content-Length header: %w", err)
				}
			}
		}

		if contentLength == 0 {
			return nil, errors.New("no Content-Length header found")
		}

		content := make([]byte, contentLength)
		_, err := io.ReadFull(c.reader, content)
		if err != nil {
			return nil, fmt.Errorf("failed to read response content: %w", err)
		}

		var response JSONRPCResponse
		if err := json.Unmarshal(content, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		c.logger.Debug("Received LSP message", "id", response.ID, "hasResult", response.Result != nil, "hasError", response.Error != nil, "expectedID", expectedID)

		// Check if this is a notification (ID = 0 and has Method field)
		var notification struct {
			Method string `json:"method"`
		}
		json.Unmarshal(content, &notification)

		if notification.Method != "" {
			c.logger.Debug("Received LSP notification", "method", notification.Method)
			continue // Skip notifications and keep reading
		}

		// Check if this is the response we're waiting for
		if response.ID == expectedID {
			return &response, nil
		}

		c.logger.Debug("Received unexpected response ID, continuing to read", "received", response.ID, "expected", expectedID)
	}
}

func (c *LSPClient) Initialize(ctx context.Context, rootURI string) error {
	params := InitializeParams{
		ProcessID: os.Getpid(),
		RootURI:   rootURI,
		Capabilities: map[string]any{
			"textDocument": map[string]any{
				"completion": map[string]any{
					"completionItem": map[string]any{
						"snippetSupport": true,
					},
				},
				"hover": map[string]any{
					"contentFormat": []string{"markdown", "plaintext"},
				},
				"documentSymbol":  map[string]any{},
				"workspaceSymbol": map[string]any{},
			},
			"workspace": map[string]any{
				"symbol": map[string]any{},
			},
		},
	}

	response, err := c.SendRequest(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("LSP initialize error [%d]: %s", response.Error.Code, response.Error.Message)
	}

	// Send initialized notification (no response expected)
	if err := c.SendNotification("initialized", map[string]any{}); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	return nil
}

func (c *LSPClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.SendRequest(ctx, "shutdown", nil)
	c.SendNotification("exit", nil)

	if err := c.stdin.Close(); err != nil {
		c.logger.Warn("Failed to close stdin", "error", err)
	}
	if err := c.stdout.Close(); err != nil {
		c.logger.Warn("Failed to close stdout", "error", err)
	}
	if err := c.stderr.Close(); err != nil {
		c.logger.Warn("Failed to close stderr", "error", err)
	}

	return c.cmd.Wait()
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printResponse(method string, response *JSONRPCResponse, format string, quiet bool) {
	switch format {
	case "json":
		data, _ := json.Marshal(response)
		fmt.Println(string(data))
	case "raw":
		if response.Result != nil {
			data, _ := json.Marshal(response.Result)
			fmt.Println(string(data))
		} else if response.Error != nil {
			data, _ := json.Marshal(response.Error)
			fmt.Println(string(data))
		}
	default: // pretty
		if quiet {
			if response.Result != nil {
				printJSON(response.Result)
			} else if response.Error != nil {
				printJSON(response.Error)
			}
		} else {
			fmt.Printf("Response for %s:\n", method)
			printJSON(response)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: clsp -server <command> -method <method> [options]")
	fmt.Println("\nRequired:")
	fmt.Println("  -server <cmd>     LSP server command (e.g., gopls, clangd, pylsp)")
	fmt.Println("  -method <method>  LSP method to call")
	fmt.Println("\nOptions:")
	fmt.Println("  -args <args>         Server arguments (comma-separated)")
	fmt.Println("  -params <json>       JSON parameters for the method")
	fmt.Println("  -params-file <file>  Read parameters from JSON file")
	fmt.Println("  -root <uri>          Root URI for initialization")
	fmt.Println("  -skip-init           Skip LSP initialization")
	fmt.Println("  -timeout <duration>  Request timeout (default: 30s)")
	fmt.Println("  -format <fmt>        Output format: pretty, json, raw (default: pretty)")
	fmt.Println("  -quiet               Only output result data")
	fmt.Println("  -verbose             Enable verbose logging")
	fmt.Println("  -list-methods        List common LSP methods")
	fmt.Println("\nExamples:")
	fmt.Println("  # Hover information")
	fmt.Println("  clsp -server gopls -method textDocument/hover -params '{\"textDocument\":{\"uri\":\"file:///path/to/file.go\"},\"position\":{\"line\":10,\"character\":5}}'")
	fmt.Println("  # Use params file")
	fmt.Println("  clsp -server gopls -method textDocument/completion -params-file hover.json")
	fmt.Println("  # List workspace symbols")
	fmt.Println("  clsp -server gopls -method workspace/symbol -params '{\"query\":\"main\"}' -format json -quiet")
}

func printCommonMethods() {
	fmt.Println("Common LSP Methods:")
	fmt.Println("\nText Document:")
	fmt.Println("  textDocument/hover           - Get hover information")
	fmt.Println("  textDocument/completion      - Get code completion")
	fmt.Println("  textDocument/definition      - Go to definition")
	fmt.Println("  textDocument/references      - Find references")
	fmt.Println("  textDocument/documentSymbol  - Get document symbols")
	fmt.Println("  textDocument/formatting      - Format document")
	fmt.Println("  textDocument/codeAction      - Get code actions")
	fmt.Println("  textDocument/rename          - Rename symbol")
	fmt.Println("\nWorkspace:")
	fmt.Println("  workspace/symbol             - Find workspace symbols")
	fmt.Println("  workspace/executeCommand     - Execute command")
	fmt.Println("\nDiagnostics:")
	fmt.Println("  textDocument/publishDiagnostics - Diagnostics (notification)")
	fmt.Println("\nExample parameter files can be created with:")
	fmt.Println("  echo '{\"textDocument\":{\"uri\":\"file:///path/to/file.go\"},\"position\":{\"line\":10,\"character\":5}}' > hover.json")
}

func main() {
	var (
		serverCmd    = flag.String("server", "", "LSP server command (required)")
		serverArgs   = flag.String("args", "", "LSP server arguments (comma-separated)")
		method       = flag.String("method", "", "LSP method to call (required)")
		paramsStr    = flag.String("params", "{}", "JSON parameters for the method")
		paramsFile   = flag.String("params-file", "", "Read parameters from JSON file")
		rootURI      = flag.String("root", "", "Root URI for initialization (defaults to current directory)")
		skipInit     = flag.Bool("skip-init", false, "Skip initialization (for raw requests)")
		timeout      = flag.Duration("timeout", 30*time.Second, "Request timeout")
		verbose      = flag.Bool("verbose", false, "Enable verbose logging")
		outputFormat = flag.String("format", "pretty", "Output format: pretty, json, raw")
		quiet        = flag.Bool("quiet", false, "Only output result data, no headers or labels")
		listMethods  = flag.Bool("list-methods", false, "List common LSP methods and exit")
	)
	flag.Parse()

	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	if *listMethods {
		printCommonMethods()
		os.Exit(0)
	}

	if *serverCmd == "" || *method == "" {
		printUsage()
		os.Exit(1)
	}

	var args []string
	if *serverArgs != "" {
		args = strings.Split(*serverArgs, ",")
		for i, arg := range args {
			args[i] = strings.TrimSpace(arg)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client, err := NewLSPClient(ctx, *serverCmd, args, logger)
	if err != nil {
		logger.Error("Failed to start LSP server", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Warn("Failed to close LSP client", "error", closeErr)
		}
	}()

	if !*skipInit {
		rootURIValue := *rootURI
		if rootURIValue == "" {
			pwd, _ := os.Getwd()
			rootURIValue = "file://" + pwd
		}

		if err := client.Initialize(ctx, rootURIValue); err != nil {
			logger.Error("Failed to initialize LSP server", "error", err)
			os.Exit(1)
		}
	}

	var params any
	if *paramsFile != "" {
		paramsData, err := os.ReadFile(*paramsFile)
		if err != nil {
			logger.Error("Failed to read params file", "file", *paramsFile, "error", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(paramsData, &params); err != nil {
			logger.Error("Failed to parse params file JSON", "file", *paramsFile, "error", err)
			os.Exit(1)
		}
	} else if *paramsStr != "" && *paramsStr != "{}" {
		if err := json.Unmarshal([]byte(*paramsStr), &params); err != nil {
			logger.Error("Failed to parse params JSON", "error", err)
			os.Exit(1)
		}
	}

	response, err := client.SendRequest(ctx, *method, params)
	if err != nil {
		logger.Error("Failed to send request", "method", *method, "error", err)
		os.Exit(1)
	}

	printResponse(*method, response, *outputFormat, *quiet)
}
