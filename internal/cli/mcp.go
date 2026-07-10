package cli

// eis mcp — a Model Context Protocol (MCP) stdio server that exposes OrbitLens's
// structural signals as agent-callable tools. Any MCP client (Cursor, Claude
// Code, Windsurf, …) can point at `eis mcp` and call a tool during its work.
//
// v0 exposes ONE tool: get_write_context — index-backed (reads the precomputed
// .eis/write-index.json, no analysis), so it's fast enough to call per write.
// It reuses buildWriteContext (the same core as `eis get-write-context` and the
// precheck hook), so the three surfaces can't drift; output is person-stripped.
//
// Transport: newline-delimited JSON-RPC 2.0 over stdin/stdout (MCP stdio).
//
// TODO(v1): anchors/graveyard tools once their signals are wired into
// write-index (they need index-backing to stay fast per-call). Hosted/team
// transport (SSE/HTTP) is a separate concern.

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
)

const mcpProtocolVersion = "2024-11-05"

func runMCP(args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	indexPath := fs.String("index", "", "write-index path (default <cwd>/.eis/write-index.json)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	s := &mcpServer{indexPath: *indexPath}
	return s.serve(os.Stdin, os.Stdout)
}

type mcpServer struct {
	indexPath string
}

type rpcMsg struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (s *mcpServer) serve(in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // tool results can be large
	enc := json.NewEncoder(out)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req rpcMsg
		if err := json.Unmarshal(line, &req); err != nil {
			continue // ignore malformed frames rather than crash
		}
		resp := s.handle(&req)
		if resp == nil {
			continue // notification (no id) → no response
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return sc.Err()
}

// handle returns the JSON-RPC response, or nil for notifications (no id).
func (s *mcpServer) handle(req *rpcMsg) map[string]any {
	isNotification := len(req.ID) == 0
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		pv := p.ProtocolVersion
		if pv == "" {
			pv = mcpProtocolVersion
		}
		return okResult(req.ID, map[string]any{
			"protocolVersion": pv,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "eis", "version": version},
		})
	case "notifications/initialized", "initialized", "notifications/cancelled":
		return nil
	case "ping":
		return okResult(req.ID, map[string]any{})
	case "tools/list":
		return okResult(req.ID, map[string]any{"tools": mcpTools()})
	case "tools/call":
		return s.toolsCall(req)
	default:
		if isNotification {
			return nil
		}
		return rpcError(req.ID, -32601, "method not found: "+req.Method)
	}
}

func (s *mcpServer) toolsCall(req *rpcMsg) map[string]any {
	var p struct {
		Name      string `json:"name"`
		Arguments struct {
			Paths     []string `json:"paths"`
			Repo      string   `json:"repo"`
			IndexPath string   `json:"index_path"`
		} `json:"arguments"`
	}
	_ = json.Unmarshal(req.Params, &p)
	if p.Name != "get_write_context" {
		return rpcError(req.ID, -32602, "unknown tool: "+p.Name)
	}

	cwd := p.Arguments.Repo
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	idxPath := p.Arguments.IndexPath
	if idxPath == "" {
		idxPath = s.indexPath
	}
	if idxPath == "" {
		idxPath = filepath.Join(cwd, defaultWriteIndexPath)
	}

	idx, _ := loadWriteIndex(idxPath) // fail-open: missing/broken index → empty
	wc := buildWriteContext(p.Arguments.Paths, cwd, idx)
	b, err := json.MarshalIndent(wc, "", "  ")
	if err != nil {
		return rpcError(req.ID, -32603, "encode error")
	}
	return okResult(req.ID, map[string]any{
		"content": []map[string]any{{"type": "text", "text": string(b)}},
		"isError": false,
	})
}

func mcpTools() []map[string]any {
	return []map[string]any{{
		"name": "get_write_context",
		"description": "Before writing or editing a file, get its OrbitLens structural-debt context: " +
			"whether the module is orphaned/dead/bus-factor-1, how long since it was owned/touched, " +
			"and directives on how to write there safely. Person-stripped (structural facts only). " +
			"Fast — reads the precomputed .eis/write-index.json (run `eis write-index` first).",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"paths": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Repo-relative file paths about to be written or edited.",
				},
				"repo":       map[string]any{"type": "string", "description": "Repo root (default: current working directory)."},
				"index_path": map[string]any{"type": "string", "description": "Override the write-index path."},
			},
			"required": []string{"paths"},
		},
	}}
}

func okResult(id json.RawMessage, res any) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "result": res}
}

func rpcError(id json.RawMessage, code int, msg string) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": msg}}
}
