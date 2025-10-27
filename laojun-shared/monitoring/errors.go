package monitoring

import "errors"

var (
	// ErrMonitorNotInitialized is returned when monitor is not initialized.
	ErrMonitorNotInitialized = errors.New("monitor not initialized")
	
	// ErrMetricNotFound is returned when metric is not found.
	ErrMetricNotFound = errors.New("metric not found")
	
	// ErrMetricAlreadyExists is returned when metric already exists.
	ErrMetricAlreadyExists = errors.New("metric already exists")
	
	// ErrInvalidMetricName is returned when metric name is invalid.
	ErrInvalidMetricName = errors.New("invalid metric name")
	
	// ErrInvalidMetricValue is returned when metric value is invalid.
	ErrInvalidMetricValue = errors.New("invalid metric value")
	
	// ErrInvalidMetricType is returned when metric type is invalid.
	ErrInvalidMetricType = errors.New("invalid metric type")
	
	// ErrExportFailed is returned when export operation fails.
	ErrExportFailed = errors.New("export failed")
	
	// ErrExportFormatNotSupported is returned when export format is not supported.
	ErrExportFormatNotSupported = errors.New("export format not supported")
	
	// ErrTimerNotStarted is returned when timer is not started.
	ErrTimerNotStarted = errors.New("timer not started")
	
	// ErrTimerAlreadyStopped is returned when timer is already stopped.
	ErrTimerAlreadyStopped = errors.New("timer already stopped")
	
	// ErrStorageNotEnabled is returned when storage is not enabled.
	ErrStorageNotEnabled = errors.New("storage not enabled")
	
	// ErrStorageFull is returned when storage is full.
	ErrStorageFull = errors.New("storage full")
)

// IsNotFound checks if the error is a "not found" error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrMetricNotFound)
}

// IsMonitorNotInitialized checks if the error is a monitor not initialized error.
func IsMonitorNotInitialized(err error) bool {
	return errors.Is(err, ErrMonitorNotInitialized)
}

// TODO: Add more error checking functions as needed
