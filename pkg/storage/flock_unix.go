//go:build unix

package storage

import "syscall"

// lockFile acquires an exclusive lock on the file descriptor.
func lockFile(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_EX)
}

// unlockFile releases the lock on the file descriptor.
func unlockFile(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_UN)
}
