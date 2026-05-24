package jarvis

import (
	"os"
	"syscall"
)

func mountBootPart(dev, path string) error {
	var flags uintptr
	flags = syscall.MS_NOATIME | syscall.MS_SILENT
	flags |= syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID
	err := syscall.Mount(dev, path, "vfat", flags, "")
	return os.NewSyscallError("mount", err)
}

func unmountBootPart(path string) error {
	err := syscall.Unmount(path, 0)
	return os.NewSyscallError("unmount", err)
}
