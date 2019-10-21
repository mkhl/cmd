/*
Acmepipe executes a command
and replaces the current Acme window's body with the output.
It writes the window's body to the command's stdin
and writes messages from its stderr to the window's errors file.

Usage:

	acmepipe [<option>...] [<command> [<argument>...]]

The options are:

	-all
		replace the body all at once
	-mark
		mark each changed region as a separate undo step

Without a command, acmepipe uses its stdin as the output instead.

Acmepipe draws inspiration from
github.com/9fans/acme/acmego,
github.com/eaburns/Fmt, and
github.com/rog-go/cmd/apipe.
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"9fans.net/go/acme"
)

var (
	this = os.Args[0]
	win  *acme.Win
	all  = flag.Bool("all", false, "replace the body all at once")
	mark = flag.Bool("mark", false, "mark each changed region as a separate undo step")
)

type thing struct {
	oldStart, oldEnd int
	op               string
	newStart, newEnd int
}

func usage() {
	usage := `usage: %s [<option>...] [<command> [<argument>...]]
Execute <command>, replace acme/$winid/body with its output,
and write diagnostic messages to acme/$winid/errors.
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
	if err := cli(flag.Args()); err != nil {
		switch e := err.(type) {
		case *exec.ExitError: // command failed
			status := e.Sys().(syscall.WaitStatus).ExitStatus()
			os.Exit(status)
		default:
			log.Fatalln(err)
		}
	}
}

func cli(args []string) error {
	if err := open(); err != nil {
		return err
	}
	defer win.CloseFiles()
	body, err := win.ReadAll("body")
	if err != nil {
		return err
	}
	stdout, stderr, exit := run(args)
	if err := writeErrors(stderr); err != nil {
		return err
	}
	if exit != nil {
		return exit
	}
	if *all {
		return writeOutput(body, stdout)
	}
	return patchOutput(body, stdout)
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

func run(args []string) ([]byte, []byte, error) {
	if len(args) == 0 {
		stdin, err := ioutil.ReadAll(os.Stdin)
		return stdin, nil, err
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func writeErrors(buf []byte) error {
	if len(buf) == 0 {
		return nil // no errors
	}
	_, err := win.Write("errors", buf)
	return err
}

func writeOutput(old, new []byte) error {
	if len(new) == 0 {
		return nil // no output
	}
	if bytes.Equal(new, old) {
		return nil // no change
	}
	_, _, err := win.ReadAddr() // make sure the address file is open
	if err != nil {
		return err
	}
	if err := win.Ctl("addr=dot"); err != nil {
		return err
	}
	q0, q1, err := win.ReadAddr()
	if err != nil {
		return err
	}
	if err := win.Addr(","); err != nil {
		return err
	}
	if _, err := win.Write("data", new); err != nil {
		return err
	}
	if err := win.Addr("#%d,#%d", q0, q1); err != nil {
		return err
	}
	return win.Ctl("dot=addr\nshow")
}

func patchOutput(old, new []byte) error {
	if len(new) == 0 {
		return nil // no output
	}
	newFile, err := tempfile(new)
	if err != nil {
		return err
	}
	oldFile, err := tempfile(old)
	if err != nil {
		return err
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("9", "diff", oldFile.Name(), newFile.Name())
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil && err.(*exec.ExitError) == nil { // ignore exit status
		return err
	}
	if err := writeErrors(stderr.Bytes()); err != nil {
		return err
	}
	return apply(stdout.Bytes(), new)
}

func tempfile(buf []byte) (*os.File, error) {
	file, err := ioutil.TempFile("", this)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(buf)
	return file, err
}

func apply(diff, buf []byte) error {
	if len(diff) == 0 {
		return nil // no diff
	}
	if !*mark {
		if err := win.Ctl("mark\nnomark"); err != nil {
			return err
		}
	}
	lines := strings.Split(string(diff), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if err := applyRegion(line, buf); err != nil {
			return err
		}
	}
	return nil
}

func applyRegion(line string, buf []byte) error {
	if len(line) == 0 {
		return nil // no change
	}
	if line[0] == '<' || line[0] == '-' || line[0] == '>' {
		return nil // details
	}
	i := strings.IndexAny(line, "acd")
	if i < 0 {
		log.Println("cannot parse diff:", line)
		return nil
	}
	n1, n2 := span(line[:i])
	n3, n4 := span(line[i+1:])
	switch line[i] {
	case 'a': // add
		if err := write(region(buf, n3, n4), "%d+#0", n1); err != nil {
			return err
		}
	case 'c': // change
		if err := write(region(buf, n3, n4), "%d,%d", n1, n2); err != nil {
			return err
		}
	case 'd': // delete
		if err := write(nil, "%d,%d", n1, n2); err != nil {
			return err
		}
	}
	return nil
}

func span(text string) (int, int) {
	i := strings.IndexByte(text, ',')
	if i < 0 {
		n, err := strconv.Atoi(text)
		if err != nil {
			return 0, 0
		}
		return n, n
	}
	x, err1 := strconv.Atoi(text[:i])
	y, err2 := strconv.Atoi(text[i+1:])
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return x, y
}

func region(text []byte, start, end int) []byte {
	start-- // line numbers are 1-based
	i := 0
	for ; i < len(text) && start > 0; i++ {
		if text[i] == '\n' {
			start--
			end--
		}
	}
	startByte := i
	for ; i < len(text) && end > 0; i++ {
		if text[i] == '\n' {
			end--
		}
	}
	endByte := i
	return text[startByte:endByte]
}

func write(buf []byte, format string, args ...interface{}) error {
	if err := win.Addr(format, args...); err != nil {
		return err
	}
	_, err := win.Write("data", buf)
	return err
}
