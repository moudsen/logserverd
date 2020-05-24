package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/takama/daemon"
)

const (
	name        = "logserverd"
	description = "logserverd: Read Apache logfiles via a browser"
	anAllowed   = "abcdefghijklmnopqrstuvwxyz0123456789-_."
)

var stdlog, errlog *log.Logger

type Service struct {
	daemon.Daemon
}

func checkLogname(s string) bool {
	for _, char := range s {
		if !strings.Contains(anAllowed, strings.ToLower(string(char))) {
			return false
		}
	}
	return true
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	fmt.Println("use /_log?site=<sitename>&type=<access|error>")
}

func handleAccessLog(w http.ResponseWriter, req *http.Request) {
	// Obtain the ip address from the requestor. As this routine likely sits behind a reverse proxy,
	// first test for an ip in the header. Only if not there, use the RemoteAddr method.

	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}


	sites, ok := req.URL.Query()["site"]
	site := req.Host

	if ok && len(sites[0]) > 0 {
		if checkLogname(sites[0]) {
			site = sites[0]
		}
	}

	rtypes, ok := req.URL.Query()["type"]
	rtype := "error"

	if ok && len(rtypes[0]) > 0 {
		if rtypes[0]=="access" { rtype = "access" }
	}

	refreshvalue, ok := req.URL.Query()["refresh"]
	refresh := 0

	if ok && len(refreshvalue[0]) > 0 {
		refresh, _ = strconv.Atoi(refreshvalue[0])
	}

	nrlines, ok := req.URL.Query()["lines"]
	maxlines := 25

	if ok && len(nrlines[0]) > 0 {
		maxlines, _ = strconv.Atoi(nrlines[0])
	}

	reversed, ok := req.URL.Query()["reverse"]
	reverse := 0

	if ok && len(reversed[0]) > 0 {
		reverse, _ = strconv.Atoi(reversed[0])
		if reverse > 1 { reverse = 1 }
	}

	filtering, ok := req.URL.Query()["filter"]
	filter := 0

	if ok && len(filtering[0]) > 0 {
		filter, _ = strconv.Atoi(filtering[0])
		if filter > 1 { filter = 1 }
	}

	logname := "/var/log/httpd/" + site + "-" + rtype + ".log"

	stdlog.Println(string(ip) + ": request = " + string(site) + ",type = " + string(rtype))

	fmt.Fprintln(w, "<html>")
	fmt.Fprintln(w, "<head>")

	if refresh > 0 {
		fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"%d\">\n", refresh)
	}

	fmt.Fprintln(w, "<style>")
	fmt.Fprintln(w, "body,td,th {font-family: courier; font-size: 12px; }")
	fmt.Fprintln(w, ".switchType {text-decoration: none; background-color: #95d1de; color: #000000; }")
	fmt.Fprintln(w, ".switchMore {text-decoration: none; background-color: #eee; color: #000000; }")
	fmt.Fprintln(w, ".switchLess {text-decoration: none; background-color: #eee; color: #000000; }")
	fmt.Fprintln(w, ".switchRefresh {text-decoration: none; background-color: #f5d193; color: #000000; }")
	fmt.Fprintln(w, ".switchReverse {text-decoration: none; background-color: #eee; color: #000000; }")
	fmt.Fprintln(w, ".switchFilter {text-decoration: none; background-color: #195e83; color: #eee; }")
	fmt.Fprintln(w, "</style>")
	fmt.Fprintln(w, "</head>")
	fmt.Fprintln(w, "<body>")

	url := "error"

	if rtype == "error" {
		url = "access"
	}

	url = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchType\">SHOW %s</a>", url, refresh, filter, maxlines, reverse, site, url)

	showmore := fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchMore\">RANGE +25</a>", rtype, refresh, filter, maxlines+25, reverse, site)

	showless := ""

	if maxlines > 25 {
		showless = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&lines=%d&reverse=%d&filter=%d&site=%s\" class=\"switchLess\">RANGE -25</a>", rtype, refresh, maxlines-25, reverse, filter, site)
	}

	inverse := 0

	if reverse == 0 {
		inverse = 1
	}

	goreverse := fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchReverse\">REVERSE SORT</a>", rtype, refresh, filter, maxlines, inverse, site)

	gorefresh := ""

	if refresh == 0 {
		gorefresh = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=10&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchRefresh\">REFRESH 10s</a>", rtype, filter, maxlines, reverse, site)
	} else {
		gorefresh = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=0&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchRefresh\">REFRESH off</a>", rtype, filter, maxlines, reverse, site)
	}

	gofilter := ""

	if filter == 0 {
		gofilter = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=1&lines=%d&reverse=%d&site=%s\" class=\"switchFilter\">FILTER on</a>", rtype, refresh, maxlines, reverse, site)
	} else {
		gofilter = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=0&lines=%d&reverse=%d&site=%s\" class=\"switchFilter\">FILTER off</a>", rtype, refresh, maxlines, reverse, site)
	}

	fmt.Fprintf(w, "<b>Logfile: %s</b> %s %s %s %s %s %s (%s)<br />\n", logname, url, showmore, showless, goreverse, gorefresh, gofilter, ip)

	file, err := os.Open(logname)

	if err != nil {
		fmt.Fprintf(w, "Cannot locate the given site log? (%s)", logname)
		return
	}
	defer file.Close()

	// Idea: skip to last section of the file to avoid reading the file in full?
	// Consideration: what if the last section only contains /_... entries? => read back another section and retry?

	Scanner := bufio.NewScanner(file)

	totallines := 0
	lines := make([]string, maxlines)
	currentline := 0
	overflowed := 0
	filtered := 0
	words := make([]string, 0)

	for Scanner.Scan() {
		aline := Scanner.Text()
		totallines++

		if strings.Contains(aline, "/_coder") {
			continue
		}
		if strings.Contains(aline, "/_log") {
			continue
		}

		if filter == 1 {
			words = strings.Fields(aline)
			if words[0] != ip {
				continue
			}
		}

		lines[currentline] = aline
		currentline++

		if overflowed == 0 {
			filtered++
		}

		if currentline == maxlines {
			currentline = 0
			overflowed = 1
		}
	}

	if err := Scanner.Err(); err != nil {
		errlog.Fatal(err)
	}

	fmt.Fprintf(w, "<b>Total lines in file = %d. Displaying filtered lines only (found = %d,maximum = %d).</b><br /><br />\n", totallines, filtered, maxlines)

	if totallines == 0 {
		fmt.Fprintf(w, "(no entries found in the logfile)\n")
	}

	if overflowed == 0 {
		if reverse == 0 {
			for i := 0; i < currentline; i++ {
				fmt.Fprintf(w, "%s<br />\n", lines[i])
			}
		} else {
			for i := currentline; i > 0; i-- {
				fmt.Fprintf(w, "%s<br />\n", lines[i-1])
			}
		}
	}

	if overflowed == 1 {
		if reverse == 0 {
			for i := 0; i < maxlines; i++ {
				fmt.Fprintf(w, "%s<br />\n", lines[currentline])
				currentline++
				if currentline == maxlines {
					currentline = 0
				}
			}
		} else {
			currentline--
			for i := 0; i < maxlines; i++ {
				if currentline < 0 {
					currentline = maxlines - 1
				}
				fmt.Fprintf(w, "%s<br />\n", lines[currentline])
				currentline--
			}
		}
	}

	fmt.Fprintln(w, "</body>")
	fmt.Fprintln(w, "</html>")
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: logserverd install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/_log", handleAccessLog)

	go func() {
		http.ListenAndServe(":7000", nil)
	}()

	stdlog.Println("Service started, listening on port 7000")

	for {
		select {
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}
}

func init() {
	stdlog = log.New(os.Stdout, "", 0)
	errlog = log.New(os.Stderr, "", 0)
}

func main() {
	srv, err := daemon.New(name, description)

	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}

	service := &Service{srv}
	status, err := service.Manage()

	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
