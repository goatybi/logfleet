package ssh

import (
	"fmt"
	"sync"
	"time"

	"logfleet/internal/config"
)

type ServerStatus struct {
	Server  config.Server
	Online  bool
	Latency time.Duration
	Error   error
}

func CheckAll(servers []config.Server, timeout time.Duration) []ServerStatus {
	results := make([]ServerStatus, len(servers))
	var wg sync.WaitGroup

	for i, srv := range servers {
		wg.Add(1)
		go func(idx int, s config.Server) {
			defer wg.Done()
			results[idx] = checkOne(s, timeout)
		}(i, srv)
	}

	wg.Wait()
	return results
}

func checkOne(srv config.Server, timeout time.Duration) ServerStatus {
	start := time.Now()

	type result struct {
		err error
	}
	ch := make(chan result, 1)

	go func() {
		client, err := connect(srv)
		if err != nil {
			ch <- result{err}
			return
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			ch <- result{fmt.Errorf("session: %w", err)}
			return
		}
		defer session.Close()

		if err := session.Run("true"); err != nil {
			ch <- result{fmt.Errorf("command: %w", err)}
			return
		}

		ch <- result{nil}
	}()

	select {
	case r := <-ch:
		latency := time.Since(start)
		if r.err != nil {
			return ServerStatus{Server: srv, Online: false, Error: r.err}
		}
		return ServerStatus{Server: srv, Online: true, Latency: latency}

	case <-time.After(timeout):
		return ServerStatus{
			Server: srv,
			Online: false,
			Error:  fmt.Errorf("timeout after %s", timeout),
		}
	}
}