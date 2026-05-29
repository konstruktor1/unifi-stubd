package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"golang.org/x/sys/unix"
)

const (
	instanceGuardFail = "fail"
	instanceGuardWarn = "warn"
	instanceGuardOff  = "off"
)

type instanceGuardHandle struct {
	file *os.File
}

type instanceGuardMetadata struct {
	PID       int    `json:"pid"`
	StartTime string `json:"start_time"`
	Profile   string `json:"profile"`
	MAC       string `json:"mac"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
}

// acquireInstanceGuard holds a non-blocking advisory lock for live daemon
// runs. The kernel releases the lock automatically if the process exits.
func acquireInstanceGuard(flags runtimeFlags, profile device.Profile, macText, hostname string) (*instanceGuardHandle, error) {
	mode := normalizeInstanceGuard(flags.instanceGuard)
	if mode == instanceGuardOff {
		return nil, nil
	}
	path := strings.TrimSpace(flags.instanceGuardPath)
	if path == "" {
		return nil, fmt.Errorf("instance_guard_path must not be empty when instance_guard is %s", mode)
	}
	handle, err := openInstanceGuard(path)
	if err != nil {
		if mode == instanceGuardWarn {
			log.Printf("instance guard warning: %v", err)
			return nil, nil
		}
		return nil, err
	}
	if err := writeInstanceGuardMetadata(handle.file, profile, macText, hostname); err != nil {
		closeErr := handle.Close()
		if mode == instanceGuardWarn {
			log.Printf("instance guard warning: %v", err)
			if closeErr != nil {
				log.Printf("instance guard close warning: %v", closeErr)
			}
			return nil, nil
		}
		if closeErr != nil {
			return nil, fmt.Errorf("%w; close lock: %w", err, closeErr)
		}
		return nil, err
	}
	log.Printf("instance guard active: mode=%s path=%s", mode, path)
	return handle, nil
}

func openInstanceGuard(path string) (*instanceGuardHandle, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create instance guard directory %s: %w", dir, err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open instance guard %s: %w", path, err)
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return nil, instanceGuardConflict(path)
		}
		return nil, fmt.Errorf("lock instance guard %s: %w", path, err)
	}
	return &instanceGuardHandle{file: file}, nil
}

func instanceGuardConflict(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("another unifi-stubd instance is already running; instance guard %s is locked and metadata could not be read: %w", path, err)
	}
	metadata := strings.TrimSpace(string(data))
	if metadata == "" {
		return fmt.Errorf("another unifi-stubd instance is already running; instance guard %s is locked", path)
	}
	return fmt.Errorf("another unifi-stubd instance is already running; instance guard %s is locked; metadata: %s", path, metadata)
}

func writeInstanceGuardMetadata(file *os.File, profile device.Profile, macText, hostname string) error {
	metadata := instanceGuardMetadata{
		PID:       os.Getpid(),
		StartTime: time.Now().UTC().Format(time.RFC3339),
		Profile:   profile.Name,
		MAC:       macText,
		Hostname:  hostname,
		Version:   version,
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal instance guard metadata: %w", err)
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate instance guard metadata: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek instance guard metadata: %w", err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write instance guard metadata: %w", err)
	}
	return nil
}

func (h *instanceGuardHandle) Close() error {
	if h == nil || h.file == nil {
		return nil
	}
	lockErr := unix.Flock(int(h.file.Fd()), unix.LOCK_UN)
	closeErr := h.file.Close()
	h.file = nil
	if lockErr != nil {
		return fmt.Errorf("unlock instance guard: %w", lockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close instance guard: %w", closeErr)
	}
	return nil
}

func normalizeInstanceGuard(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return instanceGuardFail
	}
	return value
}

func validateInstanceGuard(flags *runtimeFlags) error {
	flags.instanceGuard = normalizeInstanceGuard(flags.instanceGuard)
	switch flags.instanceGuard {
	case instanceGuardFail, instanceGuardWarn, instanceGuardOff:
	default:
		return fmt.Errorf("invalid -instance-guard %q; use fail, warn, or off", flags.instanceGuard)
	}
	if flags.instanceGuard != instanceGuardOff && strings.TrimSpace(flags.instanceGuardPath) == "" {
		return fmt.Errorf("instance_guard_path must not be empty when instance_guard is %s", flags.instanceGuard)
	}
	return nil
}
