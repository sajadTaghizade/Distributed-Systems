package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"text/tabwriter"
	"time"
)

func cpuBoundTask() int {
	var sum int64
	for i := int64(1); i <= 5000000; i++ {
		sum += i * i
	}
	return 0
}

func mixedTask() int {
	var sum int64
	blocks := 0
	for i := int64(1); i <= 5000000; i++ {
		sum += i * i
		if i%500000 == 0 {
			time.Sleep(time.Millisecond)
			blocks++
		}
	}
	return blocks
}

func runBenchmark(workloadName string, taskFunc func() int, numGoroutines int, maxProcs int, w *tabwriter.Writer) {
	runtime.GOMAXPROCS(maxProcs)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	latencies := make([]time.Duration, numGoroutines)
	blocks := make([]int, numGoroutines)

	start := time.Now()
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			startTask := time.Now()
			b := taskFunc()
			latencies[idx] = time.Since(startTask)
			blocks[idx] = b
		}(i)
	}
	wg.Wait()
	totalTime := time.Since(start)

	var minLat, maxLat time.Duration
	var latSum int64
	var totalBlocks int
	if numGoroutines > 0 {
		minLat = latencies[0]
		maxLat = latencies[0]
		for i, l := range latencies {
			if l < minLat {
				minLat = l
			}
			if l > maxLat {
				maxLat = l
			}
			latSum += int64(l)
			totalBlocks += blocks[i]
		}
	}

	avgLatency := time.Duration(latSum / int64(numGoroutines))

	var varianceSum float64
	avgLatSecs := avgLatency.Seconds()
	for _, l := range latencies {
		diff := l.Seconds() - avgLatSecs
		varianceSum += diff * diff
	}
	stdDevSecs := math.Sqrt(varianceSum / float64(numGoroutines))
	stdDev := time.Duration(stdDevSecs * float64(time.Second))

	throughput := float64(numGoroutines) / totalTime.Seconds()

	fmt.Fprintf(w, "%s\t%d\t%d\t%v\t%.2f\t%v\t%v\t%v\t%v\t%d\n", workloadName, maxProcs, numGoroutines, totalTime, throughput, avgLatency, minLat, maxLat, stdDev, totalBlocks)
}

func main() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Workload\tGOMAXPROCS\tGoroutines\tTotal Time\tThroughput (ops/sec)\tAvg Latency\tMin Latency\tMax Latency\tStd Dev\tTotal Blocks")

	procs := []int{1, 2, runtime.NumCPU()}
	routines := []int{1, 2, 4, 8, 16, 32, 64}

	for _, p := range procs {
		for _, r := range routines {
			runBenchmark("CPU-bound", cpuBoundTask, r, p, w)
			runBenchmark("Mixed", mixedTask, r, p, w)
		}
	}
	w.Flush()
}