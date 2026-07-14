package tui

import (
	netbridge "github.com/netbridge/netbridge"
)

type BenchmarkView struct {
	engine netbridge.CoreEngine
}

func NewBenchmarkView(engine netbridge.CoreEngine) *BenchmarkView {
	return &BenchmarkView{engine: engine}
}

func (b *BenchmarkView) View() string {
	return `
  Benchmark Results:
  ─────────────────────────────────────────────
  Profile         Latency    Throughput    Score
  ─────────────────────────────────────────────
  (no data yet — run 'netbridge benchmark' first)
  
  Press 'b' to go back.
`
}
