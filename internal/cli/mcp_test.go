package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mcpReq(t *testing.T, s *mcpServer, id, method string, params string) map[string]any {
	t.Helper()
	req := &rpcMsg{JSONRPC: "2.0", Method: method}
	if id != "" {
		req.ID = json.RawMessage(id)
	}
	if params != "" {
		req.Params = json.RawMessage(params)
	}
	return s.handle(req)
}

func TestMCP_Initialize(t *testing.T) {
	s := &mcpServer{}
	resp := mcpReq(t, s, "1", "initialize", `{"protocolVersion":"2024-11-05","capabilities":{}}`)
	res, _ := resp["result"].(map[string]any)
	if res == nil {
		t.Fatalf("no result: %v", resp)
	}
	if res["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion = %v", res["protocolVersion"])
	}
	if si, _ := res["serverInfo"].(map[string]any); si["name"] != "eis" {
		t.Errorf("serverInfo.name = %v", si["name"])
	}
	caps, _ := res["capabilities"].(map[string]any)
	if _, ok := caps["tools"]; !ok {
		t.Errorf("capabilities missing tools: %v", caps)
	}
}

func TestMCP_NotificationHasNoResponse(t *testing.T) {
	s := &mcpServer{}
	if resp := mcpReq(t, s, "", "notifications/initialized", ""); resp != nil {
		t.Errorf("notification must get no response, got %v", resp)
	}
}

func TestMCP_ToolsList(t *testing.T) {
	s := &mcpServer{}
	resp := mcpReq(t, s, "2", "tools/list", "")
	res, _ := resp["result"].(map[string]any)
	tools, _ := res["tools"].([]map[string]any)
	if len(tools) != 1 || tools[0]["name"] != "get_write_context" {
		t.Fatalf("tools/list = %v", res["tools"])
	}
}

func TestMCP_UnknownMethodErrors(t *testing.T) {
	s := &mcpServer{}
	resp := mcpReq(t, s, "9", "bogus/method", "")
	e, _ := resp["error"].(map[string]any)
	if e == nil || e["code"] != -32601 {
		t.Errorf("want -32601 method not found, got %v", resp)
	}
}

func TestMCP_ToolsCall_GetWriteContext(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".eis"), 0o755); err != nil {
		t.Fatal(err)
	}
	idx := `{"generated_at":"2026-07-10T00:00:00Z","commit":"x","modules":{` +
		`"svc/auth":{"debt_tier":"Orphaned","at_risk":false,"owner_active":false,` +
		`"owner_left_days":243,"untouched_days":243,"ownership_concentration":0.9,` +
		`"recommendation":"orphaned_module"}},"module_patterns":["svc/*"],"module_patterns_source":"config"}`
	idxPath := filepath.Join(dir, ".eis", "write-index.json")
	if err := os.WriteFile(idxPath, []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &mcpServer{}
	params := `{"name":"get_write_context","arguments":{"paths":["svc/auth/handler.go"],"index_path":"` + idxPath + `"}}`
	resp := mcpReq(t, s, "3", "tools/call", params)
	res, _ := resp["result"].(map[string]any)
	if res == nil {
		t.Fatalf("no result: %v", resp)
	}
	if res["isError"] != false {
		t.Errorf("isError = %v", res["isError"])
	}
	content, _ := res["content"].([]map[string]any)
	if len(content) == 0 {
		t.Fatalf("empty content")
	}
	text, _ := content[0]["text"].(string)
	if !strings.Contains(text, "orphaned_module") || !strings.Contains(text, "svc/auth") {
		t.Errorf("tool result missing orphaned context: %s", text)
	}
	// Firewall: no author identity in the tool result.
	for _, tok := range []string{"last_owner", "top_author", "\"author\"", "ponsaaan"} {
		if strings.Contains(text, tok) {
			t.Errorf("firewall: %q leaked into MCP result", tok)
		}
	}
}

func TestMCP_ToolsCall_UnknownTool(t *testing.T) {
	s := &mcpServer{}
	resp := mcpReq(t, s, "4", "tools/call", `{"name":"nope","arguments":{}}`)
	if e, _ := resp["error"].(map[string]any); e == nil || e["code"] != -32602 {
		t.Errorf("want -32602 for unknown tool, got %v", resp)
	}
}
