package main

import (
	"time"

	"github.com/NickP005/go_mcminterface"
)

// ============================================================================
// NODE DISCOVERY & BENCHMARKING
// ============================================================================

// StartNodeDiscovery starts periodic node discovery and benchmarking goroutines
// This function should be called after go_mcminterface.LoadSettings()
func StartNodeDiscovery() {
	if !Globals.NodeDiscoveryEnabled {
		mlog(4, "§bnode_discovery: §7Node discovery disabled")
		return
	}

	if !Globals.OnlineMode {
		mlog(4, "§bnode_discovery: §7Node discovery disabled (offline mode)")
		return
	}

	mlog(3, "§bnode_discovery: §aStarting node discovery and benchmarking goroutines")

	// Parse intervals
	discoveryInterval, err := time.ParseDuration(Globals.DiscoveryInterval)
	if err != nil {
		mlog(2, "§bnode_discovery: §cInvalid discovery_interval '%s', using 30m", Globals.DiscoveryInterval)
		discoveryInterval = 30 * time.Minute
	}

	benchmarkInterval, err := time.ParseDuration(Globals.BenchmarkInterval)
	if err != nil {
		mlog(2, "§bnode_discovery: §cInvalid benchmark_interval '%s', using 15m", Globals.BenchmarkInterval)
		benchmarkInterval = 15 * time.Minute
	}

	// Start discovery goroutine
	go nodeDiscoveryLoop(discoveryInterval)

	// Start benchmarking goroutine
	go nodeBenchmarkLoop(benchmarkInterval)
}

// nodeDiscoveryLoop periodically discovers new nodes using ExpandIPs
func nodeDiscoveryLoop(interval time.Duration) {
	mlog(3, "§bnode_discovery: §aDiscovery loop started (interval: %s)", interval)

	// Run immediately on startup
	discoverNodes()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		discoverNodes()
	}
}

// discoverNodes performs node discovery using go_mcminterface.ExpandIPs
func discoverNodes() {
	mlog(4, "§bnode_discovery: §fRunning ExpandIPs() to discover new nodes...")

	startTime := time.Now()

	// Get current IP count before expansion
	currentCount := len(go_mcminterface.Settings.IPs)

	// Expand IPs to discover peers
	// ExpandIPs() mutates Settings.IPs directly
	go_mcminterface.ExpandIPs()

	// Get new count after expansion
	newCount := len(go_mcminterface.Settings.IPs)
	elapsed := time.Since(startTime)

	if newCount > currentCount {
		mlog(3, "§bnode_discovery: §aDiscovered %d new nodes (total: %d → %d) in %s",
			newCount-currentCount, currentCount, newCount, elapsed)
	} else {
		mlog(4, "§bnode_discovery: §fNo new nodes discovered (%d total) in %s",
			currentCount, elapsed)
	}

	// Check if we're below minimum threshold
	if newCount < Globals.MinNodesThreshold && newCount > 0 {
		mlog(2, "§bnode_discovery: §e⚠ Node count (%d) below threshold (%d), triggering emergency discovery",
			newCount, Globals.MinNodesThreshold)
		// Trigger immediate benchmark to ensure we're using best available nodes
		go benchmarkNodes()
	}

	// Limit pool size if configured
	if Globals.MaxNodesPool > 0 && newCount > Globals.MaxNodesPool {
		mlog(4, "§bnode_discovery: §fNode pool exceeds max size (%d > %d), will trim after benchmark",
			newCount, Globals.MaxNodesPool)
	}
}

// nodeBenchmarkLoop periodically benchmarks nodes to find the fastest
func nodeBenchmarkLoop(interval time.Duration) {
	mlog(3, "§bnode_discovery: §aBenchmark loop started (interval: %s)", interval)

	// Wait a bit before first benchmark to let discovery gather some nodes
	time.Sleep(10 * time.Second)
	benchmarkNodes()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		benchmarkNodes()
	}
}

// benchmarkNodes performs node benchmarking and selects the fastest nodes
func benchmarkNodes() {
	mlog(4, "§bnode_discovery: §fRunning BenchmarkNodes() to find fastest nodes...")

	startTime := time.Now()

	// Get current IP count
	currentCount := len(go_mcminterface.Settings.IPs)
	if currentCount == 0 {
		mlog(2, "§bnode_discovery: §e⚠ No nodes available to benchmark")
		return
	}

	// Determine how many nodes to benchmark
	benchCount := Globals.BenchmarkConcurrent
	if benchCount > currentCount {
		benchCount = currentCount
	}

	// Run benchmarking (this mutates Settings.IPs, sorting by speed)
	go_mcminterface.BenchmarkNodes(benchCount)

	elapsed := time.Since(startTime)

	// Get new count after benchmarking
	newCount := len(go_mcminterface.Settings.IPs)

	mlog(3, "§bnode_discovery: §aBenchmarked nodes (%d total, kept %d best) in %s",
		currentCount, newCount, elapsed)

	// Trim to MaxNodesPool if configured
	if Globals.MaxNodesPool > 0 && newCount > Globals.MaxNodesPool {
		mlog(4, "§bnode_discovery: §fTrimming node pool to %d nodes", Globals.MaxNodesPool)
		go_mcminterface.Settings.IPs = go_mcminterface.Settings.IPs[:Globals.MaxNodesPool]
	}

	// Log top nodes
	logCount := 5
	if len(go_mcminterface.Settings.IPs) < logCount {
		logCount = len(go_mcminterface.Settings.IPs)
	}
	for i := 0; i < logCount; i++ {
		mlog(4, "  §7#%d: §f%s", i+1, go_mcminterface.Settings.IPs[i])
	}
}
