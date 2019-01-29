package metrics

import "github.com/netdata/go-plugin/pkg/stm"

// Observer is a interface that wraps the Observe method, which is used by
// Histogram and Summary to add observations.
type Observer interface {
	stm.Value
	Observe(v float64)
}
