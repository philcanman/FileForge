# FileForge

FileForge is a high-performance CLI tool designed to generate files for testing various IT solutions including:

- Backup systems
- Storage performance benchmarking
- File system stress testing
- Data transfer validations

## Features

- Parallel file generation using multiple workers
- Configurable file sizes (B, KB, MB, GB)
- Customizable directory structure
- Auto-optimized buffer sizes for maximum performance
- Cross-platform support (Windows and Linux)

## Installation

1. Download the latest release for your platform:
   - `FileForge_windows.exe` for Windows
   - `FileForge_linux` for Linux
2. Place the binary in your desired location
3. (Linux only) Make the file executable: `chmod +x FileForge_linux`

## Usage

```text
Usage: FileForge [options]
Options:
  -directory string
        Root Directory where sub-directories and files will be created
  -end int
        Ending number of files
  -files-per-dir int
        Number of files per subdirectory (default 10000)
  -no-subdirs
        Disable the creation of subdirectories
  -size string
        Size of each file. Supported formats are B, KB, MB, GB (e.g., '1 GB')
  -start int
        Starting number of files
  -workers int
        Number of workers - Default is number of CPUs
```

### Example 1: Generate 10 files that are 1GB each

```console
./FileForge_linux -directory='.exampleDIR' -start=1 -end=10 -size='1 GB'
```

Output:

```text
Starting file creation with 32 workers at 2024-06-10T15:22:02-04:00
Progress: 10/10 files created (100.00%)

Finished file creation at 2024-06-10T15:22:19-04:00
Total time taken: 16.462541903s
```
