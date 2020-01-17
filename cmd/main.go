package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/sirupsen/logrus"
)

type argList struct {
	args map[string]string
}

func (a *argList) String() string {
	return fmt.Sprintf("%v", a.args)
}

func (a *argList) Set(value string) error {
	vals := strings.Split(value, "=")
	if len(vals) != 2 {
		return fmt.Errorf("value must be in form key=value")
	}
	a.args[vals[0]] = vals[1]
	return nil
}

type config struct {
	Filters map[string]string
	Exec    []string
}

func readConfig() config {
	fil := argList{}
	flag.Var(&fil, "filters", "-filter key=value -filter key=value")
	cmd := flag.String("cmd", "", "command to execute on detected event")
	flag.Parse()
	cfg := config{Filters: fil.args, Exec: strings.Fields(*cmd)}
	return cfg
}

func getEvents(f filters.Args) <-chan events.Message {
	msgs := make(chan events.Message)
	cli, err := client.NewEnvClient()
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		for {
			m, errC := cli.Events(context.Background(), types.EventsOptions{Filters: f})
		LOOP:
			for {
				select {
				case evt := <-m:
					msgs <- evt
				case e := <-errC:
					logrus.Error(e)
					time.Sleep(time.Second)
					break LOOP
				}
			}

		}
	}()
	return msgs
}

func precommand(fields []string) func(events.Message) {

	if len(fields) == 0 {
		return func(e events.Message) {
			logrus.Info(e)
		}
	}
	return func(e events.Message) {
		var a []string
		if len(fields) > 1 {
			a = fields[1:]
		}
		cmd := exec.Command(fields[0], a...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			logrus.Error(err)
		}
	}
}

func main() {
	cfg := readConfig()

	opts := filters.NewArgs()
	for k, v := range cfg.Filters {
		opts.Add(k, v)
	}

	execute := precommand(cfg.Exec)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case m := <-getEvents(opts):
			execute(m)
		case sg := <-sig:
			logrus.Infof("Terminated with %v", sg)
			return
		}
	}
}
