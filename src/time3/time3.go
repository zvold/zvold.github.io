package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"

	"context"
	"time"

	"github.com/gorilla/websocket"
)

var portFlag = flag.Int("port", 37177, "Port on which HTTP server will listen.")

var httpsFlag = flag.Bool("https", false, "Set to 'true' to start HTTPs server on port+1."+
	" This requires 'server.crt' and 'server.key' files to be present.")

var verboseFlag = flag.Bool("v", false, "Set to 'true' for more verbose logging.")

var dbFlag = flag.String("db", "", "Database file to use. Database is not enabled when not set.")

//go:embed template.html
//go:embed tomato.ico
var f embed.FS

// Maintains a map of all currently connected clients.
var clients = WsClients{
	clients: make(map[*WsClient]int),
}

// Clock interface is injected for better testing.
type Clock interface {
	Now() time.Time
}

// *NormalClock satisfies the Clock interface.
type NormalClock struct{}

func (c *NormalClock) Now() time.Time {
	return time.Now()
}

var clock Clock = &NormalClock{}

// Host Info stores info about the remote host that connected to our server.
type HostInfo struct {
	ip        string
	userAgent string
}

func (h *HostInfo) String() string {
	return fmt.Sprintf("%s|%s", h.ip, h.userAgent)
}

// The set of remote hosts seen during the operation, and how many times each connected.
type RemoteHosts struct {
	sync.Mutex
	set map[HostInfo]uint64
}

// Adds the key to 'hosts' set and returns 'true' if it was a new key.
func (hosts *RemoteHosts) add(key HostInfo) bool {
	hosts.Lock()
	defer hosts.Unlock()
	if _, ok := hosts.set[key]; ok {
		hosts.set[key]++
		return false
	}
	hosts.set[key] = 1
	return true
}

// Formats hosts summary as a table and returns it in a string.
func (hosts *RemoteHosts) asTable() (info string) {
	hosts.Lock()
	defer hosts.Unlock()

	// Print all remote hosts in decreasing order of number of occurrences.
	keys := slices.SortedFunc(
		maps.Keys(hosts.set),
		func(a, b HostInfo) int {
			v := int(hosts.set[b]) - int(hosts.set[a])
			if v != 0 {
				return v
			}
			return strings.Compare(b.userAgent, a.userAgent)
		},
	)

	info += fmt.Sprintf("| %16s | %130s | %5s |\n", "ip", "User Agent", "Count")
	info += fmt.Sprintf("| %s | %s | %s |\n",
		strings.Repeat("-", 16), strings.Repeat("-", 130), strings.Repeat("-", 5))
	for _, k := range keys {
		info += fmt.Sprintf("| %16s | %130s | %5d |\n", k.ip, k.userAgent, hosts.set[k])
	}
	return
}

// Pretty-prints the remote hosts summary to the log.
func (hosts *RemoteHosts) log() {
	slog.Info(fmt.Sprintf("remote hosts seen so far: \n%s", hosts.asTable()))
}

var remoteHosts = RemoteHosts{
	set: make(map[HostInfo]uint64),
}

// ModeType is an enum representing the possible "modes": work, rest and off.
type ModeType int

const (
	Work ModeType = iota
	Rest
	Off
)

// Returns a string representation of a given ModeType.
func (m ModeType) toString() string {
	if m < 0 || m > 2 {
		return "unknown"
	}
	return []string{"work", "rest", "off"}[m]
}

// Returns a *ModeType for a given string, or nil for invalid strings.
func modeFromString(str string) *ModeType {
	var result ModeType
	switch str {
	case "work":
		result = Work
	case "rest":
		result = Rest
	case "off":
		result = Off
	default:
		return nil
	}
	return &result
}

// State struct represents the full state of the punch clock. Note that it is
// mutated only when the mode is changed. That is, the 'work' and 'rest' fields
// don't reflect the "total" durations on their own.
type State struct {
	sync.Mutex
	work      time.Duration // Duration of time spent working.
	rest      time.Duration // Duration of time spent resting.
	mode      ModeType      // Current mode.
	modeStart time.Time     // Time of the last mode switch.
}

// Initialize the punch clock. It starts in the 'off' mode.
var state = State{
	work:      0,
	rest:      0,
	mode:      Off,
	modeStart: clock.Now(),
}

// Populates the HTML page according to the current state and writes it to the
// provided ResponseWriter.
func (state *State) writeHtmlResponse(
	w http.ResponseWriter, tmpl *template.Template) error {

	state.Lock()
	defer state.Unlock()

	// A struct for passing the State to the HTML template.
	data := struct {
		Work      float64 // JS code expects this in seconds.
		Rest      float64 // Same.
		Mode      string
		ModeStart int64 // JS code expects this in milliseconds.
	}{
		Work:      state.work.Seconds(),
		Rest:      state.rest.Seconds(),
		Mode:      state.mode.toString(),
		ModeStart: state.modeStart.UnixMilli(),
	}
	return tmpl.Execute(w, data)
}

