package shared

import "time"

type Kind int8

const (
	KindInput Kind = iota
	KindFilter
	KindOutput
	KindAll // sentinel for filtering
)

func (k Kind) String() string { return [...]string{"input", "filter", "output", "all"}[k] }

type ComponentMetrics struct {
	ID     string
	Kind   Kind
	Fields map[string]int64
	Stamp  time.Time
}

func (c ComponentMetrics) Records() int64 {
	if c.Kind == KindOutput {
		return c.Fields["proc_records"]
	}
	return c.Fields["records"]
}

func (c ComponentMetrics) Bytes() int64 {
	if c.Kind == KindOutput {
		return c.Fields["proc_bytes"]
	}
	return c.Fields["bytes"]
}

type State struct {
	Comps     []ComponentMetrics
	PrevComps []ComponentMetrics // snapshot for rate computation
	Connected bool
	UpdatedAt time.Time
}
