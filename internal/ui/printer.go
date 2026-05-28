package ui

import (
	"fmt"
	"strings"

	"logfleet/internal/ssh"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

var serverColors = []string{cyan, green, yellow, purple, blue}
var serverColorMap = map[string]string{}
var colorIndex = 0

func serverColor(name string) string {
	if c, ok := serverColorMap[name]; ok {
		return c
	}
	c := serverColors[colorIndex%len(serverColors)]
	serverColorMap[name] = c
	colorIndex++
	return c
}

func PrintLine(line ssh.LogLine) {
	if line.Err != nil {
		fmt.Printf("%s%s[%s]%s %s! %s%s\n",
			bold, red, line.Server, reset,
			yellow, line.Err.Error(), reset)
		return
	}

	color := serverColor(line.Server)
	level := DetectLevel(line.Line)
	levelStr := formatLevel(level)

	fmt.Printf("%s%s[%-12s]%s %s[%-20s]%s %s%s%s\n",
		bold, color, line.Server, reset,
		gray, line.LogFile, reset,
		levelStr,
		line.Line,
		reset)
}

func PrintError(msg string) {
	fmt.Printf("%s%s[logfleet] %s%s\n", bold, red, msg, reset)
}

func PrintInfo(msg string) {
	fmt.Printf("%s%s[logfleet] %s%s\n", bold, gray, msg, reset)
}

func PrintConnected(serverName string) {
	color := serverColor(serverName)
	fmt.Printf("%s%s[%-12s]%s %sconnected%s\n",
		bold, color, serverName, reset,
		green, reset)
}

type LogLevel int

const (
	LevelNormal LogLevel = iota
	LevelWarn
	LevelError
)

func ParseLevel(s string) (LogLevel, error) {
	switch strings.ToLower(s) {
	case "error", "err":
		return LevelError, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "normal", "all", "":
		return LevelNormal, nil
	default:
		return LevelNormal, fmt.Errorf("unknown level '%s' — use: error, warn, all", s)
	}
}

func DetectLevel(line string) LogLevel {
	lower := strings.ToLower(line)

	errorWords := []string{"error", "err ", "fatal", "panic", "critical", "failed", "failure", "exception", "emerg", "alert", "crit"}
	for _, w := range errorWords {
		if strings.Contains(lower, w) {
			return LevelError
		}
	}

	warnWords := []string{"warn", "warning", "deprecated", "timeout", "retry", "slow"}
	for _, w := range warnWords {
		if strings.Contains(lower, w) {
			return LevelWarn
		}
	}

	return LevelNormal
}

func formatLevel(level LogLevel) string {
	switch level {
	case LevelError:
		return red
	case LevelWarn:
		return yellow
	default:
		return ""
	}
}