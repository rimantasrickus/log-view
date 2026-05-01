package filter

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"log-viewer/internal/fileio"
)

// Result holds the outcome of a filter operation.
type Result struct {
	// OriginalLineNumbers contains 0-indexed line numbers from the source file
	// that matched the filter query.
	OriginalLineNumbers []int
}

// Run filters lines in reader matching query. Returns matching original line numbers.
// If caseSensitive is false, comparison is case-insensitive.
// Cancellable via ctx. progress is called with 0.0–1.0 if non-nil.
func Run(ctx context.Context, reader *fileio.FileReader, query string, caseSensitive bool, progress func(float64)) (*Result, error) {
	totalLines := reader.LineCount()
	if totalLines == 0 || query == "" {
		return &Result{}, nil
	}

	searchQuery := query
	if !caseSensitive {
		searchQuery = strings.ToLower(query)
	}

	numWorkers := runtime.NumCPU()
	if totalLines < numWorkers {
		numWorkers = 1
	}

	chunkSize := totalLines / numWorkers

	type chunkResult struct {
		matches []int
	}

	results := make([]chunkResult, numWorkers)
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		completed int
		firstErr  error
	)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerIdx int) {
			defer wg.Done()

			start := workerIdx * chunkSize
			end := start + chunkSize
			if workerIdx == numWorkers-1 {
				end = totalLines
			}

			var matches []int
			for i := start; i < end; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				line := reader.Line(i)
				if !caseSensitive {
					line = strings.ToLower(line)
				}
				if strings.Contains(line, searchQuery) {
					matches = append(matches, i)
				}
			}

			results[workerIdx] = chunkResult{matches: matches}

			if progress != nil {
				mu.Lock()
				completed += end - start
				p := float64(completed) / float64(totalLines)
				mu.Unlock()
				progress(p)
			}
		}(w)
	}

	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}

	// Merge results in order
	total := 0
	for _, r := range results {
		total += len(r.matches)
	}

	merged := make([]int, 0, total)
	for _, r := range results {
		merged = append(merged, r.matches...)
	}

	if progress != nil {
		progress(1.0)
	}

	return &Result{OriginalLineNumbers: merged}, nil
}
