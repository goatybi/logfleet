package ui

import (
	"fmt"
	"strings"
	"time"

	"logfleet/internal/ssh"
)

func PrintStatus(results []ssh.ServerStatus) {
	online := 0
	for _, r := range results {
		if r.Online {
			online++
		}
	}
	total := len(results)

	fmt.Printf("\n%s%sSTATUS%s  %d/%d server(s) online\n\n",
		bold, gray, reset, online, total)

	maxName := 6
	for _, r := range results {
		if len(r.Server.Name) > maxName {
			maxName = len(r.Server.Name)
		}
	}

	fmt.Printf("  %s%-*s  %-6s  %-10s  %s%s\n",
		gray, maxName, "SERVER", "STATUS", "LATENCY", "INFO", reset)
	fmt.Printf("  %s%s%s\n", gray, strings.Repeat("─", maxName+35), reset)

	for _, r := range results {
		color := serverColor(r.Server.Name)

		if r.Online {
			latency := formatLatency(r.Latency)
			lColor := latencyColor(r.Latency)

			fmt.Printf("  %s%s%-*s%s  %s●%s%-6s  %s%-10s%s  %s%s:%d%s\n",
				bold, color, maxName, r.Server.Name, reset,
				green, reset, "online",
				lColor, latency, reset,
				gray, r.Server.Host, r.Server.Port, reset,
			)
		} else {
			errStr := shortenError(r.Error)
			fmt.Printf("  %s%s%-*s%s  %s●%s%-6s  %-10s  %s%s%s\n",
				bold, color, maxName, r.Server.Name, reset,
				red, reset, "offline",
				"—",
				red, errStr, reset,
			)
		}
	}

	fmt.Println()
}

func formatLatency(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func latencyColor(d time.Duration) string {
	ms := d.Milliseconds()
	switch {
	case ms < 200:
		return green
	case ms < 1000:
		return yellow
	default:
		return red
	}
}

func shortenError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	for _, prefix := range []string{
		"connection to ",
		"ssh: handshake failed: ",
		"dial tcp ",
	} {
		if idx := strings.Index(msg, prefix); idx != -1 {
			msg = msg[idx+len(prefix):]
		}
	}

	if len(msg) > 45 {
		msg = msg[:42] + "..."
	}
	return msg
}