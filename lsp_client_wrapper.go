package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LSPCompletionItem represents a completion item from LSP
type LSPCompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// LSPClientWrapper manages communication with gopls
type LSPClientWrapper struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	ready   bool
	mu      sync.RWMutex
	msgID   int
	pending map[int]chan *LSPResponse
}

// LSPRequest represents a JSON-RPC request
type LSPRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// LSPResponse represents a JSON-RPC response
type LSPResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *LSPError   `json:"error,omitempty"`
}

// LSPError represents an LSP error
type LSPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CompletionParams represents parameters for textDocument/completion
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position                `json:"position"`
}

// TextDocumentIdentifier identifies a text document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

// CompletionList represents the result of textDocument/completion
type CompletionList struct {
	IsIncomplete bool                `json:"isIncomplete"`
	Items        []LSPCompletionItem `json:"items"`
}

// NewLSPClientWrapper creates a new LSP client wrapper
func NewLSPClientWrapper() (*LSPClientWrapper, error) {
	fmt.Fprintf(os.Stderr, "🚀 [LSP] Starting gopls...\n")
	
	cmd := exec.Command("gopls", "serve")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	
	wrapper := &LSPClientWrapper{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		msgID:   1,
		pending: make(map[int]chan *LSPResponse),
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start gopls: %v", err)
	}
	
	// Start message reader
	go wrapper.readMessages()
	
	// Start stderr reader for debugging
	go wrapper.readStderr()
	
	// Send initialization
	if err := wrapper.initialize(); err != nil {
		wrapper.Shutdown()
		return nil, fmt.Errorf("failed to initialize gopls: %v", err)
	}
	
	wrapper.mu.Lock()
	wrapper.ready = true
	wrapper.mu.Unlock()
	
	fmt.Fprintf(os.Stderr, "✅ [LSP] gopls initialized successfully\n")
	return wrapper, nil
}

// IsReady returns true if the LSP client is ready
func (l *LSPClientWrapper) IsReady() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.ready
}

// GetCompletions gets completions from gopls for the given line and position
func (l *LSPClientWrapper) GetCompletions(line string, pos int) ([]LSPCompletionItem, error) {
	fmt.Fprintf(os.Stderr, "🎯 [LSP] Getting completions for line: %q, pos: %d\n", line, pos)
	
	// Get current working directory to use as file path
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}
	
	// Update the existing file
	tempFile := wd + "/repl_session.go"
	content := `package main

import "fmt"

func main() {
` + line + `
}`
	
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to update temp file: %v", err)
	}
	
	// Send didChange to update the document
	didChangeParams := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        "file://" + tempFile,
			"version":    2,
		},
		"contentChanges": []map[string]interface{}{
			{
				"text": content,
			},
		},
	}
	
	if err := l.sendMessage(LSPRequest{
		JsonRPC: "2.0",
		Method:  "textDocument/didChange",
		Params:  didChangeParams,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to send changes: %v\n", err)
	}
	
	// Give gopls a moment to process
	time.Sleep(50 * time.Millisecond)
	
	// Calculate position (in the main function, at the pos position)
	cursorPos := Position{
		Line:      5, // 0-based line number (6th line in file)
		Character: pos, // Position within that line
	}
	
	fmt.Fprintf(os.Stderr, "📍 [LSP] Cursor position: line %d, char %d\n", cursorPos.Line, cursorPos.Character)
	
	// Send completion request
	params := CompletionParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file://" + tempFile,
		},
		Position: cursorPos,
	}
	
	items, err := l.call("textDocument/completion", params)
	if err != nil {
		return nil, fmt.Errorf("completion request failed: %v", err)
	}
	
	return items, nil
}

// call sends a request and waits for the response
func (l *LSPClientWrapper) call(method string, params interface{}) ([]LSPCompletionItem, error) {
	l.mu.Lock()
	id := l.msgID
	l.msgID++
	l.mu.Unlock()
	
	// Create response channel
	responseChan := make(chan *LSPResponse, 1)
	
	l.mu.Lock()
	l.pending[id] = responseChan
	l.mu.Unlock()
	
	// Send request
	request := LSPRequest{
		JsonRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	
	if err := l.sendMessage(request); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	
	fmt.Fprintf(os.Stderr, "📤 [LSP] Sent request %d: %s\n", id, method)
	
	// Wait for response with timeout
	select {
	case response := <-responseChan:
		fmt.Fprintf(os.Stderr, "📥 [LSP] Received response %d\n", id)
		if response.Error != nil {
			return nil, fmt.Errorf("LSP error: %s", response.Error.Message)
		}
		
		// Parse completion list
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %v", err)
		}
		
		var completionList CompletionList
		if err := json.Unmarshal(resultBytes, &completionList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal completion list: %v", err)
		}
		
		fmt.Fprintf(os.Stderr, "✅ [LSP] Parsed %d completion items\n", len(completionList.Items))
		return completionList.Items, nil
		
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// sendMessage sends a JSON-RPC message to gopls
func (l *LSPClientWrapper) sendMessage(request LSPRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	
	// Send Content-Length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := io.WriteString(l.stdin, header); err != nil {
		return err
	}
	
	// Send message content
	if _, err := l.stdin.Write(data); err != nil {
		return err
	}
	
	fmt.Fprintf(os.Stderr, "📨 [LSP] Sent message: %s\n", string(data))
	return nil
}

// readMessages reads responses from gopls in a goroutine
func (l *LSPClientWrapper) readMessages() {
	reader := bufio.NewReader(l.stdout)
	
	for {
		// Read Content-Length header
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "❌ [LSP] Error reading header: %v\n", err)
			}
			break
		}
		
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			length, err := strconv.Atoi(lengthStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ [LSP] Invalid content length: %v\n", err)
				continue
			}
			
			// Read the blank line
			if _, err := reader.ReadString('\n'); err != nil {
				fmt.Fprintf(os.Stderr, "❌ [LSP] Error reading blank line: %v\n", err)
				continue
			}
			
			// Read the message content
			content := make([]byte, length)
			bytesRead := 0
			for bytesRead < length {
				n, err := reader.Read(content[bytesRead:])
				if err != nil {
					fmt.Fprintf(os.Stderr, "❌ [LSP] Error reading content: %v\n", err)
					break
				}
				bytesRead += n
			}
			
			if bytesRead == length {
				l.handleResponse(content)
			}
		}
	}
}

