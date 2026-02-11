package store

import (
	"fmt"
	"io"
	"os"
	"time"
)

func replaceDatabaseFile(dbPath, tempPath string) error {
	if err := removeSQLiteSidecars(dbPath); err != nil {
		return err
	}

	stagingOldPath := fmt.Sprintf("%s.replaced.%d", dbPath, time.Now().UnixNano())
	hadOriginal := false
	if err := os.Rename(dbPath, stagingOldPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stage old database before replacement: %w", err)
		}
	} else {
		hadOriginal = true
	}

	if err := os.Rename(tempPath, dbPath); err != nil {
		if hadOriginal {
			_ = os.Rename(stagingOldPath, dbPath)
		}
		return fmt.Errorf("rename migrated database into place: %w", err)
	}

	if hadOriginal {
		if err := os.Remove(stagingOldPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove staged old database: %w", err)
		}
	}

	if err := removeSQLiteSidecars(dbPath); err != nil {
		return err
	}

	return nil
}

func removeSQLiteSidecars(dbPath string) error {
	for _, suffix := range []string{"-wal", "-shm", "-journal"} {
		path := dbPath + suffix
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove sidecar file %s: %w", path, err)
		}
	}
	return nil
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return fmt.Errorf("copy data: %w", err)
	}

	if err := dst.Sync(); err != nil {
		_ = dst.Close()
		return fmt.Errorf("sync destination: %w", err)
	}

	if err := dst.Close(); err != nil {
		return fmt.Errorf("close destination: %w", err)
	}

	return nil
}