// Returns the State as a human-readable string.
func (state *State) String() string {
	return state.toJson()
}

// Returns the State as a JSON string.
func (state *State) toJson() string {
	state.Lock()
	defer state.Unlock()

	return fmt.Sprintf(
		`{"mode": "%s", "work": %.2f, "rest": %.2f, "modeStart": %d}`,
		state.mode.toString(),
		state.work.Seconds(),
		state.rest.Seconds(),
		clock.Now().Sub(state.modeStart).Milliseconds()) // X milliseconds ago.
}

// Resets 'modeStart' to 'time.Now()', and updates the 'work' and 'rest' times.
// Assumes the mutex is locked and unlocked by the caller.
func (state *State) resetModeStart() {
	var now = clock.Now()
	var duration = now.Sub(state.modeStart)
	if duration < 0 {
		slog.Error("resetting backwards in time, ignoring.", "now", now, "modeStart", state.modeStart)
		return
	}

	switch state.mode {
	case Work:
		state.work += duration
	case Rest:
		state.rest += duration
	case Off:
		// No-op
	default:
		panic(fmt.Sprintf("Unhandled mode: %v.", state.mode))
	}

	state.modeStart = now
}

// Changes the current mode (if necessary).
func (state *State) changeMode(modeString string) {
	newMode := modeFromString(modeString)
	if newMode == nil {
		slog.Info("unknown mode specified, ignoring.", "mode", modeString)
		return
	}
	if state.mode == *newMode {
		return
	}
	state.Lock()
	defer state.Unlock()

	state.resetModeStart()
	state.mode = *newMode
}

// Patches the value at time.Duration address. Minimum resulting duration is 1s.
// Assumes the State mutex is handled by the caller as necessary.
func patchDuration(field *time.Duration, str string) {
	if str == "" {
		return
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		slog.Info("invalid duration string, ignoring.", "duration", str)
		return
	}
	*field += duration
	if *field < 0 {
		*field = 0
	}
}

// Patches work/rest durations, based on strings in time.Duration format.
func (state *State) patchDurations(workString string, restString string) {
	state.Lock()
	defer state.Unlock()

	state.resetModeStart()
	patchDuration(&state.work, workString)
	patchDuration(&state.rest, restString)
}

// Returns the total work/rest durations.
func (state *State) getTotalDurations(cutoff time.Time) (work, rest time.Duration) {
	state.Lock()
	defer state.Unlock()

	work, rest = state.work, state.rest

	var duration = cutoff.Sub(state.modeStart)
	if duration < 0 {
		slog.Error("time goes backwards, ignoring.", "cutoff", cutoff, "modeStart", state.modeStart)
		return
	}

	switch state.mode {
	case Work:
		work += duration
	case Rest:
		rest += duration
	case Off:
		// No-op
	default:
		panic(fmt.Sprintf("Unhandled mode: %v.", state.mode))
	}
	return
}

// Constructs a human-readable string representing the remote host.
func getRemoteHost(r *http.Request) HostInfo {
	addr := r.RemoteAddr
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		addr = host
	}
	return HostInfo{
		ip:        truncate(addr, 16),
		userAgent: truncate(r.UserAgent(), 130),
	}
}

func truncate(s string, l int) string {
	if len(s) <= l {
		return s
	}
	return s[:l-3] + "..."
}

// JsonRequest struct represents the body of an HTTP POST request.
type JsonRequest struct {
	Mode string
	Work string
	Rest string
}

// Returns JsonRequest for the HTTP request body, or nil in case of errors.
func parseRequestBody(r *http.Request) (*JsonRequest, error) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var result JsonRequest
	err := decoder.Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("Error while unmarshalling: %v", err)
	}
	return &result, nil
}

// Understands various possibilities present in the JsonRequest and updates the
// state accordingly.
func handleJsonRequest(jsonRequest *JsonRequest) {
	if jsonRequest.Work != "" || jsonRequest.Rest != "" {
		// This is a request for patching work/rest durations.
		state.patchDurations(jsonRequest.Work, jsonRequest.Rest)
		clients.broadcast(state.toJson())
	} else if jsonRequest.Mode != "" {
		// This is a request attempting to update the mode.
		// TODO(zvold): consider updating the daily total on mode changing to 'off'.
		state.changeMode(jsonRequest.Mode)
		clients.broadcast(state.toJson())
	}
}

