//go:build windows

package beatify

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	stdOutputHandle = uint32(0xFFFFFFF5) // STD_OUTPUT_HANDLE (-11)
	stdErrorHandle  = uint32(0xFFFFFFF4) // STD_ERROR_HANDLE (-12)
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle    = kernel32.NewProc("GetStdHandle")
	procSetStdHandle    = kernel32.NewProc("SetStdHandle")
	procDuplicateHandle = kernel32.NewProc("DuplicateHandle")
	procGetCurrentProc  = kernel32.NewProc("GetCurrentProcess")
)

// redirectStdio duplicates the real stdout handle, then points the process
// standard output/error handles at NUL so testing chatter is discarded.
func redirectStdio() (*os.File, error) {
	stdoutHandle, _, err := procGetStdHandle.Call(uintptr(stdOutputHandle))
	if stdoutHandle == 0 {
		return nil, err
	}

	proc, _, err := procGetCurrentProc.Call()
	if proc == 0 {
		return nil, err
	}

	var dupHandle syscall.Handle
	r, _, err := procDuplicateHandle.Call(
		proc,
		stdoutHandle,
		proc,
		uintptr(unsafe.Pointer(&dupHandle)),
		0,
		1, // bInheritHandle
		2, // DUPLICATE_SAME_ACCESS
	)
	if r == 0 {
		return nil, err
	}

	saved := os.NewFile(uintptr(dupHandle), "stdout")

	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		_ = saved.Close()

		return nil, err
	}

	nullHandle := null.Fd()
	if r, _, err := procSetStdHandle.Call(uintptr(stdOutputHandle), nullHandle); r == 0 {
		_ = null.Close()
		_ = saved.Close()

		return nil, err
	}

	if r, _, err := procSetStdHandle.Call(uintptr(stdErrorHandle), nullHandle); r == 0 {
		_ = null.Close()
		_ = saved.Close()

		return nil, err
	}

	os.Stdout = null
	os.Stderr = null

	return saved, nil
}
