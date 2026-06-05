package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileMigrationResult describes an in-place config migration attempt.
type FileMigrationResult struct {
	Result     MigrationResult
	BackupPath string
}

// MigrateFile migrates one config file in place after writing a timestamped
// backup. Dry-run mode only returns the candidate document.
func MigrateFile(path string, dryRun bool) (FileMigrationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileMigrationResult{}, fmt.Errorf("read config %s: %w", path, err)
	}
	result, err := MigrateData(data)
	if err != nil {
		return FileMigrationResult{}, err
	}
	out := FileMigrationResult{Result: result}
	if dryRun || !result.Changed {
		return out, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return FileMigrationResult{}, fmt.Errorf("stat config %s: %w", path, err)
	}
	backupPath, err := writeMigrationBackup(path, data, info.Mode().Perm())
	if err != nil {
		return FileMigrationResult{}, err
	}
	if err := writeMigrationFile(path, result.Data, info.Mode().Perm()); err != nil {
		return FileMigrationResult{}, err
	}
	out.BackupPath = backupPath
	return out, nil
}

func writeMigrationBackup(path string, data []byte, mode os.FileMode) (string, error) {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	for attempt := 0; attempt < 100; attempt++ {
		suffix := stamp
		if attempt > 0 {
			suffix = fmt.Sprintf("%s.%d", stamp, attempt)
		}
		backupPath := path + ".bak." + suffix
		err := writeExclusiveFile(backupPath, data, mode)
		if err == nil {
			return backupPath, nil
		}
		if !os.IsExist(err) {
			return "", fmt.Errorf("write config backup %s: %w", backupPath, err)
		}
	}
	return "", fmt.Errorf("write config backup %s.bak.%s: too many existing backups", path, stamp)
}

func writeMigrationFile(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	temp, err := os.CreateTemp(dir, "."+base+".tmp.")
	if err != nil {
		return fmt.Errorf("create migrated config temp file: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write migrated config temp file: %w", err)
	}
	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return fmt.Errorf("chmod migrated config temp file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close migrated config temp file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace config %s: %w", path, err)
	}
	return nil
}

func writeExclusiveFile(path string, data []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}
