/*
Autoacme watches acme/log for events
and executes a command for each.

Usage:

	autoacme <command> [<argument>...]

Autoacme sets $winid to the event's window ID
and passes its operation and target as additional arguments.
Output and errors from the command are written to the window's errors file.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"9fans.net/go/acme"
)

var (
	this = os.Args[0]
)

func usage() {
	usage := `usage: %s <command> [<argument>...]
Watch acme/log for events and execute <command> for each,
setting $winid and passing the operation and target as arguments.
`
	fmt.Fprintf(os.Stderr, usage, this)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", this))
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}
	if err := cli(flag.Args()); err != nil {
		log.Fatalln(err)
	}
}

func cli(args []string) error {
	src, err := acme.Log()
	if err != nil {
		return err
	}
	for {
		event, err := src.Read()
		if err != nil {
			return err
		}
		if err := run(args[0], args[1:], event); err != nil {
			switch err.(type) {
			case *exec.ExitError: // command failed
				log.Println(err)
			default:
				return err
			}
		}
	}
}

func run(name string, args []string, event acme.LogEvent) error {
	cmd := exec.Command(name, append(args, event.Op, event.Name)...)
	cmd.Dir = dir(event)
	cmd.Env = env(event)
	buf, exit := cmd.CombinedOutput()
	switch event.Op {
	case "del":
		return exit // can't write, win is gone
	default:
		if err := writeErrors(event.ID, buf); err != nil {
			return err
		}
		return exit
	}
}

func writeErrors(winid int, buf []byte) error {
	if len(buf) == 0 {
		return nil // no errors
	}
	win, err := acme.Open(winid, nil)
	if err != nil {
		return err
	}
	defer win.CloseFiles()
	_, err = win.Write("errors", buf)
	return err
}

func dir(event acme.LogEvent) string {
	name := event.Name
	if !strings.HasSuffix(name, "/") {
		name = path.Dir(name)
	}
	if _, err := os.Stat(name); err != nil {
		return ""
	}
	return name
}

func env(event acme.LogEvent) []string {
	return append(os.Environ(), fmt.Sprintf("winid=%v", event.ID))
}
