package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	configPath    = "/data/adb/kpa_rgb_control.conf"
	statePath     = "/data/adb/kpa_rgb_control.state"
	logPath       = "/data/adb/kpa_rgb_control.log"
	prefPath      = "/data/user/0/com.ayaneo.home/shared_prefs/com.ayaneo.home_preferences.xml"
	redPath       = "/sys/class/leds/red/brightness"
	greenPath     = "/sys/class/leds/green/brightness"
	bluePath      = "/sys/class/leds/blue/brightness"
	backlightPath = "/sys/class/leds/lcd-backlight/brightness"
)

type config struct {
	maxBrightness     int
	cpuTempWarn       int
	cpuTempDanger     int
	batteryTempWarn   int
	batteryTempDanger int
	batteryLow        int
	interval          time.Duration
}

type controller struct {
	cfg                config
	red                *os.File
	green              *os.File
	blue               *os.File
	lastR              int
	lastG              int
	lastB              int
	lastState          string
	lastErrorLog       time.Time
	mode               int
	capacity           int
	batteryStatus      string
	cpuTemperature     int
	batteryTemperature int
	phase              int
	breathe            int
	tick               int
	screenOn           bool
}

var localZone = time.FixedZone("UTC+8", 8*60*60)
var modeMarker = []byte(`name="fanMode" value="`)

const (
	cpuTempPath     = "/sys/class/thermal/thermal_zone3/temp" // mtktscpu
	batteryTempPath = "/sys/class/thermal/thermal_zone0/temp" // mtktsbattery
)

func defaultConfig() config {
	return config{64, 85, 90, 50, 55, 15, 200 * time.Millisecond}
}

func readConfig() config {
	cfg := defaultConfig()
	f, err := os.Open(configPath)
	if err != nil {
		return cfg
	}
	defer f.Close()
	values := map[string]*int{
		"MAX_BRIGHTNESS":      &cfg.maxBrightness,
		"CPU_TEMP_WARN":       &cfg.cpuTempWarn,
		"CPU_TEMP_DANGER":     &cfg.cpuTempDanger,
		"BATTERY_TEMP_WARN":   &cfg.batteryTempWarn,
		"BATTERY_TEMP_DANGER": &cfg.batteryTempDanger,
		"BATTERY_LOW":         &cfg.batteryLow,
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), "=", 2)
		if len(parts) != 2 {
			continue
		}
		if target, ok := values[parts[0]]; ok {
			if value, err := strconv.Atoi(parts[1]); err == nil {
				*target = value
			}
		}
		if parts[0] == "INTERVAL" {
			if seconds, err := strconv.ParseFloat(parts[1], 64); err == nil && seconds >= 0.20 {
				cfg.interval = time.Duration(seconds * float64(time.Second))
			}
		}
	}
	cfg.maxBrightness = clamp(cfg.maxBrightness, 1, 255)
	return cfg
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func readInt(path string, fallback int) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fallback
	}
	return value
}

func readText(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}

