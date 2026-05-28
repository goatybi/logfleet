package ssh

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"logfleet/internal/config"
)

type LogLine struct {
	Server  string
	LogFile string
	Line    string
	Time    time.Time
	Err     error
}

func TailServer(srv config.Server, lines chan<- LogLine, done <-chan struct{}) {
	for {
		err := tailOnce(srv, lines, done)

		select {
		case <-done:
			return
		default:
		}

		if err != nil {
			lines <- LogLine{
				Server: srv.Name,
				Err:    fmt.Errorf("connection lost: %w — reconnecting in 5 sec", err),
				Time:   time.Now(),
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func tailOnce(srv config.Server, lines chan<- LogLine, done <-chan struct{}) error {
	client, err := connect(srv)
	if err != nil {
		return err
	}
	defer client.Close()

	errCh := make(chan error, len(srv.Logs))

	for _, logPath := range srv.Logs {
		go func(path string) {
			errCh <- tailFile(client, srv.Name, path, lines, done)
		}(logPath)
	}

	return <-errCh
}

func tailFile(client *ssh.Client, serverName, logPath string, lines chan<- LogLine, done <-chan struct{}) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	if err := session.Start(fmt.Sprintf("tail -F -n 50 %s 2>/dev/null", logPath)); err != nil {
		return fmt.Errorf("start tail: %w", err)
	}

	shortName := filepath.Base(logPath)
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		select {
		case <-done:
			session.Signal(ssh.SIGTERM)
			return nil
		default:
		}

		lines <- LogLine{
			Server:  serverName,
			LogFile: shortName,
			Line:    scanner.Text(),
			Time:    time.Now(),
		}
	}

	return scanner.Err()
}

func connect(srv config.Server) (*ssh.Client, error) {
	key, err := loadPrivateKey(srv.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("key %s: %w", srv.KeyPath, err)
	}

	cfg := &ssh.ClientConfig{
		User:            srv.User,
		Auth:            []ssh.AuthMethod{key},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", srv.Host, srv.Port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("connection to %s: %w", addr, err)
	}

	return client, nil
}

func loadPrivateKey(path string) (ssh.AuthMethod, error) {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}

	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("cannot parse key (passphrase protected?): %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

func TestConnect(srv config.Server) error {
	client, err := connect(srv)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return err
	}

	out, err := session.Output("echo ok")
	session.Close()
	client.Close()

	if err != nil {
		return err
	}

	if strings.TrimSpace(string(out)) != "ok" {
		return fmt.Errorf("unexpected response from server")
	}

	_ = net.ResolveTCPAddr

	return nil
}