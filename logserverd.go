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
	fmt.Println("use /_log?site=<sitename>&type=<access|error>&filter=<0|1>&reverse=<0|1>&refresh=<seconds>&lines=<nr of lines>")
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

	// Default is to use the req.Host name for identifying the current site (web host) name. This parameter
	// makes it possible to view logfiles for a totally different site.
	// TODO: lock the site name to the web host by configuration (cannot read other sites logfiles)

	sites, ok := req.URL.Query()["site"]
	site := req.Host

	if ok && len(sites[0]) > 0 {
		if checkLogname(sites[0]) {
			site = sites[0]
		}
	}

	// Default is to show the ERROR logile. Alternative is to show the ACCESS logfile.

	rtypes, ok := req.URL.Query()["type"]
	rtype := "error"

	if ok && len(rtypes[0]) > 0 {
		if rtypes[0]=="access" { rtype = "access" }
	}

	// Default is not to refresh the web page automatically. If set > 0, the page will refresh every X seconds.

	refreshvalue, ok := req.URL.Query()["refresh"]
	refresh := 0

	if ok && len(refreshvalue[0]) > 0 {
		refresh, _ = strconv.Atoi(refreshvalue[0])
	}

	// Default is to show the 25 lines of the bottom of the log file. This number can be changed (must be above 5).
	nrlines, ok := req.URL.Query()["lines"]
	maxlines := 25

	if ok && len(nrlines[0]) > 0 {
		maxlines, _ = strconv.Atoi(nrlines[0])
		if maxlines < 5 { maxlines = 5 }
	}

	// Default is to show oldest to newest log lines. This order can be reversed with this parameter.

	reversed, ok := req.URL.Query()["reverse"]
	reverse := 0

	if ok && len(reversed[0]) > 0 {
		reverse, _ = strconv.Atoi(reversed[0])
		if reverse > 1 { reverse = 1 }
	}

	// Default is to show all log lines. When set to 1 only lines from our current ip address will be shown.

	filtering, ok := req.URL.Query()["filter"]
	filter := 0

	if ok && len(filtering[0]) > 0 {
		filter, _ = strconv.Atoi(filtering[0])
		if filter > 1 { filter = 1 }
	}

	// Assemble the log name (full path)
	// TODO: allow for a different base path (more flexible where the logs are to be found)
	// TODO: allow for a different type of log? (more flexible for not only webserver logs)
	// TODO: allow for one or more specific file names only? (more restrictive)
	// TODO: consider multiple configuration items in sections?

	logname := "/var/log/httpd/" + site + "-" + rtype + ".log"

	// Log our request

	stdlog.Println(string(ip) + ": request = " + string(site) + ",type = " + string(rtype))

	// Start output to the browser

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

	// Prepare our click items at the top of the page.

	// --- Switch between ERROR and ACCESS type log files

	url := "error"

	if rtype == "error" {
		url = "access"
	}

	url = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchType\">SHOW %s</a>", url, refresh, filter, maxlines, reverse, site, url)

	// --- Show more lines of the log file (+25)

	showmore := fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchMore\">RANGE +25</a>", rtype, refresh, filter, maxlines+25, reverse, site)

	// --- Show less line of the log file (-25)

	showless := ""

	if maxlines > 25 {
		showless = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&lines=%d&reverse=%d&filter=%d&site=%s\" class=\"switchLess\">RANGE -25</a>", rtype, refresh, maxlines-25, reverse, filter, site)
	}

	// --- Show reverse sort of the log lines (newest to oldest and vice versa)

	inverse := 0

	if reverse == 0 {
		inverse = 1
	}

	goreverse := fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchReverse\">REVERSE SORT</a>", rtype, refresh, filter, maxlines, inverse, site)

	// --- Show option to switch refresh on (10 seconds) or off

	gorefresh := ""

	if refresh == 0 {
		gorefresh = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=10&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchRefresh\">REFRESH 10s</a>", rtype, filter, maxlines, reverse, site)
	} else {
		gorefresh = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=0&filter=%d&lines=%d&reverse=%d&site=%s\" class=\"switchRefresh\">REFRESH off</a>", rtype, filter, maxlines, reverse, site)
	}

	// --- Show option to filter on our ip address on/off

	gofilter := ""

	if filter == 0 {
		gofilter = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=1&lines=%d&reverse=%d&site=%s\" class=\"switchFilter\">FILTER on</a>", rtype, refresh, maxlines, reverse, site)
	} else {
		gofilter = fmt.Sprintf("<a href=\"/_log?type=%s&refresh=%d&filter=0&lines=%d&reverse=%d&site=%s\" class=\"switchFilter\">FILTER off</a>", rtype, refresh, maxlines, reverse, site)
	}

	// Output our log name and menu to the browser.

	fmt.Fprintf(w, "<b>Logfile: %s</b> %s %s %s %s %s %s (%s)<br />\n", logname, url, showmore, showless, goreverse, gorefresh, gofilter, ip)

	// Open the log file for reading.

	file, err := os.Open(logname)

	// If we fail to read the requested log file show an error.

	if err != nil {
		fmt.Fprintf(w, "Cannot locate the given site log? (%s)", logname)
		return
	}

	// Close the file when we are done with this process.

	defer file.Close()

	// TODO: Idea: skip to last section of the file to avoid reading the file in full?
	// TODO: Consideration: what if the last section only contains /_... entries? => read back another section and retry?

	// Start a new reader

	Scanner := bufio.NewScanner(file)

	// Prepare counters and buffers (cyclic and line elements).

	totallines := 0
	lines := make([]string, maxlines)
	currentline := 0
	overflowed := 0
	filtered := 0
	words := make([]string, 0)

	// Read the logfile line per line

	for Scanner.Scan() {
		aline := Scanner.Text()
		totallines++

		// Skip content we are not interested in.

		// --- Skip /_coder (ICEcoder location)

		if strings.Contains(aline, "/_coder") {
			continue
		}

		// --- Skip /_log (our deamon location)

		if strings.Contains(aline, "/_log") {
			continue
		}

		// --- If filtering is on, skip lines not containing our ip address in the first element of the log line

		if filter == 1 {
			words = strings.Fields(aline)
			if words[0] != ip {
				continue
			}
		}

		// Store the new content.

		lines[currentline] = aline
		currentline++

		// If we are not at the end of the buffer, count the number of lines stored.

		if overflowed == 0 {
			filtered++
		}

		// Once we reach the end of the buffer, start 'overflow' mode: overwrite our cyclic buffer.

		if currentline == maxlines {
			currentline = 0
			overflowed = 1
		}
	}

	// If an error occurred, we log the error.

	if err := Scanner.Err(); err != nil {
		errlog.Fatal(err)
	}

	// Output some statistics.

	fmt.Fprintf(w, "<b>Total lines in file = %d. Displaying filtered lines only (found = %d,maximum = %d).</b><br /><br />\n", totallines, filtered, maxlines)

	if totallines == 0 {
		fmt.Fprintf(w, "(no entries found in the logfile)\n")
	}

	// If we did not overflow, processing is straightforward (start at 0, stop at currentline).

	if overflowed == 0 {
		// If reversed perform downcount.

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

	// If we did overflow we need to include the cyclic content once we hit the end (or the start) of the cyclic buffer.

	if overflowed == 1 {
		if reverse == 0 {
			// Our currentline is the oldest line in the cyclic buffer (currentline is always 1 ahead).

			for i := 0; i < maxlines; i++ {
				fmt.Fprintf(w, "%s<br />\n", lines[currentline])
				currentline++
				if currentline == maxlines {
					currentline = 0
				}
			}
		} else {
			// Otherwise we have to reverse display, starting at the last entered line in the cyclic buffer.

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

	// Finish our output to the browser.

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
