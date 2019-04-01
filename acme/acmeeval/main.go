/*
Acmeeval evaluates commands
from inside the current Acme window's tag.

Usage:

	acmeeval [<command>...]

Acmeeval adds each command to the tag,
executes the command as Acme would have with button 2,
and then restores the original tag.

In addition to commands passed as arguments,
acmeeval also executes each line from stdin as a command.

Use acmeeval to access the Acme interactive command language
from external client programs.
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode/utf8"

	"9fans.net/go/acme"
)

var (
	this = os.Args[0]
	win  *acme.Win
)

func usage() {
	usage := `usage: %s [<command>...]
Execute each <command> from inside acme/$winid/tag.
Handles each line read from stdin as a <command>.
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
	if err := run(flag.Args()); err != nil {
		log.Fatalln(err)
	}
}

func run(args []string) error {
	if err := open(); err != nil {
		return err
	}
	defer win.CloseFiles()
	if err := win.Ctl("nomenu"); err != nil {
		return err
	}
	defer win.Ctl("menu")
	tag, err := win.ReadAll("tag")
	if err != nil {
		return err
	}
	idx := bytes.IndexRune(tag, '|')
	rest := tag[idx+1:]
	defer restore(rest)
	for _, arg := range args {
		if err := eval(arg); err != nil {
			return err
		}
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if err := eval(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func open() error {
	winid := os.Getenv("winid")
	id, err := strconv.Atoi(winid)
	if err != nil {
		return err
	}
	win, err = acme.Open(id, nil)
	return err
}

func eval(cmd string) error {
	if err := win.Ctl("cleartag"); err != nil {
		return err
	}
	tag, err := win.ReadAll("tag")
	if err != nil {
		return err
	}
	offset := utf8.RuneCount(tag)
	cmdlen := utf8.RuneCountInString(cmd)
	if err := win.Fprintf("tag", "%s", cmd); err != nil {
		return err
	}
	evt := new(acme.Event)
	evt.C1 = 'M'
	evt.C2 = 'x'
	evt.Q0 = offset
	evt.Q1 = offset + cmdlen
	return win.WriteEvent(evt)
}

func restore(tag []byte) error {
	if err := win.Ctl("cleartag"); err != nil {
		return err
	}
	_, err := win.Write("tag", tag)
	return err
}
