package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

//go:embed template.html
var f embed.FS

// IPv4 address on which the server should be reachable.
var serverAddress net.IP

// ModeType is an enum representing the possible "modes": work, rest and off.
type ModeType int

const (
	Work ModeType = iota
	Rest
	Off
)

// Returns a string representation of a given ModeType.
func (m ModeType) toString() string {
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

// State struct represents the full state of the counter. Note that it is
// mutated only when the mode is changed. That is, the 'work' and 'rest' fields
// don't reflect the "total" time on their own.
type State struct {
	sync.Mutex
	work      time.Duration // Duration of time spent working.
	rest      time.Duration // Duration of time spent resting.
	mode      ModeType      // Current mode.
	modeStart time.Time     // Time of the last mode switch.
}

// Initialize the counter in the 'off' mode.
var state = State{
	work:      time.Second,
	rest:      time.Second,
	mode:      Off,
	modeStart: time.Now(),
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
		ModeStart int64  // JS code expects this in milliseconds.
		Address   string // IPv4 address of the server.
	}{
		Work:      state.work.Seconds(),
		Rest:      state.rest.Seconds(),
		Mode:      state.mode.toString(),
		ModeStart: state.modeStart.UnixMilli(),
		Address:   serverAddress.String(),
	}

	return tmpl.Execute(w, data)
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
		state.modeStart.UnixMilli())
}

// Returns the full time as a (work, rest) tuple, in seconds.
// Assumes the caller handles the mutex appropriately.
func (state *State) totalTime() (time.Duration, time.Duration) {
	var duration = time.Since(state.modeStart)
	switch state.mode {
	case Work:
		return state.work + duration, state.rest
	case Rest:
		return state.work, state.rest + duration
	case Off:
		return state.work, state.rest
	}
	panic(fmt.Sprintf("Unhandled mode: %v.", state.mode))
}

// Changes the current mode (if necessary).
func (state *State) changeMode(newMode ModeType) {
	if state.mode == newMode {
		return
	}
	state.Lock()
	defer state.Unlock()

	state.work, state.rest = state.totalTime()
	state.modeStart = time.Now()
	state.mode = newMode
}

// Finds IPv4 non-loopback address among available interfaces, or fails.
func findNonLoopbackAddress() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("Cannot list network interfaces: %v", err)
		os.Exit(1)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4
			}
		}
	}

	fmt.Printf("Cannot find non-loopback IPv4 address.")
	os.Exit(1)

	panic("Unreachable")
}

// Starts an HTTP server which:
//   - Responds with a full HTML page to HTTP GET requests. The page will
//     represent the current State.
//   - Accepts requests for mode changes via HTTP POST and responds with a JSON
//     encoded current State.
func main() {
	serverAddress = findNonLoopbackAddress()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
			// Request (potentially) modifying the mode and getting the new state.
			defer func() {
				// Always return the current state, even on invalid requests.
				fmt.Fprintf(w, state.toJson())
			}()

			bodyBytes, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading request body: %v.", err), http.StatusBadRequest)
				return
			}

			type Request struct {
				Mode string
			}

			var request Request
			err = json.Unmarshal(bodyBytes, &request)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error while unmarshalling: %v.", err), http.StatusBadRequest)
				return
			}

			if len(request.Mode) == 0 {
				// This is just a refresh request (empty body).
				return
			}

			newMode := modeFromString(request.Mode)
			if newMode == nil {
				fmt.Printf("Unknown mode specified: %v\n", request.Mode)
				return
			}

			state.changeMode(*newMode)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Printf("Server listening on %s:37177...\n", serverAddress.String())
	http.ListenAndServe(":37177", nil)
}