// Logs the remote peer if it's seen for the first time.
func logNewPeer(r *http.Request) {
	hostInfo := getRemoteHost(r)
	if remoteHosts.add(hostInfo) {
		slog.Info("new remote host.", "host", &hostInfo)
	}
}

// Reads favicon from tomato.ico and writes its contents to the response.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	logNewPeer(r)

	favicon, err := f.ReadFile("tomato.ico")
	if err != nil {
		slog.Error("error reading the icon file.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/x-icon")

	_, err = w.Write(favicon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// On HTTP GET requests, responds with a full HTML/Javascript page representing the current State.
// On HTTP POST requests, executes the requested command (like a mode change) and responds with a
// JSON encoded updated State.
func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	logNewPeer(r)

	if r.Method == "GET" {
		// Request fetching the full HTML page for the current state.
		tmpl, err := template.ParseFS(f, "template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = state.writeHtmlResponse(w, tmpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		// Request (potentially) modifying the state.
		defer func() {
			// Always return the current state, even on invalid requests.
			fmt.Fprintf(w, "%s", state.toJson())
		}()

		jsonRequest, err := parseRequestBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("HTTP POST message received.", "request", jsonRequest)
		handleJsonRequest(jsonRequest)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Upgrades incoming connection to websocket, executes incoming commands (like a mode change), keeps
// track of all connected websocket clients and broadcasts State changes when they happen.
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	logNewPeer(r)

	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Error("websocket upgrade failed, closing.", "error", err)
		return
	}

	client := WsClient{conn: c}
	clients.add(&client)

	defer func() {
		clients.remove(&client)
		c.Close()
	}()

	slog.Debug("websocket connection established, looping...")
	err = client.send(state.toJson())
	if err == nil {
		slog.Debug("sent the current state.", "state", &state)
	}

	for {
		mtype, message, err := c.ReadMessage()
		if err != nil {
			slog.Debug("websocket read error, closing.", "error", err)
			break
		}

		slog.Info("websocket message received.", "type", mtype, "message", message)
		switch mtype {
		case websocket.TextMessage:
			decoder := json.NewDecoder(bytes.NewReader(message))
			decoder.DisallowUnknownFields()

			var jsonRequest JsonRequest
			err := decoder.Decode(&jsonRequest)
			if err != nil {
				slog.Info("error while unmarshalling, ignoring.", "error", err)
				break
			}
			handleJsonRequest(&jsonRequest)
		case websocket.CloseMessage:
			slog.Debug("websocket close received, closing.")
			return
		default:
			// no-op
		}
	}
}

func main() {
	flag.Parse()

	if *verboseFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var db *Database
	if *dbFlag != "" {
		var dbErr error
		db, dbErr = InitDB(*dbFlag)
		if dbErr != nil {
			slog.Error("cannot open database.", "err", dbErr)
			os.Exit(1)
		}

		totalDays := db.DaysCount()
		slog.Info("Logged days:", "count", totalDays)

		db.StartLogger(&state)
	}

	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/ws", websocketHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/graph", graphPageHandler(db))

	// Log cumulative remote hosts stats every hour.
	hostsLogger := time.NewTicker(1 * time.Hour)
	go func() {
		for range hostsLogger.C {
			remoteHosts.log()
		}
	}()

	var wg sync.WaitGroup

	// Start HTTP server on 'port'.
	srv1 := http.Server{Addr: fmt.Sprintf(":%d", *portFlag)}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("HTTP  server started.", "address", srv1.Addr)
		slog.Info("HTTP  server stopped.", "result", srv1.ListenAndServe())
	}()

	// Start an HTTPS server on 'port+1'.
	srv2 := http.Server{Addr: fmt.Sprintf(":%d", *portFlag+1)}
	if *httpsFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("HTTPs server started.", "address", srv2.Addr)
			slog.Info("HTTPs server stopped.", "result", srv2.ListenAndServeTLS("server.crt", "server.key"))
		}()
	}

	// Request both HTTP and HTTPS servers to shutdown when SIGINT is received.
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// Block until daily logger gorouting is stopped.
		if db != nil {
			db.StopLogger()
		}

		hostsLogger.Stop()
		slog.Info("Final remote hosts stats on shutdown:")
		remoteHosts.log()

		if err := srv1.Shutdown(context.Background()); err != nil {
			slog.Error("HTTP  server shutdown error.", "error", err)
		}
		if err := srv2.Shutdown(context.Background()); err != nil {
			slog.Error("HTTPs server shutdown error.", "error", err)
		}
	}()

	// Block until both servers terminate.
	wg.Wait()
}