func appendLog(message string) {
	if info, err := os.Stat(logPath); err == nil && info.Size() > 16384 {
		data, _ := os.ReadFile(logPath)
		if len(data) > 8192 {
			data = data[len(data)-8192:]
		}
		_ = os.WriteFile(logPath, data, 0644)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s %s\n", time.Now().In(localZone).Format("2006-01-02 15:04:05"), message)
}

func waitForBoot() {
	for {
		output, _ := exec.Command("getprop", "sys.boot_completed").Output()
		if strings.TrimSpace(string(output)) == "1" {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func openLED(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY, 0)
}

func writeFileValue(file *os.File, value int) error {
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	_, err := fmt.Fprintf(file, "%d\n", value)
	return err
}

func (c *controller) writeLED(r, g, b int) {
	r = r * c.cfg.maxBrightness / 255
	g = g * c.cfg.maxBrightness / 255
	b = b * c.cfg.maxBrightness / 255
	c.writeRawLED(r, g, b)
}

func (c *controller) writeRawLED(r, g, b int) {
	if r == c.lastR && g == c.lastG && b == c.lastB {
		return
	}
	if writeFileValue(c.red, r) != nil || writeFileValue(c.green, g) != nil || writeFileValue(c.blue, b) != nil {
		if time.Since(c.lastErrorLog) >= time.Minute {
			appendLog("error=led_write_failed")
			c.lastErrorLog = time.Now()
		}
	}
	c.lastR, c.lastG, c.lastB = r, g, b
}

func (c *controller) off() {
	c.writeLED(0, 0, 0)
}

// releaseLED lets Android own the next battery-light update.
func (c *controller) releaseLED() {
	c.lastR, c.lastG, c.lastB = -1, -1, -1
}

func detectMode() int {
	data, err := os.ReadFile(prefPath)
	if err == nil {
		index := bytes.Index(data, modeMarker)
		if index >= 0 {
			valueAt := index + len(modeMarker)
			if valueAt < len(data) && data[valueAt] >= '0' && data[valueAt] <= '3' {
				return int(data[valueAt] - '0')
			}
		}
	}
	output, _ := exec.Command("getprop", "persist.cpuschedulingmode.value").Output()
	switch strings.TrimSpace(string(output)) {
	case "powersave":
		return 0
	case "performance":
		return 2
	default:
		return 1
	}
}

func screenIsOn() bool {
	if brightness := readInt(backlightPath, -1); brightness >= 0 {
		return brightness > 0
	}
	output, err := exec.Command("getprop", "debug.tracing.screen_state").Output()
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(output)) == "2"
}

func readTemperature(path string) int {
	value := readInt(path, 0)
	if value > 1000 {
		value /= 1000
	}
	return value
}

func (c *controller) sampleTemperatures() {
	c.cpuTemperature = readTemperature(cpuTempPath)
	c.batteryTemperature = readTemperature(batteryTempPath)
}

func (c *controller) sampleBattery() {
	c.capacity = readInt("/sys/class/power_supply/battery/capacity", 50)
	c.batteryStatus = readText("/sys/class/power_supply/battery/status", "Unknown")
}

func (c *controller) systemOwnsBatteryLED() bool {
	return c.batteryStatus == "Charging" || c.batteryStatus == "Full" || c.capacity <= c.cfg.batteryLow
}

func (c *controller) handoffBatteryLED() {
	if c.lastState != "SYSTEM_BATTERY_LED" {
		if c.batteryStatus == "Charging" || c.batteryStatus == "Full" {
			c.writeRawLED(40, 40, 40)
		} else {
			c.writeRawLED(40, 5, 0)
		}
	}
	c.setState("SYSTEM_BATTERY_LED")
	c.releaseLED()
}

func (c *controller) setState(state string) {
	if state == c.lastState {
		return
	}
	c.lastState = state
	_ = os.WriteFile(statePath, []byte(state+"\n"), 0644)
	appendLog(fmt.Sprintf("state=%s mode=%d battery=%d status=%s cpu_temp=%dC battery_temp=%dC", state, c.mode, c.capacity, c.batteryStatus, c.cpuTemperature, c.batteryTemperature))
}

func (c *controller) rainbowStep() {
	segment := c.phase / 10
	offset := c.phase % 10
	rising := offset * 255 / 10
	falling := 255 - rising
	var r, g, b int
	switch segment {
	case 0:
		r, g = 255, rising
	case 1:
		r, g = falling, 255
	case 2:
		g, b = 255, rising
	case 3:
		g, b = falling, 255
	case 4:
		r, b = rising, 255
	default:
		r, b = 255, falling
	}
	// Rainbow uses a softer peak than warnings and Full Power mode.
	c.writeLED(r*200/255, g*200/255, b*200/255)
	c.phase = (c.phase + 1) % 60
}

func (c *controller) breatheStep(r, g, b, speed int) {
	wave := c.breathe
	if wave > 20 {
		wave = 40 - wave
	}
	scale := 51 + wave*204/20
	c.writeLED(r*scale/255, g*scale/255, b*scale/255)
	c.breathe = (c.breathe + speed) % 40
}

func (c *controller) updateEffect() {
	if !c.screenOn {
		if c.systemOwnsBatteryLED() {
			c.handoffBatteryLED()
			return
		}
		c.setState("SCREEN_OFF")
		c.off()
		return
	}
	if c.cpuTemperature >= c.cfg.cpuTempDanger || c.batteryTemperature >= c.cfg.batteryTempDanger {
		c.setState("TEMP_DANGER_RED_FAST_BLINK")
		if c.tick%4 < 2 {
			c.writeLED(255, 0, 0)
		} else {
			c.off()
		}
	} else if c.cpuTemperature >= c.cfg.cpuTempWarn || c.batteryTemperature >= c.cfg.batteryTempWarn {
		c.setState("TEMP_WARN_ORANGE_BREATHE")
		c.breatheStep(230, 65, 0, 3)
	} else if c.systemOwnsBatteryLED() {
		c.handoffBatteryLED()
	} else {
		switch c.mode {
		case 0:
			c.setState("POWER_SAVE_GREEN_BREATHE")
			c.breatheStep(0, 165, 55, 1)
		case 1:
			c.setState("BALANCED_RAINBOW")
			c.rainbowStep()
		case 2:
			c.setState("GAME_ELECTRIC_BLUE_BREATHE")
			c.breatheStep(20, 90, 230, 1)
		case 3:
			c.setState("FULL_POWER_MAGENTA_PULSE")
			c.breatheStep(255, 0, 90, 3)
		}
	}
}

func main() {
	waitForBoot()
	time.Sleep(3 * time.Second)

	red, errR := openLED(redPath)
	green, errG := openLED(greenPath)
	blue, errB := openLED(bluePath)
	if errR != nil || errG != nil || errB != nil {
		appendLog("error=missing_led_node")
		_ = os.WriteFile(statePath, []byte("ERROR_MISSING_LED_NODE\n"), 0644)
		return
	}
	defer red.Close()
	defer green.Close()
	defer blue.Close()

	for _, color := range []string{"red", "green", "blue"} {
		_ = os.WriteFile("/sys/class/leds/"+color+"/trigger", []byte("none\n"), 0644)
	}

	c := &controller{
		cfg: readConfig(), red: red, green: green, blue: blue,
		lastR: -1, lastG: -1, lastB: -1,
		mode: 1, capacity: 50, batteryStatus: "Unknown", screenOn: true,
	}
	c.screenOn = screenIsOn()
	c.mode = detectMode()
	c.sampleBattery()
	c.sampleTemperatures()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		c.off()
		_ = os.WriteFile(statePath, []byte("stopped\n"), 0644)
		os.Exit(0)
	}()

	for {
		if c.tick%10 == 0 {
			c.screenOn = screenIsOn()
		}
		if !c.screenOn {
			c.updateEffect()
			// Android freezes this process during suspend; no wake lock is held.
			time.Sleep(5 * time.Second)
			c.sampleBattery()
			c.screenOn = screenIsOn()
			if c.screenOn {
				c.mode = detectMode()
				c.sampleTemperatures()
			}
			continue
		}
		if c.tick%10 == 0 {
			c.mode = detectMode()
		}
		// Fast battery handoff keeps Android's charging and warning lights authoritative.
		if c.tick%25 == 0 {
			c.sampleBattery()
		}
		if c.tick%50 == 0 {
			c.sampleTemperatures()
		}
		c.updateEffect()
		c.tick++
		time.Sleep(c.cfg.interval)
		if c.tick > 1000000 {
			c.tick = 0
		}
	}
}
