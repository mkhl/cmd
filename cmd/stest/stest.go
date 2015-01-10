package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

type statTests struct {
	stat  func(name string) (info os.FileInfo, err error)
	tests []flagTest
}
type flagTest struct {
	flag *bool
	test testFunc
}
type testFunc func(info os.FileInfo) bool

var (
	L = flag.Bool("L", false, "Matches files that are symbolic links")
	S = flag.Bool("S", false, "Matches files that are sockets")
	d = flag.Bool("d", false, "Matches files that are directories")
	e = flag.Bool("e", false, "Matches files that exist (regardless of type)")
	f = flag.Bool("f", false, "Matches files that are regular files")
	g = flag.Bool("g", false, "Matches files whose set group ID flag is set")
	k = flag.Bool("k", false, "Matches files whose sticky bit is set")
	p = flag.Bool("p", false, "Matches files that are named pipes (FIFOs)")
	r = flag.Bool("r", false, "Matches files that are readable")
	s = flag.Bool("s", false, "Matches files that have a size greater than zero")
	u = flag.Bool("u", false, "Matches files whose set user ID flag is set")
	v = flag.Bool("v", false, "Invert the sense of matching, to select non-matching files")
	w = flag.Bool("w", false, "Matches files that are writable")
	x = flag.Bool("x", false, "Matches files that are executable")

	tests = []statTests{
		{os.Lstat, []flagTest{
			{L, mode(os.ModeSymlink)}},
		},
		{os.Stat, []flagTest{
			{e, yes},
			{s, size},
			{f, not(mode(os.ModeType))},
			{d, mode(os.ModeDir)},
			{p, mode(os.ModeNamedPipe)},
			{S, mode(os.ModeSocket)},
			{u, mode(os.ModeSetuid)},
			{g, mode(os.ModeSetgid)},
			{k, mode(os.ModeSticky)},
			{r, perm(0444)},
			{w, perm(0222)},
			{x, perm(0111)},
		}},
	}
)

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		path := in.Text()
		if *v != test(path) {
			fmt.Println(path)
		}
	}
	if err := in.Err(); err != nil {
		log.Fatal(err)
	}
}

func test(path string) bool {
	for _, st := range tests {
		info, err := st.stat(path)
		if err != nil {
			continue
		}
		for _, ft := range st.tests {
			if *ft.flag && ft.test(info) {
				return true
			}
		}
	}
	return false
}

func yes(info os.FileInfo) bool {
	return true
}

func size(info os.FileInfo) bool {
	return info.Size() > 0
}

func mode(mode os.FileMode) testFunc {
	return func(info os.FileInfo) bool {
		return info.Mode()&mode > 0
	}
}

func perm(perm os.FileMode) testFunc {
	return func(info os.FileInfo) bool {
		return info.Mode().Perm()&perm > 0
	}
}

func not(test testFunc) testFunc {
	return func(info os.FileInfo) bool {
		return !test(info)
	}
}
