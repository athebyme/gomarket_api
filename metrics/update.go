package metrics

import "sync/atomic"

type UpdateMetrics struct {
	UpdatedCount         atomic.Int32
	ProcessedCount       atomic.Int32
	ErroredNomenclatures atomic.Int32
	GoroutinesNmsCount   atomic.Int32
}
