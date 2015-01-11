/*
Walk walks a filesystem hierarchy and prints the name of every file it encounters.

Usage:

	walk [options] [paths]

The flags are:

	-a
		include files whose names begin with a dot
	-d depth
		descend at most depth directory levels
	-q
		print file names as quoted strings

Each argument is interpreted as a path to walk.
If no arguments are present, walks the current working directory.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	paths = []string{"."}
	all   = flag.Bool("a", false, "include files whose names begin with a dot")
	depth = flag.Uint("d", 0, "descend at most <depth> directory levels")
	quote = flag.Bool("q", false, "print file names as quoted strings")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("walk: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: walk [flags] [paths]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() > 0 {
		paths = flag.Args()
	}
	for _, path := range paths {
		if err := walk(filepath.Clean(path)); err != nil {
			log.Fatal(err)
		}
	}
}

func skip(info os.FileInfo) error {
	if info.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

func walk(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if !*all {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") {
				return skip(info)
			}
		}
		if *depth > 0 {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if uint(len(strings.Split(rel, "/"))) > *depth {
				return skip(info)
			}
		}
		if *quote {
			fmt.Printf("%q\n", path)
		} else {
			fmt.Println(path)
		}
		return nil
	})
}