// handleResponse processes an incoming response
func (l *LSPClientWrapper) handleResponse(data []byte) {
	fmt.Fprintf(os.Stderr, "🟢 [LSP] Received response: %s\n", string(data))
	
	// Try to parse as a response first
	var response LSPResponse
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Fprintf(os.Stderr, "❌ [LSP] Failed to parse response: %v\n", err)
		return
	}
	
	// Check if this is a notification (no ID) or a response (has ID)
	if response.ID == 0 {
		// This is a notification (like window/showMessage), ignore it for now
		fmt.Fprintf(os.Stderr, "🔔 [LSP] Ignoring notification (ID=0)\n")
		return
	}
	
	// Find the waiting channel for this response ID
	l.mu.Lock()
	if responseChan, exists := l.pending[response.ID]; exists {
		delete(l.pending, response.ID)
		l.mu.Unlock()
		
		// Send response to the waiting goroutine
		select {
		case responseChan <- &response:
			fmt.Fprintf(os.Stderr, "✅ [LSP] Delivered response %d to caller\n", response.ID)
		default:
			fmt.Fprintf(os.Stderr, "❌ [LSP] Response channel blocked for ID %d\n", response.ID)
		}
	} else {
		l.mu.Unlock()
		fmt.Fprintf(os.Stderr, "🔍 [LSP] No waiting channel for response ID %d\n", response.ID)
	}
}

// readStderr reads error messages from gopls for debugging
func (l *LSPClientWrapper) readStderr() {
	scanner := bufio.NewScanner(l.stderr)
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, "🚨 [LSP] STDERR: %s\n", scanner.Text())
	}
}

// initialize sends the initialize request to gopls  
func (l *LSPClientWrapper) initialize() error {
	// Get current working directory 
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	
	// Create AND OPEN a valid Go document BEFORE sending initialized
	tempFile := wd + "/repl_session.go"
	initialContent := `package main

import "fmt"

func main() {
	// REPL context placeholder
}`
	
	if err := os.WriteFile(tempFile, []byte(initialContent), 0644); err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}

	initParams := map[string]interface{}{
		"processId": 12345,
		"rootUri":   "file://" + wd,
		"workspaceFolders": []map[string]interface{}{
			{
				"uri":  "file://" + wd,
				"name": "gosh-workspace",
			},
		},
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"completion": map[string]interface{}{
					"completionItem": map[string]interface{}{
						"snippetSupport": false,
					},
				},
			},
		},
	}
	
	response, err := l.call("initialize", initParams)
	if err != nil {
		return fmt.Errorf("initialize failed: %v", err)
	}
	
	fmt.Fprintf(os.Stderr, "✅ [LSP] Initialize response: %+v\n", response)
	
	// IMPORTANT: Open the document BEFORE sending initialized notification
	didOpenParams := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        "file://" + tempFile,
			"languageId": "go",
			"version":    1,
		},
	}
	
	if err := l.sendMessage(LSPRequest{
		JsonRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  didOpenParams,
	}); err != nil {
		return fmt.Errorf("failed to open document: %v", err)
	}
	
	// Give gopls time to process the open
	time.Sleep(100 * time.Millisecond)
	
	// NOW send initialized notification
	notif := LSPRequest{
		JsonRPC: "2.0",
		Method:  "initialized",
	}
	
	if err := l.sendMessage(notif); err != nil {
		return fmt.Errorf("failed to send initialized notification: %v", err)
	}
	
	// Give gopls time to learn about the view
	time.Sleep(200 * time.Millisecond)
	
	return nil
}

// updateVirtualFile updates the temporary file with current REPL line
func (l *LSPClientWrapper) updateVirtualFile(line string) error {
	tempFile := "/tmp/repl_workspace/session.go"
	
	// Create realistic Go context for gopls
	content := `package main

import "fmt"

func session() {
` + line + `
}`
	
	return os.WriteFile(tempFile, []byte(content), 0644)
}

// Shutdown closes the LSP client
func (l *LSPClientWrapper) Shutdown() error {
	l.mu.Lock()
	l.ready = false
	l.mu.Unlock()
	
	if l.cmd != nil && l.cmd.Process != nil {
		// Send shutdown request
		if err := l.sendMessage(LSPRequest{
			JsonRPC: "2.0",
			ID:      l.msgID,
			Method:  "shutdown",
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to send shutdown: %v\n", err)
		}
		
		// Send exit notification
		l.sendMessage(LSPRequest{
			JsonRPC: "2.0",
			Method:  "exit",
		})
		
		// Wait for process to exit
		if err := l.cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: gopls shutdown error: %v\n", err)
		}
	}
	
	return nil
}

// ConvertLSPCompletions converts LSP completion items to our format
func ConvertLSPCompletions(lspItems []LSPCompletionItem) []CompletionItem {
	var suggestions []CompletionItem
	
	for _, item := range lspItems {
		suggestion := CompletionItem{
			Label:  item.Label,
			Kind:   "function", // Simplified - could map LSP kinds to our kinds
			Detail: item.Detail,
		}
		suggestions = append(suggestions, suggestion)
	}
	
	return suggestions
}
