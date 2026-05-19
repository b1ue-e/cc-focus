package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Event struct {
	Reason string `json:"stop_reason"`
}

func cmdStart() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if daemonRunning() {
		fmt.Println("daemon is already running")
		os.Exit(0)
	}

	os.MkdirAll(cacheDir(), 0755)
	sockPath := cfg.Hook.SocketPath

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find executable: %v\n", err)
		os.Exit(1)
	}

	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, nil},
	}
	proc, err := os.StartProcess(exe, []string{exe, "--daemon", sockPath}, attr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start daemon: %v\n", err)
		os.Exit(1)
	}

	os.WriteFile(pidPath(), []byte(strconv.Itoa(proc.Pid)), 0644)
	fmt.Printf("daemon started (pid: %d)\n", proc.Pid)
}

func cmdStop() {
	data, err := os.ReadFile(pidPath())
	if err != nil {
		fmt.Println("daemon is not running")
		os.Exit(0)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid pid file: %v\n", err)
		os.Exit(1)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "find process: %v\n", err)
		os.Exit(1)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "signal daemon: %v\n", err)
		os.Exit(1)
	}

	os.Remove(pidPath())
	os.Remove(socketPath())
	fmt.Println("daemon stopped")
}

func cmdStatus() {
	cfg, _ := loadConfig()
	if daemonRunning() {
		fmt.Println("daemon: running")
	} else {
		fmt.Println("daemon: stopped")
	}
	fmt.Printf("terminal: %s\n", cfg.Terminal.Name)
	fmt.Printf("socket: %s\n", cfg.Hook.SocketPath)
}

func daemonRunning() bool {
	data, err := os.ReadFile(pidPath())
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func runDaemon(sockPath string) {
	cfg, _ := loadConfig()
	logPath := filepath.Join(cacheDir(), "events.log")

	logFile, err := os.OpenFile(filepath.Join(cacheDir(), "daemon.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}

	os.Remove(sockPath)
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		listener.Close()
		os.Remove(pidPath())
		os.Exit(0)
	}()

	log.Printf("daemon listening on %s", sockPath)

	startMonitor(cfg)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("accept: %v", err)
			}
			continue
		}
		go handleConnection(conn, cfg, logPath)
	}
}

func handleConnection(conn net.Conn, cfg Config, logPath string) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			log.Printf("parse event: %v", err)
			continue
		}

		log.Printf("event=%s activating terminal: %s", event.Reason, cfg.Terminal.Name)
		if err := activateTerminal(cfg); err != nil {
			log.Printf("activate: %v", err)
		}

		logEvent(logPath, event)
	}
}

func logEvent(logPath string, event Event) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	entry, _ := json.Marshal(event)
	f.Write(append(entry, '\n'))
}
