package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarning
	LogError
	LogProgress // Special level for progress updates
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarning:
		return "WARNING"
	case LogError:
		return "ERROR"
	case LogProgress:
		return "PROGRESS"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source,omitempty"`
	UserVisible bool    `json:"user_visible"`
}

// LogFilter defines rules for filtering log output
type LogFilter struct {
	MinLevel        LogLevel
	IncludeSources  []string
	ExcludePatterns []*regexp.Regexp
	UserFriendly    bool // Transform technical messages to user-friendly ones
}

// Logger provides structured logging with user/file separation
type Logger struct {
	mu           sync.Mutex
	fileWriter   io.Writer
	userWriter   io.Writer
	filter       LogFilter
	entries      []LogEntry
	maxEntries   int
	progressLine string // Current progress message (for updates)

	// Patterns to filter out from user output (noisy Docker output)
	noisePatterns []*regexp.Regexp

	// Patterns to transform to user-friendly messages
	transformPatterns map[*regexp.Regexp]string
}

// NewLogger creates a new logger with default settings
func NewLogger(fileWriter, userWriter io.Writer) *Logger {
	l := &Logger{
		fileWriter: fileWriter,
		userWriter: userWriter,
		filter: LogFilter{
			MinLevel:     LogInfo,
			UserFriendly: true,
		},
		maxEntries: 1000,
		noisePatterns: []*regexp.Regexp{
			// Docker pull progress
			regexp.MustCompile(`^[a-f0-9]+: (Pulling|Waiting|Downloading|Extracting|Pull complete|Already exists)`),
			// Docker layer hashes
			regexp.MustCompile(`^[a-f0-9]{12}$`),
			// Docker digest
			regexp.MustCompile(`^Digest: sha256:`),
			// Status messages during pull
			regexp.MustCompile(`^Status: Downloaded`),
			regexp.MustCompile(`^Status: Image is up to date`),
			// Docker network creation noise
			regexp.MustCompile(`^Creating network`),
			// Volume creation noise
			regexp.MustCompile(`^Creating volume`),
			// Container creation (keep start/stop messages)
			regexp.MustCompile(`^Container [a-f0-9]+ Creating$`),
			// Empty lines
			regexp.MustCompile(`^\s*$`),
		},
		transformPatterns: map[*regexp.Regexp]string{
			// Transform pull messages
			regexp.MustCompile(`Pulling from (.+)`):     "Downloading image: $1",
			regexp.MustCompile(`Container (.+) Started`): "Started: $1",
			regexp.MustCompile(`Container (.+) Stopped`): "Stopped: $1",
			regexp.MustCompile(`Container (.+) Running`): "Running: $1",
		},
	}
	return l
}

// SetMinLevel sets the minimum log level for user output
func (l *Logger) SetMinLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filter.MinLevel = level
}

// SetVerbose enables verbose (debug) logging
func (l *Logger) SetVerbose(verbose bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if verbose {
		l.filter.MinLevel = LogDebug
		l.filter.UserFriendly = false
	} else {
		l.filter.MinLevel = LogInfo
		l.filter.UserFriendly = true
	}
}

// Log writes a log entry
func (l *Logger) Log(level LogLevel, source, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		Source:      source,
		UserVisible: level >= l.filter.MinLevel && !l.isNoise(message),
	}

	// Store entry
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[1:]
	}

	// Always write to file
	if l.fileWriter != nil {
		fmt.Fprintf(l.fileWriter, "[%s] [%s] [%s] %s\n",
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.Level.String(),
			entry.Source,
			entry.Message)
	}

	// Write to user if visible
	if entry.UserVisible && l.userWriter != nil {
		msg := message
		if l.filter.UserFriendly {
			msg = l.transformMessage(message)
		}

		// Format based on level
		switch level {
		case LogError:
			fmt.Fprintf(l.userWriter, "\033[31m[ERROR]\033[0m %s\n", msg)
		case LogWarning:
			fmt.Fprintf(l.userWriter, "\033[33m[WARN]\033[0m  %s\n", msg)
		case LogInfo:
			fmt.Fprintf(l.userWriter, "\033[34m[INFO]\033[0m  %s\n", msg)
		case LogDebug:
			fmt.Fprintf(l.userWriter, "\033[90m[DEBUG]\033[0m %s\n", msg)
		case LogProgress:
			// Progress uses carriage return for in-place updates
			fmt.Fprintf(l.userWriter, "\r\033[K%s", msg)
		}
	}
}

