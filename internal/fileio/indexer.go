package fileio

import (
	"bytes"
	"runtime"
	"sync"
)

// BuildLineIndex scans data for newline characters and returns a slice of byte offsets
// where each line starts. The first entry is always 0 (start of file).
// Parallelized across CPU cores for large files.
// progress is called with values 0.0–1.0 if non-nil.
func BuildLineIndex(data []byte, progress func(float64)) []int64 {
	if len(data) == 0 {
		return []int64{0}
	}

	numCPU := runtime.NumCPU()
	dataLen := int64(len(data))

	// For small files, single-threaded is faster
	if dataLen < 1<<20 || numCPU == 1 { // < 1MB
		return buildLineIndexSingle(data, progress)
	}

	return buildLineIndexParallel(data, numCPU, progress)
}

func buildLineIndexSingle(data []byte, progress func(float64)) []int64 {
	offsets := make([]int64, 0, len(data)/80) // estimate ~80 bytes per line
	offsets = append(offsets, 0)

	dataLen := int64(len(data))
	var lastReport float64

	for i := int64(0); i < dataLen; i++ {
		if data[i] == '\n' && i+1 < dataLen {
			offsets = append(offsets, i+1)
		}
		if progress != nil && i%1_000_000 == 0 {
			p := float64(i) / float64(dataLen)
			if p-lastReport >= 0.01 {
				progress(p)
				lastReport = p
			}
		}
	}

	if progress != nil {
		progress(1.0)
	}
	return offsets
}

type chunkResult struct {
	index   int
	offsets []int64
}

func buildLineIndexParallel(data []byte, numWorkers int, progress func(float64)) []int64 {
	dataLen := int64(len(data))
	chunkSize := dataLen / int64(numWorkers)

	var (
		mu          sync.Mutex
		completed   int64
		results     = make([]chunkResult, numWorkers)
		wg          sync.WaitGroup
	)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerIdx int) {
			defer wg.Done()

			start := int64(workerIdx) * chunkSize
			end := start + chunkSize
			if workerIdx == numWorkers-1 {
				end = dataLen
			}

			chunk := data[start:end]
			localOffsets := make([]int64, 0, len(chunk)/80)

			// First worker always starts at offset 0
			if workerIdx == 0 {
				localOffsets = append(localOffsets, 0)
			}

			searchFrom := 0
			for {
				idx := bytes.IndexByte(chunk[searchFrom:], '\n')
				if idx < 0 {
					break
				}
				absPos := start + int64(searchFrom) + int64(idx)
				if absPos+1 < dataLen {
					localOffsets = append(localOffsets, absPos+1)
				}
				searchFrom += idx + 1
			}

			results[workerIdx] = chunkResult{index: workerIdx, offsets: localOffsets}

			if progress != nil {
				mu.Lock()
				completed += end - start
				p := float64(completed) / float64(dataLen)
				mu.Unlock()
				progress(p)
			}
		}(w)
	}

	wg.Wait()

	// Merge results in order
	total := 0
	for _, r := range results {
		total += len(r.offsets)
	}

	merged := make([]int64, 0, total)
	for _, r := range results {
		merged = append(merged, r.offsets...)
	}

	if progress != nil {
		progress(1.0)
	}
	return merged
}
