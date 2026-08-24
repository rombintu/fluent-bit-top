package shared

import "fmt"

// Health represents per-component pipeline health.
type Health int8

const (
	HealthOK      Health = iota // records flowing, no errors
	HealthWarning               // retries, throughput anomaly
	HealthError                 // errors, dropped records, retries_failed
	HealthIdle                  // zero throughput — possible misconfig
)

func (h Health) String() string {
	return [...]string{"ok", "warn", "error", "idle"}[h]
}

// Badge returns a short colored indicator for table display.
func (h Health) Badge() string {
	switch h {
	case HealthOK:
		return "●"
	case HealthWarning:
		return "▲"
	case HealthError:
		return "✖"
	case HealthIdle:
		return "○"
	default:
		return "?"
	}
}

// Diag is a single diagnostic message for the warnings panel.
type Diag struct {
	Level   Health
	CompID  string // component ID, or "" for global
	Kind    Kind   // component kind
	Message string
	Hint    string // suggested config area to check
}

// AnalyzeDiagnostics inspects a State and produces actionable diagnostics.
func AnalyzeDiagnostics(s State, prevComps []ComponentMetrics) []Diag {
	var diags []Diag

	prevMap := make(map[string]ComponentMetrics, len(prevComps))
	for _, c := range prevComps {
		prevMap[c.ID] = c
	}

	for _, c := range s.Comps {
		diags = append(diags, componentDiags(c, prevMap[c.ID])...)
	}

	// Global diagnostics
	totalOut, totalIn := flowTotals(s.Comps)
	if totalIn > 0 && totalOut == 0 {
		diags = append(diags, Diag{
			Level:   HealthError,
			Message: "No output records — pipeline is blocked or outputs misconfigured",
			Hint:    "Check output plugin destination, auth, and network connectivity",
		})
	}

	return diags
}

// ComponentHealth returns the health status for a single component.
func ComponentHealth(c ComponentMetrics, prev ComponentMetrics) Health {
	switch c.Kind {
	case KindOutput:
		if c.Fields["errors"] > 0 || c.Fields["retries_failed"] > 0 {
			return HealthError
		}
		if c.Fields["retries"] > 0 || c.Fields["dropped_records"] > 0 {
			return HealthWarning
		}
		if c.Records() == 0 {
			return HealthIdle
		}
	case KindFilter:
		total := c.Records() + c.Fields["drop_records"]
		if total > 0 {
			dropRatio := float64(c.Fields["drop_records"]) / float64(total)
			if dropRatio > 0.5 {
				return HealthWarning
			}
		}
		if c.Records() == 0 {
			return HealthIdle
		}
	case KindInput:
		if c.Records() == 0 {
			return HealthIdle
		}
	}
	return HealthOK
}

// PipelineSummary is a high-level summary for the header.
type PipelineSummary struct {
	TotalInputs  int
	TotalFilters int
	TotalOutputs int
	TotalErrors  int64
	TotalRetries int64
	TotalDropped int64
	InputRec     int64
	OutputRec    int64
	HasErrors    bool
	HasWarnings  bool
	Flowing      bool
}

func SummarizePipeline(comps []ComponentMetrics) PipelineSummary {
	var ps PipelineSummary
	for _, c := range comps {
		switch c.Kind {
		case KindInput:
			ps.TotalInputs++
			ps.InputRec += c.Records()
		case KindFilter:
			ps.TotalFilters++
		case KindOutput:
			ps.TotalOutputs++
			ps.OutputRec += c.Records()
			ps.TotalErrors += c.Fields["errors"]
			ps.TotalRetries += c.Fields["retries"]
			ps.TotalDropped += c.Fields["dropped_records"]
		}
	}
	ps.HasErrors = ps.TotalErrors > 0
	ps.HasWarnings = ps.TotalRetries > 0 || ps.TotalDropped > 0
	ps.Flowing = ps.InputRec > 0 && ps.OutputRec > 0
	return ps
}

func componentDiags(c, prev ComponentMetrics) []Diag {
	var diags []Diag
	h := ComponentHealth(c, prev)

	switch h {
	case HealthError:
		if c.Kind == KindOutput {
			if e := c.Fields["errors"]; e > 0 {
				diags = append(diags, Diag{
					Level:   HealthError,
					CompID:  c.ID,
					Kind:    c.Kind,
					Message: fmt.Sprintf("%d processing errors", e),
					Hint:    "Check output destination availability, TLS certs, and credentials",
				})
			}
			if rf := c.Fields["retries_failed"]; rf > 0 {
				diags = append(diags, Diag{
					Level:   HealthError,
					CompID:  c.ID,
					Kind:    c.Kind,
					Message: fmt.Sprintf("%d failed retries", rf),
					Hint:    "Destination may be unreachable — check network, DNS, and firewall rules",
				})
			}
		}
	case HealthWarning:
		if c.Kind == KindOutput {
			if r := c.Fields["retries"]; r > 0 {
				diags = append(diags, Diag{
					Level:   HealthWarning,
					CompID:  c.ID,
					Kind:    c.Kind,
					Message: fmt.Sprintf("%d retries", r),
					Hint:    "Transient failures — monitor or increase retry_limit in output config",
				})
			}
			if d := c.Fields["dropped_records"]; d > 0 {
				diags = append(diags, Diag{
					Level:   HealthWarning,
					CompID:  c.ID,
					Kind:    c.Kind,
					Message: fmt.Sprintf("%d records dropped", d),
					Hint:    "Check queue_limit / mem_buf_limit in output config",
				})
			}
		}
		if c.Kind == KindFilter {
			total := c.Records() + c.Fields["drop_records"]
			if total > 0 && float64(c.Fields["drop_records"])/float64(total) > 0.5 {
				diags = append(diags, Diag{
					Level:   HealthWarning,
					CompID:  c.ID,
					Kind:    c.Kind,
					Message: fmt.Sprintf("dropping >%d%% of records", int(float64(c.Fields["drop_records"])/float64(total)*100)),
					Hint:    "Verify filter conditions — high drop rate may indicate misconfigured rules",
				})
			}
		}
	case HealthIdle:
		switch c.Kind {
		case KindInput:
			diags = append(diags, Diag{
				Level:   HealthIdle,
				CompID:  c.ID,
				Kind:    c.Kind,
				Message: "No records collected",
				Hint:    "Check path/tag/source in input config — file may not exist or be empty",
			})
		case KindOutput:
			diags = append(diags, Diag{
				Level:   HealthIdle,
				CompID:  c.ID,
				Kind:    c.Kind,
				Message: "No records processed",
				Hint:    "Upstream may be blocked — verify filters are not dropping all records",
			})
		case KindFilter:
			diags = append(diags, Diag{
				Level:   HealthIdle,
				CompID:  c.ID,
				Kind:    c.Kind,
				Message: "No records passed through",
				Hint:    "Check filter match pattern and upstream output tags",
			})
		}
	}
	return diags
}

func flowTotals(comps []ComponentMetrics) (in, out int64) {
	for _, c := range comps {
		switch c.Kind {
		case KindInput:
			in += c.Records()
		case KindOutput:
			out += c.Records()
		}
	}
	return
}
