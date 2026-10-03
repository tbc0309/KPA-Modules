package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLEDWriteFailureRetriesAllChannels(t *testing.T) {
	dir := t.TempDir()
	open := func(name string) *os.File {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
		return f
	}
	red, green, blue := open("red"), open("green"), open("blue")
	// 关闭测试文件模拟通道故障，不操作实际 LED 节点。
	// Simulate a failed channel without changing real LED nodes.
	red.Close()
	c := controller{red: red, green: green, blue: blue, lastR: -1, lastG: -1, lastB: -1}
	c.writeRawLED(12, 34, 56)
	if c.lastR != -1 || c.retryAfter.IsZero() {
		t.Fatal("failed write was cached as successful")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "green"))
	if string(data) != "34\n" {
		t.Fatal("one failed channel skipped the remaining channels")
	}
	c.red = open("recovered-red")
	c.retryAfter = time.Time{}
	c.writeRawLED(12, 34, 56)
	if c.lastR != 12 || c.lastG != 34 || c.lastB != 56 {
		t.Fatal("recovered write did not update cache")
	}
}

func TestEffectsAndScreenOff(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mode, cpu int
		state     string
	}{
		{"powersave", 0, 40, "POWER_SAVE_GREEN_BREATHE"},
		{"balanced", 1, 40, "BALANCED_RAINBOW"},
		{"gaming", 2, 40, "GAME_ELECTRIC_BLUE_BREATHE"},
		{"full", 3, 40, "FULL_POWER_MAGENTA_PULSE"},
		{"warning", 1, 85, "TEMP_WARN_ORANGE_BREATHE"},
		{"danger", 1, 90, "TEMP_DANGER_RED_FAST_BLINK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := make([]*os.File, 3)
			for i := range files {
				f, err := os.CreateTemp(t.TempDir(), "led")
				if err != nil {
					t.Fatal(err)
				}
				files[i] = f
				defer f.Close()
			}
			c := controller{cfg: defaultConfig(), red: files[0], green: files[1], blue: files[2],
				lastR: -1, lastG: -1, lastB: -1, lastState: tc.state,
				mode: tc.mode, cpuTemperature: tc.cpu, capacity: 80, batteryStatus: "Discharging", screenOn: true}
			c.updateEffect()
			if c.lastR+c.lastG+c.lastB <= 0 {
				t.Fatal("effect did not light any channel")
			}
			c.screenOn = false
			c.lastState = "SCREEN_OFF"
			c.updateEffect()
			if c.lastR != 0 || c.lastG != 0 || c.lastB != 0 {
				t.Fatal("screen-off failed to turn LEDs off")
			}
		})
	}
}
