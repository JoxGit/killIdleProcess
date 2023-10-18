// ps provides an API for finding and listing processes in a platform-agnostic
// way.
//
// NOTE: If you're reading these docs online via GoDocs or some other system,
// you might only see the Unix docs. This project makes heavy use of
// platform-specific implementations. We recommend reading the source if you
// are interested.
package ps

import "time"

// Process is the generic interface that is implemented on every platform
// and provides common operations for processes.
type Process interface {
	// Pid is the process ID for this process.
	Pid() int

	// PPid is the parent process ID for this process.
	PPid() int

	// Executable name running this process. This is not a path to the
	// executable.
	Executable() string

	//CPUTimes represents time spent on porcess
	CPUTime() (CPUTimes, error)

	//Kill is used to kill process
	Kill() error
}

//CPUTimes is struct to hold data about times application spent working
type CPUTimes struct {
	//User duration represent time from userTime
	User time.Duration
	//System duration represent time from kernelTime
	System time.Duration
}

// Processes returns all processes.
//
// This of course will be a point-in-time snapshot of when this method was
// called. Some operating systems don't provide snapshot capability of the
// process table, in which case the process table returned might contain
// ephemeral entities that happened to be running when this was called.
func Processes() ([]Process, error) {
	return processes()
}

// FindProcess looks up a single process by pid.
//
// Process will be nil and error will be nil if a matching process is
// not found.
func FindProcess(pid int) (Process, error) {
	return findProcess(pid)
}

//FilterProcesses returns filteres slice of processes withc have executable with gven name
func FilterProcesses(name string) ([]Process, error) {
	return filterProcesses(name)
}
