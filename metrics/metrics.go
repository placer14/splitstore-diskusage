package metrics

import (
	"log"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

func init() {
	if err := view.Register(AllViews...); err != nil {
		log.Fatal(err)
	}
}

var (
	// Measures
	ColdStoreBadgerSize    = stats.Int64("splitstore/coldstore_badger_size", "Size of the coldstore badger store", stats.UnitBytes)
	HotStoreBadgerSize     = stats.Int64("splitstore/hotstore_badger_size", "Size of the hotstore badger store", stats.UnitBytes)
	MarkSetBadgerSize      = stats.Int64("splitstore/markset_badger_size", "Size of the markset badger store", stats.UnitBytes)
	DiskUsageLastUpdatedAt = stats.Int64("splitstore/diskusage_last_updated_at", "Number of seconds since Unix epoch that usage was most recently updated", stats.UnitSeconds)

	// Views
	AllViews = []*view.View{
		ColdStoreBadgerSizeView,
		HotStoreBadgerSizeView,
		MarkSetBadgerSizeView,
		DiskUsageLastUpdatedAtView,
	}
	ColdStoreBadgerSizeView = &view.View{
		Measure:     ColdStoreBadgerSize,
		Aggregation: view.LastValue(),
	}
	HotStoreBadgerSizeView = &view.View{
		Measure:     HotStoreBadgerSize,
		Aggregation: view.LastValue(),
	}
	MarkSetBadgerSizeView = &view.View{
		Measure:     MarkSetBadgerSize,
		Aggregation: view.LastValue(),
	}
	DiskUsageLastUpdatedAtView = &view.View{
		Measure:     DiskUsageLastUpdatedAt,
		Aggregation: view.LastValue(),
	}
)
