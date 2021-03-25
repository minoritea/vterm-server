package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/gdamore/tcell"
	"golang.org/x/xerrors"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var cmdstr *string

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	notify := make(chan os.Signal)
	signal.Notify(notify, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-notify; cancel() }()

	cmdstr = flag.String("command", "go run main.go", "specify the API command")
	flag.Parse()

	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if cmdstr == nil {
		return xerrors.New("command option is not initialized")
	}

	ts, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	defer ts.Fini()

	err = ts.Init()
	if err != nil {
		return err
	}

	ts.Clear()

	cmd := exec.Command("sh", "-c", *cmdstr)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go parser(func(d Damage) error {
		runes := []rune(d.Text)
		if len(runes) == 0 {
			runes = []rune{' '}
		}
		ts.SetContent(d.Col, d.Row, runes[0], runes[1:], tcell.StyleDefault)
		ts.Show()
		return nil
	}).ReadFrom(out)

	err = cmd.Start()
	if err != nil {
		return err
	}

	<-ctx.Done()
	cmd.Process.Signal(syscall.SIGINT)
	log.Println("stopping...")

	return nil
}

type Damage struct {
	Row  int
	Col  int
	Text string
}

type parser func(Damage) error

func (p parser) ReadFrom(r io.Reader) (int64, error) {
	var d Damage
	dec := json.NewDecoder(r)
	for dec.More() {
		err := dec.Decode(&d)
		if err != nil {
			return 0, err
		}

		err = p(d)
		if err != nil {
			return 0, err
		}
	}

	return 0, nil
}
