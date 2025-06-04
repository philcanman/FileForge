// Package main implements a high-performance file generator utility called FileForge.
// FileForge creates multiple files with random content in parallel, with configurable
// file sizes, directory structure, and worker count for optimal performance.
package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// Adjust constants for better performance
	minBufferSize      = 256 * 1024       // 256KB
	maxBufferSize      = 16 * 1024 * 1024 // 16MB
	defaultFilesPerDir = 10000
)

// Pull in version from VERSION file
func getVersion() string {
	version, err := os.ReadFile("VERSION")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(version))
}

// Enhanced getOptimalBufferSize with better heuristics
func getOptimalBufferSize() int {
	pageSize := syscall.Getpagesize()

	// Use 256 pages or 1MB, whichever is larger
	bufferSize := max(pageSize*256, 1024*1024)

	// Ensure buffer size is within bounds
	if bufferSize < minBufferSize {
		return minBufferSize
	}
	if bufferSize > maxBufferSize {
		return maxBufferSize
	}

	return bufferSize
}

// parseSize parses a size string and converts it to bytes.
func parseSize(sizeStr string) (int, error) {
	// Remove any quotes and trim spaces
	sizeStr = strings.Trim(sizeStr, "'\"")
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Try to find the unit by looking for known suffixes
	units := []string{"GB", "MB", "KB", "B", "GIGABYTE", "MEGABYTE", "KILOBYTE", "BYTE", "GIGABYTES", "MEGABYTES", "KILOBYTES", "BYTES"}
	var value string
	var unit string

	// Find the first matching unit
	found := false
	for _, u := range units {
		if strings.HasSuffix(sizeStr, u) {
			value = strings.TrimSpace(strings.TrimSuffix(sizeStr, u))
			unit = u
			found = true
			break
		}
	}

	if !found {
		return 0, fmt.Errorf("invalid size format. Expected format: '1GB' or '1 GB'")
	}

	// Parse the numeric value
	numValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	// Convert based on unit
	switch unit {
	case "B", "BYTE", "BYTES":
		return numValue, nil
	case "KB", "KILOBYTE", "KILOBYTES":
		return numValue * 1024, nil
	case "MB", "MEGABYTE", "MEGABYTES":
		return numValue * 1024 * 1024, nil
	case "GB", "GIGABYTE", "GIGABYTES":
		return numValue * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported size unit: %s", unit)
	}
}

// createRandomFile creates a random file with the specified parameters.
func createRandomFile(filePath string, fileSize int, bufferSize int) error {
	// Ensure directory exists
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory for file %s: %v", filePath, err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", filePath, err)
	}
	defer file.Close()

	// Buffered write with specified buffer size
	bufWriter := bufio.NewWriterSize(file, bufferSize)

	// Generate random data and write to file in chunks
	remaining := fileSize
	buffer := make([]byte, bufferSize)
	for remaining > 0 {
		chunkSize := bufferSize
		if remaining < bufferSize {
			chunkSize = remaining
		}
		_, err = rand.Read(buffer[:chunkSize]) // Fill buffer with random bytes
		if err != nil {
			return fmt.Errorf("error generating random data: %v", err)
		}
		_, err = bufWriter.Write(buffer[:chunkSize])
		if err != nil {
			return fmt.Errorf("error writing to file %s: %v", filePath, err)
		}
		remaining -= chunkSize
	}

	// Flush buffer
	err = bufWriter.Flush()
	if err != nil {
		return fmt.Errorf("error flushing buffer to file %s: %v", filePath, err)
	}

	return nil
}

// worker represents a worker for creating random files.
func worker(id int, jobs <-chan string, fileSize int, bufferSize int, wg *sync.WaitGroup, progress chan<- int64) {
	defer wg.Done()
	for filePath := range jobs {
		start := time.Now()
		err := createRandomFile(filePath, fileSize, bufferSize)
		if err != nil {
			fmt.Printf("Worker %d: %v\n", id, err)
		}
		elapsed := time.Since(start)
		progress <- int64(fileSize) // Send the size of the file through the progress channel
		progress <- int64(elapsed)  // Send the elapsed time through the progress channel
	}
}

