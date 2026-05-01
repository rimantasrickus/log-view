package fileio

import (
	"fmt"
	"os"
	"strings"

	mmap "github.com/edsrzf/mmap-go"
)

// FileReader provides memory-mapped, read-only, random-access to a file's lines.
type FileReader struct {
	file     *os.File
	data     mmap.MMap
	offsets  []int64
	fileSize int64
	filePath string
}

// Open opens a file in read-only mode, memory-maps it, and builds a line index.
// progress is called with 0.0–1.0 during indexing (may be nil).
func Open(path string, progress func(float64)) (*FileReader, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("stat file: %w", err)
	}

	size := info.Size()

	// Handle empty files
	if size == 0 {
		return &FileReader{
			file:     f,
			data:     nil,
			offsets:  []int64{0},
			fileSize: 0,
			filePath: path,
		}, nil
	}

	data, err := mmap.MapRegion(f, int(size), mmap.RDONLY, 0, 0)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("mmap file: %w", err)
	}

	offsets := BuildLineIndex(data, progress)

	return &FileReader{
		file:     f,
		data:     data,
		offsets:  offsets,
		fileSize: size,
		filePath: path,
	}, nil
}

// LineCount returns the total number of lines in the file.
func (r *FileReader) LineCount() int {
	return len(r.offsets)
}

// Line returns the text content of line n (0-indexed), with \r\n stripped.
func (r *FileReader) Line(n int) string {
	if n < 0 || n >= len(r.offsets) {
		return ""
	}

	start := r.offsets[n]
	var end int64
	if n+1 < len(r.offsets) {
		end = r.offsets[n+1]
	} else {
		end = int64(len(r.data))
	}

	line := string(r.data[start:end])
	line = strings.TrimRight(line, "\r\n")
	return line
}

// FileSize returns the size of the file in bytes.
func (r *FileReader) FileSize() int64 {
	return r.fileSize
}

// FilePath returns the original file path.
func (r *FileReader) FilePath() string {
	return r.filePath
}

// Close unmaps the file and closes the file handle.
func (r *FileReader) Close() error {
	var firstErr error
	if r.data != nil {
		if err := r.data.Unmap(); err != nil {
			firstErr = err
		}
	}
	if err := r.file.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
