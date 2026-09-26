package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const evGrab = 0x40044590 // Linux EVIOCGRAB; closed descriptors release the grab.
type event struct {
	kind, code uint16
	value      int32
}

func device(name string) (string, error) {
	paths, _ := filepath.Glob("/sys/class/input/event*/device/name")
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil && strings.TrimSpace(string(b)) == name {
			return "/dev/input/" + filepath.Base(filepath.Dir(filepath.Dir(p))), nil
		}
	}
	return "", fmt.Errorf("input device unavailable: %s", name)
}

func readEvents(f *os.File, out chan<- event, failed chan<- error) {
	var b [24]byte // ARM64 input_event: timeval + type/code/value.
	for {
		if _, err := io.ReadFull(f, b[:]); err != nil {
			failed <- err
			return
		}
		if binary.LittleEndian.Uint16(b[16:18]) == 1 {
			out <- event{1, binary.LittleEndian.Uint16(b[18:20]), int32(binary.LittleEndian.Uint32(b[20:24]))}
		}
	}
}

func run(dir string) error {
	lock, err := os.OpenFile(filepath.Join(dir, "guard.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return err
	}
	keyPath, err := device("Microsoft X-box 360 pad")
	if err != nil {
		return err
	}
	touchPath, err := device("sitronix_ts_i2c")
	if err != nil {
		return err
	}
	key, err := os.Open(keyPath)
	if err != nil {
		return err
	}
	defer key.Close()
	touch, err := os.Open(touchPath)
	if err != nil {
		return err
	}
	defer touch.Close()
	statePath := filepath.Join(dir, "state")
	writeState := func(s string) { _ = os.WriteFile(statePath, []byte(s+"\n"), 0600) }
	writeState("enabled")
	defer writeState("enabled (service stopped)")
	pidPath := filepath.Join(dir, "guard.pid")
	if err = os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		return err
	}
	defer os.Remove(pidPath)
	fmt.Printf("MODE=%s TOUCH=%s\n", keyPath, touchPath)
	keys := make(chan event, 16)
	touches := make(chan event, 16)
	failed := make(chan error, 2)
	go readEvents(key, keys, failed)
	go readEvents(touch, touches, failed) // Drain touch events while grabbed.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(stop)
	var deadline <-chan time.Time
	var timer *time.Timer
	var pressed, disabled, touching, pending bool
	// Probe mode exits automatically, releasing touch even if testing is interrupted.
	var expiry <-chan time.Time
	if len(os.Args) > 2 && os.Args[2] == "--test" {
		expiry = time.After(90 * time.Second)
	}
	toggle := func() error {
		var value uintptr = 1
		if disabled {
			value = 0
		}
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, touch.Fd(), evGrab, value)
		if errno != 0 {
			return errno
		}
		disabled = !disabled
		if disabled {
			writeState("disabled")
		} else {
			writeState("enabled")
		}
		return nil
	}
	for {
		select {
		case <-stop:
			return nil
		case <-expiry:
			return nil
		case err := <-failed:
			return fmt.Errorf("input disconnected: %w", err)
		case e := <-touches:
			if e.code == 330 { // BTN_TOUCH: never grab a finger mid-gesture.
				touching = e.value != 0
				if !touching && pending {
					pending = false
					if err := toggle(); err != nil {
						return err
					}
				}
			}
		case e := <-keys:
			if e.code != 316 {
				continue
			} // BTN_MODE
			if e.value == 1 && !pressed {
				pressed = true
				timer = time.NewTimer(2 * time.Second)
				deadline = timer.C
			} else if e.value == 0 {
				pressed = false
				pending = false
				if timer != nil {
					timer.Stop()
				}
				deadline = nil
			}
		case <-deadline:
			deadline = nil
			if pressed {
				if touching {
					pending = true
				} else if err := toggle(); err != nil {
					return err
				}
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: kpa_touch_guard MODULE_DIR [--test]")
		os.Exit(2)
	}
	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
