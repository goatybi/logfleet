package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"logfleet/internal/config"
	"logfleet/internal/ssh"
	"logfleet/internal/ui"
)

const version = "0.1.3"

func main() {
	var (
		configPath  = flag.String("config", config.DefaultPath(), "path to config file")
		showVersion = flag.Bool("version", false, "print version")
	)

	tailCmd    := flag.NewFlagSet("tail", flag.ExitOnError)
	tailServer := tailCmd.String("server", "", "show only this server")
	tailGrep   := tailCmd.String("grep", "", "filter by text")
	tailLevel  := tailCmd.String("level", "", "min level: error, warn, all")

	statusCmd     := flag.NewFlagSet("status", flag.ExitOnError)
	statusTimeout := statusCmd.Duration("timeout", 10*time.Second, "connection timeout")

	flag.Parse()

	if *showVersion {
		fmt.Printf("logfleet v%s\n", version)
		return
	}

	args := flag.Args()
	subcommand := "tail"
	if len(args) > 0 {
		subcommand = args[0]
		args = args[1:]
	}

	switch subcommand {
	case "tail", "":
		tailCmd.Parse(args)
		runTail(*configPath, *tailServer, *tailGrep, *tailLevel)
	case "status":
		statusCmd.Parse(args)
		runStatus(*configPath, *statusTimeout)
	case "init":
		runInit(*configPath)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcommand)
		printHelp()
		os.Exit(1)
	}
}

func runTail(configPath, filterServer, filterGrep, filterLevel string) {
	minLevel, err := ui.ParseLevel(filterLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg := loadConfig(configPath)

	servers := cfg.Servers
	if filterServer != "" {
		servers = filterServers(servers, filterServer)
		if len(servers) == 0 {
			fmt.Fprintf(os.Stderr, "server '%s' not found in config\n", filterServer)
			os.Exit(1)
		}
	}

	lines := make(chan ssh.LogLine, 1000)
	done := make(chan struct{})

	if minLevel == ui.LevelError {
		ui.PrintInfo("mode: errors only (error/fatal/panic/critical...)")
	} else if minLevel == ui.LevelWarn {
		ui.PrintInfo("mode: warnings and errors")
	}
	ui.PrintInfo(fmt.Sprintf("connecting to %d server(s)...", len(servers)))

	for _, srv := range servers {
		go ssh.TailServer(srv, lines, done)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nstopping...")
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		case line, ok := <-lines:
			if !ok {
				return
			}
			if line.Err != nil {
				ui.PrintLine(line)
				continue
			}
			if minLevel > ui.LevelNormal {
				if ui.DetectLevel(line.Line) < minLevel {
					continue
				}
			}
			if filterGrep != "" {
				if !strings.Contains(strings.ToLower(line.Line), strings.ToLower(filterGrep)) {
					continue
				}
			}
			ui.PrintLine(line)
		}
	}
}

func runStatus(configPath string, timeout time.Duration) {
	cfg := loadConfig(configPath)
	ui.PrintInfo(fmt.Sprintf("checking %d server(s)...", len(cfg.Servers)))
	results := ssh.CheckAll(cfg.Servers, timeout)
	ui.PrintStatus(results)
	for _, r := range results {
		if !r.Online {
			os.Exit(1)
		}
	}
}

func runInit(path string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("config already exists: %s\n", path)
		fmt.Println("delete it manually if you want to start over.")
		return
	}

	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, []byte(config.ExampleConfig()), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "cannot write config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ config created: %s\n", path)
	fmt.Println("edit it and add your servers.")
	fmt.Println("then run: logfleet")
}

func loadConfig(path string) *config.Config {
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: run 'logfleet init' to create config\n")
		os.Exit(1)
	}
	if len(cfg.Servers) == 0 {
		fmt.Fprintf(os.Stderr, "no servers in config. add at least one.\n")
		os.Exit(1)
	}
	return cfg
}

func filterServers(servers []config.Server, name string) []config.Server {
	var result []config.Server
	for _, s := range servers {
		if strings.EqualFold(s.Name, name) {
			result = append(result, s)
		}
	}
	return result
}

func printHelp() {
	fmt.Print(`logfleet — stream logs from all your servers in one terminal

usage:
  logfleet [command] [flags]

commands:
  (none)    stream logs from all servers
  tail      stream logs
  status    check that all servers are alive
  init      create example config

flags for tail:
  --level error   errors only (error, fatal, panic...)
  --level warn    warnings and errors
  --server NAME   one server only
  --grep TEXT     filter by text

flags for status:
  --timeout 10s   connection timeout (default 10s)

global flags:
  --config PATH   path to config (default ~/.logfleet/config.yaml)
  --version       print version

examples:
  logfleet
  logfleet tail --level error
  logfleet status
  logfleet status --timeout 5s
  logfleet tail --server prod-web --level warn
`)
}