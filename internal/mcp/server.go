package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Server implements the MCP protocol over stdin/stdout using JSON-RPC 2.0.
type Server struct {
	tools    map[string]ToolDefinition
	handlers map[string]ToolHandler
	reader   *bufio.Scanner
	writer   io.Writer
	logger   io.Writer
}

// NewServer creates a new MCP server that reads from reader, writes JSON-RPC responses
// to writer, and logs diagnostic messages to logger.
func NewServer(reader io.Reader, writer io.Writer, logger io.Writer) *Server {
	return &Server{
		tools:    make(map[string]ToolDefinition),
		handlers: make(map[string]ToolHandler),
		reader:   bufio.NewScanner(reader),
		writer:   writer,
		logger:   logger,
	}
}

// RegisterTool registers a tool definition and its handler with the server.
func (s *Server) RegisterTool(def ToolDefinition, handler ToolHandler) {
	s.tools[def.Name] = def
	s.handlers[def.Name] = handler
}

// Serve starts the main loop, reading JSON-RPC requests line by line from stdin
// and dispatching them to the appropriate handler.
func (s *Server) Serve() error {
	for s.reader.Scan() {
		line := s.reader.Bytes()

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			// Parse error - respond with error, ID=null
			s.sendResponse(&Response{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &RPCError{Code: ParseError, Message: "Parse error"},
			})
			continue
		}

		resp := s.handleRequest(&req)
		if resp != nil {
			s.sendResponse(resp)
		}
	}
	return s.reader.Err()
}

func (s *Server) handleRequest(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "racore-cli",
					"version": "0.1.0",
				},
			},
		}

	case "notifications/initialized":
		// Notification - no response needed
		return nil

	case "tools/list":
		toolList := make([]ToolDefinition, 0, len(s.tools))
		for _, t := range s.tools {
			toolList = append(toolList, t)
		}
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{"tools": toolList},
		}

	case "tools/call":
		// Parse params to get tool name and arguments
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: InvalidParams, Message: "Invalid params"},
			}
		}

		handler, ok := s.handlers[params.Name]
		if !ok {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: InvalidParams, Message: fmt.Sprintf("Unknown tool: %s", params.Name)},
			}
		}

		result, err := handler(params.Arguments)
		if err != nil {
			// Per MCP spec: tool execution errors are returned as result with isError=true
			// NOT as JSON-RPC errors (those are for protocol issues only)
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": err.Error()},
					},
					"isError": true,
				},
			}
		}

		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: MethodNotFound, Message: "Method not found"},
		}
	}
}

func (s *Server) sendResponse(resp *Response) error {
	if resp == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(s.logger, "Error: failed to marshal response: %v\n", err)
		return err
	}
	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	return err
}
