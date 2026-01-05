//go:build windows

package storage

// lockFile is a no-op on Windows.
// Windows has different file locking semantics (LockFileEx),
// but for this use case we skip locking on Windows.
func lockFile(fd uintptr) error {
	return nil
}

// unlockFile is a no-op on Windows.
func unlockFile(fd uintptr) error {
	return nil
}