// Progress logs a progress update (updates in place)
func (l *Logger) Progress(source, message string) {
	l.mu.Lock()
	l.progressLine = message
	l.mu.Unlock()

	l.Log(LogProgress, source, message)
}

// ProgressDone completes a progress line
func (l *Logger) ProgressDone() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.userWriter != nil && l.progressLine != "" {
		fmt.Fprintln(l.userWriter) // Move to next line
		l.progressLine = ""
	}
}

// Debug logs a debug message
func (l *Logger) Debug(source, message string) {
	l.Log(LogDebug, source, message)
}

// Info logs an info message
func (l *Logger) Info(source, message string) {
	l.Log(LogInfo, source, message)
}

// Warning logs a warning message
func (l *Logger) Warning(source, message string) {
	l.Log(LogWarning, source, message)
}

// Error logs an error message
func (l *Logger) Error(source, message string) {
	l.Log(LogError, source, message)
}

// isNoise checks if a message should be filtered out
func (l *Logger) isNoise(message string) bool {
	for _, pattern := range l.noisePatterns {
		if pattern.MatchString(message) {
			return true
		}
	}
	return false
}

// transformMessage transforms a technical message to a user-friendly one
func (l *Logger) transformMessage(message string) string {
	for pattern, replacement := range l.transformPatterns {
		if pattern.MatchString(message) {
			return pattern.ReplaceAllString(message, replacement)
		}
	}
	return message
}

// GetEntries returns stored log entries
func (l *Logger) GetEntries(level LogLevel) []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []LogEntry
	for _, entry := range l.entries {
		if entry.Level >= level {
			result = append(result, entry)
		}
	}
	return result
}

// GetUserEntries returns only user-visible entries
func (l *Logger) GetUserEntries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []LogEntry
	for _, entry := range l.entries {
		if entry.UserVisible {
			result = append(result, entry)
		}
	}
	return result
}

// StreamFilter wraps a reader and filters output in real-time
type StreamFilter struct {
	logger *Logger
	source string
	reader io.Reader
}

// NewStreamFilter creates a filter for streaming output
func (l *Logger) NewStreamFilter(source string, reader io.Reader) *StreamFilter {
	return &StreamFilter{
		logger: l,
		source: source,
		reader: reader,
	}
}

// Process reads from the stream and logs filtered output
func (sf *StreamFilter) Process() {
	scanner := bufio.NewScanner(sf.reader)

	// Track progress for multi-line operations like docker pull
	var pullProgress = make(map[string]bool)
	progressShown := false

	for scanner.Scan() {
		line := scanner.Text()

		// Detect pull progress and consolidate
		if strings.Contains(line, ": Pulling") || strings.Contains(line, ": Downloading") {
			// Extract image name
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				pullProgress[parts[0]] = true
				if !progressShown {
					sf.logger.Progress(sf.source, fmt.Sprintf("Pulling images... (%d layers)", len(pullProgress)))
					progressShown = true
				} else {
					sf.logger.Progress(sf.source, fmt.Sprintf("Pulling images... (%d layers)", len(pullProgress)))
				}
			}
			continue
		}

		// Detect pull complete
		if strings.Contains(line, ": Pull complete") || strings.Contains(line, ": Already exists") {
			continue
		}

		// End of pull operation
		if strings.HasPrefix(line, "Digest:") || strings.HasPrefix(line, "Status:") {
			if progressShown {
				sf.logger.ProgressDone()
				progressShown = false
				pullProgress = make(map[string]bool)
			}
			// Show status message
			if strings.HasPrefix(line, "Status:") {
				sf.logger.Info(sf.source, line)
			}
			continue
		}

		// Container lifecycle events - always show
		if strings.Contains(line, "Container") && (strings.Contains(line, "Started") ||
			strings.Contains(line, "Stopped") || strings.Contains(line, "Created")) {
			sf.logger.Info(sf.source, line)
			continue
		}

		// Error messages - always show
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			sf.logger.Error(sf.source, line)
			continue
		}

		// Warning messages
		if strings.Contains(strings.ToLower(line), "warning") ||
			strings.Contains(strings.ToLower(line), "warn") {
			sf.logger.Warning(sf.source, line)
			continue
		}

		// Everything else goes to debug
		sf.logger.Debug(sf.source, line)
	}

	// Ensure progress is closed
	if progressShown {
		sf.logger.ProgressDone()
	}
}

