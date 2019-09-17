package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	statusCode = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "command_exit_code",
		Help: "Exit code of command.",
	})
	commandDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "command_duration_seconds",
		Help: "Time spent running command.",
	})
)

func init() {
	prometheus.MustRegister(statusCode)
	prometheus.MustRegister(commandDuration)
}

var (
	period     = flag.Duration("period", 10*time.Second, "Period with which to run the command.")
	timeout    = flag.Duration("timeout", 10*time.Minute, "Amount of time to give the command to run.")
	listenAddr = flag.String("listen-addr", ":9152", "Address to listen on")
	cwd        = flag.String("working-dir", "", "Working directory for command")

	outputLock      sync.Mutex
	outputBuf       []byte
	lastRunStart    time.Time
	lastRunDuration time.Duration

	tmpl = template.Must(template.New("index").Parse(`<html>
	<head><title>Prometheus Command Runner</title></head>
	<body>
		<h2>Prometheus Command Runner</h2>
		<p>"{{.Command}}" output:</p>
		<pre>{{.Output}}</pre>
		<p>Run at {{.Time}} took {{.Duration}}<p>
	</body>
	</html>`))
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s (options) command...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) <= 0 {
		flag.Usage()
		os.Exit(2)
	}
	command := flag.Args()[0]
	args := flag.Args()[1:]

	go func() {
		// Run once immediately at startup
		run(command, args)

		// Then start the delay loop
		for range time.Tick(*period) {
			run(command, args)
		}
	}()

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		outputLock.Lock()
		defer outputLock.Unlock()

		tmpl.Execute(w, struct {
			Command, Output string
			Time            time.Time
			Duration        time.Duration
		}{
			Command:  command + " " + strings.Join(args, " "),
			Output:   string(outputBuf),
			Time:     lastRunStart,
			Duration: lastRunDuration,
		})
	}))
	http.Handle("/metrics", prometheus.Handler())

	log.Printf("Listening on address %s", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Printf("ListenAndServe failed: %v", err)
	}
}

func run(command string, args []string) {
	start := time.Now()
	log.Infof("Running '%s' with argments %v", command, args)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = *cwd
	out, err := cmd.CombinedOutput()

	reap()

	duration := time.Now().Sub(start)
	commandDuration.Observe(duration.Seconds())

	outputLock.Lock()
	outputBuf = out
	lastRunStart = start
	lastRunDuration = duration
	outputLock.Unlock()

	if err == nil {
		log.Printf("Command exited successfully")
		statusCode.Set(0)
	} else {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				code := status.ExitStatus()
				log.Infof("Command exited with code: %d", code)
				log.Printf("Output:\n%s", out)
				statusCode.Set(float64(code))
				return
			}
		}

		log.Warnf("Error running command: %v", err)
		log.Printf("Output:\n%s", out)
		statusCode.Set(255)
	}
}

func reap() {
	log.Debugf("Reaping children")
	var wstatus syscall.WaitStatus
	for {
		pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
		if err == syscall.ECHILD {
			// No more children to reap, stop
			return
		}
		log.Debugf("Reaped child %d, wstatus=%+v, err=%v", pid, wstatus, err)
	}
}
