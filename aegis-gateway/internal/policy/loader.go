package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/fsnotify/fsnotify"
)

type PolicySet struct {
	Version int `yaml:"version"`
	Agents  []AgentPolicy `yaml:"agents"`
	Source  string
}

type AgentPolicy struct {
	ID    string `yaml:"id"`
	Allow []AllowRule `yaml:"allow"`
}

type AllowRule struct {
	Tool       string                 `yaml:"tool"`
	Actions    []string               `yaml:"actions"`
	Conditions map[string]interface{} `yaml:"conditions"`
}

// Loader watches policies dir and keeps an in-memory map
type Loader struct {
	dir string
	mu sync.RWMutex
	policies map[string]PolicySet
	lastLoad time.Time
	watcher *fsnotify.Watcher
}

func NewLoader(dir string) (*Loader, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	l := &Loader{dir:dir, policies: make(map[string]PolicySet), watcher: w}
	if err := l.loadAll(); err != nil {
		log.Printf("policy load initial error: %v", err)
	}
	if err := w.Add(dir); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Loader) Watch() {
	for {
		select {
		case ev, ok := <-l.watcher.Events:
			if !ok { return }
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				log.Printf("policy dir change: %s", ev.Name)
				if err := l.loadAll(); err != nil {
					log.Printf("policy reload error: %v", err)
				}
			}
		case err, ok := <-l.watcher.Errors:
			if !ok { return }
			log.Printf("policy watcher error: %v", err)
		}
	}
}

func (l *Loader) loadAll() error {
	files, err := ioutil.ReadDir(l.dir)
	if err != nil {
		return err
	}
	loaded := make(map[string]PolicySet)
	var lastErr error
	for _, f := range files {
		if f.IsDir() { continue }
		if filepath.Ext(f.Name()) != ".yaml" && filepath.Ext(f.Name()) != ".yml" { continue }
		p, err := l.loadFile(filepath.Join(l.dir, f.Name()))
		if err != nil {
			lastErr = err
			log.Printf("policy %s failed to load: %v", f.Name(), err)
			continue
		}
		p.Source = f.Name()
		loaded[f.Name()] = p
	}
	l.mu.Lock()
	l.policies = loaded
	l.lastLoad = time.Now()
	l.mu.Unlock()
	return lastErr
}

func (l *Loader) loadFile(path string) (PolicySet, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil { return PolicySet{}, err }
	var p PolicySet
	if err := yaml.Unmarshal(b, &p); err != nil { return PolicySet{}, err }
	// basic validation
	if p.Version != 1 { return PolicySet{}, errors.New("unsupported version") }
	for _, a := range p.Agents {
		if a.ID == "" { return PolicySet{}, errors.New("agent missing id") }
		for _, rule := range a.Allow {
			if rule.Tool == "" { return PolicySet{}, errors.New("rule missing tool") }
			if len(rule.Actions) == 0 { return PolicySet{}, errors.New("rule missing actions") }
		}
	}
	return p, nil
}

// Evaluate whether agent is allowed
func (l *Loader) Evaluate(agentID, tool, action string, params []byte) (allow bool, reason string, policyVersion string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	// build params hash
	h := sha256.Sum256(params)
	phash := hex.EncodeToString(h[:])
	for _, p := range l.policies {
		for _, a := range p.Agents {
			if a.ID != agentID { continue }
			for _, r := range a.Allow {
				if r.Tool != tool { continue }
				// action match
				okAction := false
				for _, ac := range r.Actions { if ac == action { okAction = true; break } }
				if !okAction { continue }
				// conditions
				if cmax, ok := r.Conditions["max_amount"]; ok {
					// expect numeric
					amt := extractAmount(params)
					if amt < 0 { return false, "invalid params (amount missing)", p.Source }
					if float64(amt) > cmax.(float64) {
						return false, "Amount exceeds max_amount="+formatNumber(int(cmax.(float64))), p.Source
					}
				}
				if cc, ok := r.Conditions["currencies"]; ok {
					cur := extractCurrency(params)
					if cur == "" { return false, "invalid params (currency missing)", p.Source }
					allowed := cc.([]interface{})
					ok := false
					for _, v := range allowed { if v.(string) == cur { ok = true; break } }
					if !ok { return false, "currency not allowed", p.Source }
				}
				if prefix, ok := r.Conditions["folder_prefix"]; ok {
					path := extractPath(params)
					if path == "" { return false, "invalid params (path missing)", p.Source }
					pp := prefix.(string)
					if len(path) < len(pp) || path[:len(pp)] != pp {
						return false, "path outside allowed folder", p.Source
					}
				}
				// allowed
				_ = phash
				return true, "", p.Source
			}
			}
		}
	}
	return false, "no matching policy allow rule", ""
}

func formatNumber(n int) string { return strconv.Itoa(n) }

func extractAmount(b []byte) float64 {
	// naive parse to find "amount": number
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil { return -1 }
	if v, ok := m["amount"]; ok {
		switch vv := v.(type) {
		case int: return float64(vv)
		case int64: return float64(vv)
		case float64: return vv
		}
	}
	return -1
}

func extractCurrency(b []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil { return "" }
	if v, ok := m["currency"]; ok {
		if s, ok := v.(string); ok { return s }
	}
	return ""
}

func extractPath(b []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil { return "" }
	if v, ok := m["path"]; ok {
		if s, ok := v.(string); ok { return s }
	}
	return ""
}
