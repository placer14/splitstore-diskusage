package splitstore

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/placer14/splitstore-diskusage/metrics"
	"go.opencensus.io/stats"
)

// SplitstoreDiskUsageAgent runs a daemon which periodically checks the disk usage of the splitstore and
// and emits metrics to prometheus.
// It accepts the following flags:
//
//	--interval duration: the interval at which to check the disk usage (default 10m0s)
//	--repo-path string: the path to the splitstore (required)
//	--metrics-endpoint string: the endpoint to expose metrics on (default ":8080")
//	--metrics-path string: the path to expose metrics on (default "/metrics")
type SplitstoreDiskUsageAgent struct {
	// contains filtered or unexported fields
	opts AgentOptions
}
type AgentOptions struct {
	Interval        string `long:"interval" description:"the interval at which to check the disk usage" default:"10m"`
	RepoPath        string `long:"repo-path" description:"the path to the splitstore" required:"true"`
	MetricsEndpoint string `long:"metrics-endpoint" description:"the endpoint to expose metrics on" default:":8080"`
	MetricsPath     string `long:"metrics-path" description:"the path to expose metrics on" default:"/metrics"`
}

// NewDiskUsageAgent creates a new SplitstoreDiskUsageAgent
func NewDiskUsageAgent(o AgentOptions) *SplitstoreDiskUsageAgent {
	return &SplitstoreDiskUsageAgent{
		opts: o,
	}
}

// Start initializes the prometheus endpoint then starts a goroutine which periodically
// calls getDiskUsage() and provides results to updateMetrics() every o.Interval duration.
func (a *SplitstoreDiskUsageAgent) Start(ctx context.Context) {
	a.initMetricsEndpoint()
	d, err := time.ParseDuration(a.opts.Interval)
	if err != nil {
		log.Fatal("error parsing interval:", err)
	}
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			a.updateMetrics()
		case <-ctx.Done():
			log.Println("context cancelled, stopping agent")
			return
		}
	}
}

type diskUsageResult struct {
	marksetBadger   int64
	hotstoreBadger  int64
	coldstoreBadger int64
}

// getDiskUsage executes the du command on the splitstore directory and returns the result
func (a *SplitstoreDiskUsageAgent) getDiskUsage() diskUsageResult {
	var (
		result diskUsageResult
		err    error
	)
	result.coldstoreBadger, err = a.parseDiskUsageOn("chain")
	if err != nil {
		log.Println("error parsing usage on 'chain':", err)
	}
	result.hotstoreBadger, err = a.parseDiskUsageOn("splitstore/hot.badger")
	if err != nil {
		log.Println("error parsing usage on 'hot.badger':", err)
	}
	result.marksetBadger, err = a.parseDiskUsageOn("splitstore/markset.badger")
	if err != nil {
		log.Println("error parsing usage on 'markset.badger':", err)
	}
	return result
}

func (a *SplitstoreDiskUsageAgent) parseDiskUsageOn(target string) (int64, error) {
	// check if target exists using os.Stat
	_, err := os.Stat(path.Join(a.opts.RepoPath, target))
	if err != nil {
		return 0, err
	}

	folderPath := path.Join(a.opts.RepoPath, target)
	duCmd := exec.Command("du", "-s", folderPath)
	var out bytes.Buffer
	duCmd.Stdout = &out
	if err := duCmd.Run(); err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.Fields(out.String())[0], 10, 64)
}

func (a *SplitstoreDiskUsageAgent) initMetricsEndpoint() {
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "splitstore",
	})
	if err != nil {
		log.Fatal("error creating exporter:", err)
	}
	http.Handle(a.opts.MetricsPath, exporter)

	go func() {
		log.Println("starting metrics endpoint on:", a.opts.MetricsEndpoint)
		if err := http.ListenAndServe(a.opts.MetricsEndpoint, nil); err != http.ErrServerClosed {
			log.Fatal("error starting metrics endpoint:", err)
		}
	}()
}

func (a *SplitstoreDiskUsageAgent) updateMetrics() {
	duResult := a.getDiskUsage()
	stats.Record(context.TODO(), metrics.ColdStoreBadgerSize.M(duResult.coldstoreBadger))
	stats.Record(context.TODO(), metrics.HotStoreBadgerSize.M(duResult.hotstoreBadger))
	stats.Record(context.TODO(), metrics.MarkSetBadgerSize.M(duResult.marksetBadger))
	stats.Record(context.TODO(), metrics.DiskUsageLastUpdatedAt.M(time.Now().Unix()))
	log.Println("updated metrics", time.Now().Format("2006-01-02 15:04:05"))
}
