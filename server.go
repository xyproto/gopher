package gopher

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

// Config is the Gopher server configuration struct
type Config struct {
	Host string
	Port int
	Root string
	addr string
}

const (
	GopherError byte = '3'
	GopherInfo  byte = 'i'
	GopherMenu  byte = '1'
	GopherText  byte = '0'
)

type Item struct {
	Type     byte
	Name     string
	Selector string
	Host     string
	Port     int
}

type List []Item

func New(host string, port int, path string) (*Config, error) {
	conf := &Config{host, port, path, ""}

	// check and correct root directory
	if !filepath.IsAbs(conf.Root) {
		path, err := filepath.Abs(conf.Root)
		if err != nil {
			return nil, err
		}
		conf.Root = path
	}
	if conf.Root[len(conf.Root)-1:] != "/" {
		conf.Root = conf.Root + "/"
	}
	conf.addr = net.JoinHostPort("0.0.0.0", strconv.Itoa(conf.Port))
	return conf, nil
}

func (i Item) String() string {
	switch i.Type {
	case 'i':
		return fmt.Sprintf("%c%s\t\tinfo.host\t1\r\n",
			i.Type, i.Name)
	default:
		return fmt.Sprintf("%c%s\t%s\t%s\t%d\r\n",
			i.Type, i.Name, i.Selector, i.Host, i.Port)
	}
}

func (l List) String() string {
	var b bytes.Buffer
	for _, i := range l {
		fmt.Fprint(&b, i)
	}
	fmt.Fprint(&b, ".\r\n")
	return b.String()
}

// Row returns a gopher item ready to be served
func (conf *Config) Row(t byte, n, s, h string, p int) Item {
	switch t {
	case GopherError:
		s = ""
		h = "error.host"
		p = 1
	case GopherInfo:
		s = ""
		h = "info.host"
		p = 1
	case GopherMenu:
		if h == "" {
			h = conf.Host
			p = conf.Port
		}
	case GopherText:
		if h == "" {
			h = conf.Host
			p = conf.Port
		}
	default:
		return conf.Row(GopherError, "Internal server error", "", "", 0)
	}
	return Item{t, n, s, h, p}
}

// Exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// ListDir scans the given 'path' and returns a gopher list of entries
func (conf *Config) ListDir(path string) List {
	var l List
	count := 0
	filepath.Walk(path, (func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// due to how Walk works, the first folder here is the given path
			// itself, and we need to not SkipDir it. So for the first run of
			// this recursive Walk function, we will allow it to nest deeper
			// by just returning nil
			if count > 0 {
				l = append(l, conf.Row(GopherMenu, info.Name(), p[len(conf.Root)-1:], "", 0))
				count++
				return filepath.SkipDir
			}
			count++
			return nil
		}
		if info.Name() != "gophermap" {
			l = append(l, conf.Row(GopherText, info.Name(), p[len(conf.Root)-1:], "", 0))
		}
		return nil
	}))
	return l
}

// ListenAndServe starts a gopher server at 'conf.addr'
func (conf *Config) ListenAndServe(logRequest func(string, string)) error {
	ln, err := net.Listen("tcp", conf.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// Ignore invalid request
			//customPrintf("%v\n", err)
			continue
		}
		go conf.handleConn(conn, logRequest)
	}
}

// handleConn manages open and close of network conn
func (conf *Config) handleConn(conn net.Conn, logRequest func(string, string)) {
	defer conn.Close()

	buf := bufio.NewReader(conn)
	req, _, err := buf.ReadLine()
	if err != nil {
		fmt.Fprint(conn, conf.Error("Invalid request."))
	}

	logRequest(fmt.Sprintf("%v", conn.RemoteAddr()), filepath.Clean("/"+string(req)))
	conf.handleRequest(string(req), conn)
}

// handleRequest parses the request and sends an answer
func (conf *Config) handleRequest(req string, conn net.Conn) {
	safeReq := filepath.Clean("/" + req)
	req = conf.Root + safeReq

	f, err := os.Open(req)
	defer f.Close()
	if err != nil {
		fmt.Fprint(conn, conf.Error("Resource not found."))
		return
	}

	fi, _ := f.Stat()
	if fi.IsDir() {
		var l List
		if ok, err := Exists(req + "/gophermap"); ok == true && err == nil {
			l = append(l, conf.Gophermap(req+"/gophermap")...)
		} else {
			l = append(l, conf.ListDir(req)...)
		}
		fmt.Fprint(conn, l)
		return
	}

	io.Copy(conn, f)
	return
}

// responseError returns a full response with a gopher-formatted error 's'
func (conf *Config) Error(s string) List {
	return List{conf.Row(GopherError, s, "", "", 0)}
}
