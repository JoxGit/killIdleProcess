//go:build windows
// +build windows

package ps

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// Windows API functions
var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
)

// Some constants from the Windows API
const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
)

// PROCESSENTRY32 is the Windows API structure that contains a process's
// information.
type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

// WindowsProcess is an implementation of Process for Windows.
type WindowsProcess struct {
	pid  int
	ppid int
	exe  string
}

func (p *WindowsProcess) Pid() int {
	return p.pid
}

func (p *WindowsProcess) PPid() int {
	return p.ppid
}

func (p *WindowsProcess) Executable() string {
	return p.exe
}

func (p *WindowsProcess) CPUTime() (CPUTimes, error) {
	handle, err := getProcessHandle(p.pid)
	if err != nil {
		return CPUTimes{}, err
	}
	defer syscall.CloseHandle(handle)

	var creationTime, exitTime, kernelTime, userTime syscall.Filetime
	if err := syscall.GetProcessTimes(handle, &creationTime, &exitTime, &kernelTime, &userTime); err != nil {
		return CPUTimes{}, err
	}

	return CPUTimes{
		User:   filetimeToDuration(&userTime),
		System: filetimeToDuration(&kernelTime),
	}, nil
}

func (p *WindowsProcess) Kill() error {
	process, err := os.FindProcess(p.pid)

	if err != nil {
		return err
	}

	err = process.Kill()

	if err != nil {
		return err
	}

	return nil
}

func newWindowsProcess(e *PROCESSENTRY32) *WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return &WindowsProcess{
		pid:  int(e.ProcessID),
		ppid: int(e.ParentProcessID),
		exe:  syscall.UTF16ToString(e.ExeFile[:end]),
	}
}

func findProcess(pid int) (Process, error) {
	ps, err := processes()
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		if p.Pid() == pid {
			return p, nil
		}
	}

	return nil, nil
}

func processes() ([]Process, error) {
	handle, _, _ := procCreateToolhelp32Snapshot.Call(
		0x00000002,
		0)
	if handle < 0 {
		return nil, syscall.GetLastError()
	}
	defer procCloseHandle.Call(handle)

	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("Error retrieving process info.")
	}

	results := make([]Process, 0, 50)
	for {
		results = append(results, newWindowsProcess(&entry))

		ret, _, _ := procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return results, nil
}

func filterProcesses(name string) ([]Process, error) {
	processes, err := processes()

	if err != nil {
		return nil, err
	}

	var filtered []Process

	for _, v := range processes {
		if v.Executable() == name {
			filtered = append(filtered, v)
		}
	}

	return filtered, err
}

func getProcessHandle(pid int) (handle syscall.Handle, err error) {

	// Try different access rights, from broader to more limited.
	// PROCESS_VM_READ is needed to get command-line and working directory
	// PROCESS_QUERY_LIMITED_INFORMATION is only available in Vista+
	for _, permissions := range [4]uint32{
		syscall.PROCESS_QUERY_INFORMATION,
	} {
		if handle, err = syscall.OpenProcess(permissions, false, uint32(pid)); err == nil {
			break
		}
	}
	return handle, err
}

// FiletimeToDuration converts a Filetime to a time.Duration. Do not use this
// method to convert a Filetime to an actual clock time, for that use
// Filetime.Nanosecond().
func filetimeToDuration(ft *syscall.Filetime) time.Duration {
	n := int64(ft.HighDateTime)<<32 + int64(ft.LowDateTime) // in 100-nanosecond intervals
	return time.Duration(n * 100)
}
