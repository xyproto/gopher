package gopher

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	filename string = "gophermap"
)

// Gophermap parses the given file 'fn' and returns a proper gopher list of items
func (conf *Config) Gophermap(fn string) List {
	var l List

	f, err := os.Open(fn)
	defer f.Close()
	if err != nil {
		return conf.Error("gophermap error")
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		row := scanner.Text()
		if strings.Contains(row, "\t") {
			itemtype, cols := parse(row)
			p, _ := strconv.Atoi(cols[3]) // port
			switch itemtype {
			case GopherMenu:
				fallthrough
			case GopherText:
				i := conf.Row(itemtype, cols[0], cols[1], cols[2], p)
				l = append(l, i)
			case '!':
				if cols[0] == "!" && cols[1] == "list" {
					l = append(l, conf.ListDir(filepath.Dir(fn))...)
				}
			default:
				i := conf.Row(GopherError, strings.Replace(row, "\t", "\\t", -1), "", "", 0)
				l = append(l, i)
			}
		} else {
			l = append(l, conf.Row(GopherInfo, row, "", "", 0))
		}
	}

	return l
}

func parse(s string) (byte, []string) {
	cols := strings.Split(s, "\t")
	itemtype := []byte(cols[0][:1])[0]
	fName := cols[0][1:len(cols[0])]
	fSelector := ""
	fHost := ""
	fPort := ""

	if len(cols) >= 4 {
		fSelector = cols[1]
		fHost = cols[2]
		fPort = cols[3]
	} else if len(cols) >= 3 {
		fSelector = cols[1]
		fHost = cols[2]
	} else if len(cols) >= 2 {
		fSelector = cols[1]
	}

	fields := []string{fName, fSelector, fHost, fPort}
	return itemtype, fields
}
