package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"fbtop/internal/shared"
)

func Fetch(url string) (shared.State, error) {
	resp, err := (&http.Client{Timeout: 3 * time.Second}).Get(url + "/api/v1/metrics")
	if err != nil {
		return shared.State{}, err
	}
	defer resp.Body.Close()

	var raw map[string]map[string]map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return shared.State{}, err
	}

	kinds := map[string]shared.Kind{
		"input":  shared.KindInput,
		"filter": shared.KindFilter,
		"output": shared.KindOutput,
	}

	now := time.Now()
	var cs []shared.ComponentMetrics
	for section, comps := range raw {
		k, ok := kinds[section]
		if !ok {
			continue
		}
		for id, fields := range comps {
			cs = append(cs, shared.ComponentMetrics{ID: id, Kind: k, Fields: fields, Stamp: now})
		}
	}

	// Stable sort by kind for consistent display order
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].Kind < cs[j].Kind })

	return shared.State{Comps: cs, Connected: true, UpdatedAt: now}, nil
}