// DockerOutputFilter provides specialized filtering for docker-compose output
type DockerOutputFilter struct {
	logger     *Logger
	source     string
	showPull   bool // Show pull progress
	showBuild  bool // Show build progress
}

// NewDockerOutputFilter creates a docker-compose specific filter
func (l *Logger) NewDockerOutputFilter(source string) *DockerOutputFilter {
	return &DockerOutputFilter{
		logger:    l,
		source:    source,
		showPull:  false, // Default to summarized pull
		showBuild: true,  // Show build steps
	}
}

// SetShowPull enables/disables detailed pull output
func (f *DockerOutputFilter) SetShowPull(show bool) {
	f.showPull = show
}

// FilterLine processes a single line of docker-compose output
func (f *DockerOutputFilter) FilterLine(line string) {
	// Detect the type of output and handle appropriately

	// Pull progress - summarize unless verbose
	if strings.Contains(line, "Pulling") && !f.showPull {
		// Only show start of pull
		if strings.HasSuffix(line, "Pulling") {
			f.logger.Info(f.source, line)
		}
		return
	}

	// Layer-level progress - skip unless verbose
	if matched, _ := regexp.MatchString(`^[a-f0-9]+:`, line); matched && !f.showPull {
		return
	}

	// Build progress
	if strings.HasPrefix(line, "Step ") || strings.Contains(line, "---") {
		if f.showBuild {
			f.logger.Info(f.source, line)
		} else {
			f.logger.Debug(f.source, line)
		}
		return
	}

	// Container events - always show
	if strings.Contains(line, "Container") {
		f.logger.Info(f.source, line)
		return
	}

	// Network/volume events - debug only
	if strings.Contains(line, "Network") || strings.Contains(line, "Volume") {
		f.logger.Debug(f.source, line)
		return
	}

	// Errors - always show
	if strings.Contains(strings.ToLower(line), "error") {
		f.logger.Error(f.source, line)
		return
	}

	// Default - debug
	f.logger.Debug(f.source, line)
}

// CreateLogFile creates a log file with proper permissions
func CreateLogFile(path string) (*os.File, error) {
	// Ensure directory exists
	dir := strings.TrimSuffix(path, "/"+strings.Split(path, "/")[len(strings.Split(path, "/"))-1])
	if dir != "" && dir != path {
		os.MkdirAll(dir, 0755)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

// UserMessage represents a user-facing message with severity
type UserMessage struct {
	Level   LogLevel
	Message string
	Details string
}

// StandardMessages provides canned messages for common operations
var StandardMessages = map[string]UserMessage{
	"docker_pull_start": {
		Level:   LogInfo,
		Message: "Downloading container images...",
		Details: "This may take a few minutes on first run",
	},
	"docker_pull_done": {
		Level:   LogInfo,
		Message: "Container images ready",
	},
	"services_starting": {
		Level:   LogInfo,
		Message: "Starting services...",
	},
	"services_started": {
		Level:   LogInfo,
		Message: "All services started successfully",
	},
	"port_conflict": {
		Level:   LogWarning,
		Message: "Port conflict detected",
		Details: "An existing service is using the requested port",
	},
	"migration_needed": {
		Level:   LogInfo,
		Message: "Existing installation detected",
		Details: "Your data will be preserved during upgrade",
	},
	"health_check_pass": {
		Level:   LogInfo,
		Message: "Health check passed",
	},
	"health_check_fail": {
		Level:   LogError,
		Message: "Health check failed",
		Details: "Check the logs for more details",
	},
}
