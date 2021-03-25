package main

import (
	"encoding/json"
	"github.com/creack/pty"
	"github.com/minoritea/libvterm-go"
	"io"
	"log"
	"os"
	"os/exec"
)

type Damage struct {
	Row  int
	Col  int
	Text string
}

func createDamageStreamingCallback(vts *libvterm.Screen, out io.Writer) func(rect libvterm.Rect) error {
	enc := json.NewEncoder(out)
	return func(rect libvterm.Rect) error {
		for i := rect.StartRow; i < rect.EndRow; i++ {
			for j := rect.StartCol; j < rect.EndCol; j++ {
				cell, err := vts.FetchCell(i, j)
				if err != nil {
					return err
				}

				err = enc.Encode(Damage{Row: i, Col: j, Text: cell.Text})
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func run() error {
	vt := libvterm.New(10, 80)
	defer vt.Close()

	fd, err := pty.Start(exec.Command("bash"))
	if err != nil {
		return err
	}
	defer fd.Close()

	vts := vt.ObtainScreen()
	vts.SetDamageCallback(createDamageStreamingCallback(vts, os.Stdout))

	_, err = io.Copy(vt, fd)
	return err
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