// createRandomDataFiles creates random data files based on the provided parameters.
func createRandomDataFiles(directory string, startNum, endNum, fileSize, filesPerDir, bufferSize, numWorkers int, noSubdirs bool) {
	jobs := make(chan string, numWorkers*2)
	progress := make(chan int64, numWorkers*2)
	var wg sync.WaitGroup

	startTime := time.Now()
	totalFiles := endNum - startNum + 1

	fmt.Printf("Starting file creation with %d workers at %s\n", numWorkers, startTime.Format(time.RFC3339))

	// Start progress monitor
	go func() {
		completed := 0
		totalBytes := int64(0)
		totalTime := int64(0)

		for {
			select {
			case size, ok := <-progress:
				if !ok {
					return
				}
				timeTaken := <-progress
				totalBytes += size
				totalTime += timeTaken
				completed++
				mbRate := float64(totalBytes) / (1024 * 1024) / (float64(totalTime) / float64(time.Second))
				fmt.Printf("\rProgress: %d/%d files created in directory %s (%.2f%%) - Bit Rate: %.2f MBps", completed, totalFiles, filepath.Dir(directory), float64(completed)/float64(totalFiles)*100, mbRate)
			}
		}
	}()

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, fileSize, bufferSize, &wg, progress)
	}

	// Generate file paths and send them to workers
	go func() {
		defer close(jobs)
		for i := startNum; i <= endNum; i++ {
			var filePath string
			if noSubdirs {
				filePath = fmt.Sprintf("%s/file_%d.bin", directory, i)
			} else {
				subdirNum := i / filesPerDir
				subDir := fmt.Sprintf("%s/subdir_%d", directory, subdirNum)
				filePath = fmt.Sprintf("%s/file_%d.bin", subDir, i)
			}
			jobs <- filePath
		}
	}()

	wg.Wait()
	close(progress)

	endTime := time.Now()
	time.Sleep(1 * time.Second)
	fmt.Printf("\nFinished file creation at %s\n", endTime.Format(time.RFC3339))
	fmt.Printf("Total time taken: %s\n", endTime.Sub(startTime))
}

// Add this function after the imports
func getOptimalWorkerCount() int {
	cpus := runtime.NumCPU()

	// For I/O bound operations, optimal workers is typically CPU count + 1
	// This accounts for I/O wait times
	workers := cpus + 1

	// Ensure at least 2 workers
	if workers < 2 {
		return 2
	}

	return workers
}

// Add this helper function after imports
func humanReadableSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	// Command-line flags
	bufferSize := getOptimalBufferSize()
	directoryPtr := flag.String("directory", "", "Root Directory where sub-directories and files will be created")
	startNumPtr := flag.Int("start", 0, "Starting number of files")
	endNumPtr := flag.Int("end", 0, "Ending number of files")
	sizePtr := flag.String("size", "", "Size of each file. Supported formats are B, KB, MB, GB (e.g., '1 GB')")
	filesPerDirPtr := flag.Int("files-per-dir", defaultFilesPerDir, "Number of files per subdirectory")
	numWorkersPtr := flag.Int("workers", getOptimalWorkerCount(), "Number of workers - Default is number of CPUs")
	noSubdirsPtr := flag.Bool("no-subdirs", false, "Disable the creation of subdirectories")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "FileForge v%s - High Performance File Generator\n\n", getVersion())
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Buffer Size: %s (auto-optimized)\n", humanReadableSize(bufferSize))
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Check if required options are provided
	if *directoryPtr == "" || *startNumPtr == 0 || *endNumPtr == 0 || *sizePtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Parse file size
	fileSize, err := parseSize(*sizePtr)
	if err != nil || fileSize <= 0 {
		fmt.Fprintf(os.Stderr, "Error parsing file size: %v\n", err)
		os.Exit(1)
	}

	// Start file creation process
	createRandomDataFiles(*directoryPtr, *startNumPtr, *endNumPtr, fileSize, *filesPerDirPtr, bufferSize, *numWorkersPtr, *noSubdirsPtr)
}
