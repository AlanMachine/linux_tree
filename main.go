package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type counts struct {
	dirs, files int
}

type flags struct {
	f, h, d bool
}

type sortedFiles struct {
	list []os.FileInfo
}

func (s sortedFiles) Len() int {
	return len(s.list)
}

func (s sortedFiles) Less(i, j int) bool {
	return strings.ToLower(s.list[i].Name()) < strings.ToLower(s.list[j].Name())
}

func (s sortedFiles) Swap(i, j int) {
	s.list[i], s.list[j] = s.list[j], s.list[i]
}

func main() {
	f := flag.Bool("f", false, "Print the full path prefix for each file.")
	h := flag.Bool("h", false, "Print the size in a more human readable way.")
	d := flag.Bool("d", false, "List directories only.")

	flag.Parse()

	dirName := "."
	if len(flag.Args()) > 0 {
		dirName = flag.Args()[0]
	}

	r, err := os.Open(dirName)
	if err != nil {
		printError(dirName, "error opening dir")
		os.Exit(1)
	}
	defer func() {
		err = r.Close()
		if err != nil {
			panic(err)
		}
	}()

	flags := flags{*f, *h, *d}
	counts := counts{0, 0}

	_, _ = color.New(color.FgHiBlue, color.Bold).Println(dirName)

	dirTree(dirName, "", flags, &counts)

	if flags.d {
		fmt.Printf("\n%d directories\n", counts.dirs)
	} else {
		fmt.Printf("\n%d directories, %d files\n", counts.dirs, counts.files)
	}
}

// Directory tree traversal
func dirTree(dirName string, sep string, flags flags, counts *counts) {
	r, err := os.Open(dirName)
	if err != nil {
		return
	}

	d, err := r.Readdir(0)
	if err != nil {
		return
	}

	var sorted sortedFiles

	sorted.list = append(sorted.list, d...)
	sort.Sort(sorted)

	sorted.list = dirTreeFilter(sorted.list, flags)

	count := len(sorted.list)
	for k, v := range sorted.list {
		sepNext := "├── "
		if count == 1 || k == count-1 {
			sepNext = "└── "
		}

		fullPath := dirName + string(filepath.Separator) + v.Name()

		dirTreePrint(v, sep+sepNext, fullPath, flags)

		if v.IsDir() {
			counts.dirs++
			sepNext := sep + "│   "
			if k == count-1 {
				sepNext = sep + "    "
			}
			dirTree(fullPath, sepNext, flags, counts)
		} else {
			counts.files++
		}
	}

	defer func() {
		err = r.Close()
		if err != nil {
			panic(err)
		}
	}()
}

// Formatting and printing a tree node
func dirTreePrint(v os.FileInfo, sep, fullPath string, flags flags) {
	var colorSet *color.Color

	if v.IsDir() {
		colorSet = color.New(color.FgHiBlue, color.Bold)
	} else if strings.HasPrefix(v.Mode().String(), "L") {
		colorSet = color.New(color.FgHiCyan, color.Bold)
	} else if strings.Contains(v.Mode().String(), "x") {
		colorSet = color.New(color.FgHiGreen, color.Bold)
	} else {
		colorSet = color.New()
	}

	name := v.Name()

	if flags.f {
		name = fullPath
	}

	fmt.Printf("%s", sep)

	if flags.h {
		fmt.Printf("[%4s]  ", formatSize(int(v.Size())))
	}

	_, _ = colorSet.Println(name)
}

// Directory list filtering
func dirTreeFilter(list []os.FileInfo, flags flags) []os.FileInfo {
	if !flags.d {
		return list
	}

	var tmpList []os.FileInfo

	for _, v := range list {
		if v.IsDir() {
			tmpList = append(tmpList, v)
		}
	}

	return tmpList
}

// Formatting file and directory sizes
func formatSize(size int) string {
	kilo := 1024
	mega := kilo * 1024
	giga := mega * 1024
	tera := giga * 1024
	peta := tera * 1024

	if size <= kilo {
		return fmt.Sprintf("%d", size)
	} else if size > kilo && size <= mega {
		return fmt.Sprintf("%.1fK", float64(size)/float64(kilo))
	} else if size > mega && size <= giga {
		return fmt.Sprintf("%.1fM", float64(size)/float64(mega))
	} else if size > giga && size <= tera {
		return fmt.Sprintf("%.1fG", float64(size)/float64(giga))
	} else if size > tera && size <= peta {
		return fmt.Sprintf("%.1fT", float64(size)/float64(tera))
	} else {
		return fmt.Sprintf(">1P")
	}
}

// Formatting and printing error messages
func printError(dirName, message string) {
	_, _ = color.New(color.FgHiBlue, color.Bold).Print(dirName)
	fmt.Printf(" [%s]\n\n", message)
	fmt.Println("0 directories, 0 files")
}
