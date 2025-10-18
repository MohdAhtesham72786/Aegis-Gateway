package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/example/aegis-gateway/internal/policy"
	"github.com/example/aegis-gateway/pkg/logging"
    "github.com/example/aegis-gateway/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/attribute"
	"github.com/example/aegis-gateway/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Gateway struct {
	pl *policy.Loader
	tr trace.Tracer
	paymentsURL string
	filesURL string
	r *mux.Router
}

func New(pl *policy.Loader, tr trace.Tracer, paymentsURL, filesURL string) *Gateway {
	g := &Gateway{pl:pl, tr:tr, paymentsURL:paymentsURL, filesURL:filesURL}
	g.r = mux.NewRouter()
	g.r.Handle("/metrics", promhttp.Handler())
	g.r.HandleFunc("/tools/{tool}/{action}", g.handle).Methods("POST")
	return g
}

func (g *Gateway) Router() http.Handler { return g.r }

func (g *Gateway) handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tool := vars["tool"]
	action := vars["action"]
	agent := r.Header.Get("X-Agent-ID")
	parent := r.Header.Get("X-Parent-Agent")
	b, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil { http.Error(w, "bad body", http.StatusBadRequest); return }
	// params hash
	h := sha256.Sum256(b)
	ph := hex.EncodeToString(h[:])

	// capture caller IP
	callerIP := r.RemoteAddr

	ctx, span := g.tr.Start(ctx, "policy.evaluate")
	span.SetAttributes(attribute.String("agent.id", agent), attribute.String("tool.name", tool), attribute.String("tool.action", action))
	if tid := telemetry.TraceIDFromContext(ctx); tid != "" { span.SetAttributes(attribute.String("trace.id", tid)) }
	if parent != "" { span.SetAttributes(attribute.String("agent.parent", parent)) }
	allow, reason, pv := g.pl.Evaluate(agent, tool, action, b)
	span.SetAttributes(attribute.Bool("decision.allow", allow))
	span.End()

	// include trace id
	traceID := ""
	if tid := telemetry.TraceIDFromContext(ctx); tid != "" { traceID = tid }

	logEntry := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339),
		"agent.id": agent,
		"agent.parent": parent,
		"tool.name": tool,
		"tool.action": action,
		"decision.allow": allow,
		"policy.version": pv,
		"params.hash": ph,
		"trace.id": traceID,
	}

	if !allow {
		logEntry["reason"] = reason
		logging.LogJSON(logEntry)
		// metrics
		metrics.Requests.WithLabelValues(agent, tool, action, "deny").Inc()
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error":"PolicyViolation","reason":reason})
		return
	}

	// forward to adapter with tracing
	ctx, span := g.tr.Start(ctx, "adapter.forward")
	span.SetAttributes(attribute.String("policy.version", pv), attribute.String("params.hash", ph))
	if tid := telemetry.TraceIDFromContext(ctx); tid != "" { span.SetAttributes(attribute.String("trace.id", tid)) }
	respBody, status, latencyMs, err := g.forward(ctx, tool, action, b)
	span.SetAttributes(attribute.Int64("latency.ms", latencyMs))
	if err != nil {
		span.RecordError(err)
		span.End()
		logging.LogJSON(map[string]interface{}{"time":time.Now().Format(time.RFC3339), "agent.id":agent, "tool.name":tool, "tool.action":action, "error":"forward","msg":err.Error()})
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	span.End()
	logEntry["upstream.status"] = status
	logEntry["latency.ms"] = latencyMs
	logEntry["caller.ip"] = callerIP
	logging.LogJSON(logEntry)
	// metrics
	metrics.Requests.WithLabelValues(agent, tool, action, "allow").Inc()
	metrics.Latency.WithLabelValues(agent, tool, action).Observe(float64(latencyMs) / 1000.0)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)
	w.Write(respBody)
}

func (g *Gateway) forward(ctx context.Context, tool, action string, body []byte) ([]byte, int, int64, error) {
	var url string
	switch tool {
	case "payments": url = g.paymentsURL + "/" + action
	case "files": url = g.filesURL + "/" + action
	default: return nil, 404, nil
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type","application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(start).Milliseconds()
	if err != nil { return nil, 0, lat, err }
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return b, resp.StatusCode, lat, nil
}


