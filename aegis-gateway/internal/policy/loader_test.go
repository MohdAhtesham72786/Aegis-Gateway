package policy

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"
)

func writeTempPolicy(t *testing.T, name, content string) string {
    t.Helper()
    dir := t.TempDir()
    p := filepath.Join(dir, name)
    if err := ioutil.WriteFile(p, []byte(content), 0o644); err != nil {
        t.Fatalf("write policy: %v", err)
    }
    return dir
}

func TestPolicyEvaluatePayments(t *testing.T) {
    yaml := `version: 1
agents:
  - id: finance-agent
    allow:
      - tool: payments
        actions: [create]
        conditions:
          max_amount: 5000
          currencies: [USD]
`
    dir := writeTempPolicy(t, "finance.yaml", yaml)
    l, err := NewLoader(dir)
    if err != nil { t.Fatalf("loader: %v", err) }
    allow, _, _ := l.Evaluate("finance-agent", "payments", "create", []byte(`{"amount":100,"currency":"USD"}`))
    if !allow { t.Fatalf("expected allow") }
    allow, reason, _ := l.Evaluate("finance-agent", "payments", "create", []byte(`{"amount":6000,"currency":"USD"}`))
    if allow { t.Fatalf("expected deny") }
    if reason == "" { t.Fatalf("expected reason") }
}

func TestPolicyEvaluateFiles(t *testing.T) {
    yaml := `version: 1
agents:
  - id: hr-agent
    allow:
      - tool: files
        actions: [read]
        conditions:
          folder_prefix: "/hr-docs/"
`
    dir := writeTempPolicy(t, "hr.yaml", yaml)
    l, err := NewLoader(dir)
    if err != nil { t.Fatalf("loader: %v", err) }
    allow, _, _ := l.Evaluate("hr-agent", "files", "read", []byte(`{"path":"/hr-docs/employee1.txt"}`))
    if !allow { t.Fatalf("expected allow") }
    allow, reason, _ := l.Evaluate("hr-agent", "files", "read", []byte(`{"path":"/legal/contract.docx"}`))
    if allow { t.Fatalf("expected deny") }
    if reason == "" { t.Fatalf("expected reason") }
}
