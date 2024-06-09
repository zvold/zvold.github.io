package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var portFlag = flag.Int("port", 37177, "Port to listen on.")

//go:embed template.html
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

// The set of remote hosts (IP + user agent) seen during the operation.
type RemoteHosts struct {
	sync.Mutex
	set map[string]bool
}

// Adds the key to 'hosts' set and returns 'true' if it was a new key.
func (hosts *RemoteHosts) add(key string) bool {
	hosts.Lock()
	defer hosts.Unlock()
	if _, ok := hosts.set[key]; ok {
		return false
	}
	hosts.set[key] = true
	return true
}

var remoteHosts = RemoteHosts{
	set: make(map[string]bool),
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
	work:      time.Second,
	rest:      time.Second,
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
		state.modeStart.UnixMilli())
}

// Resets 'modeStart' to 'time.Now()', and updates the 'work' and 'rest' times.
// Assumes the mutex is locked and unlocked by the caller.
func (state *State) resetModeStart() {
	var now = clock.Now()
	var duration = now.Sub(state.modeStart)
	if duration < 0 {
		fmt.Println("Resetting backwards in time, ignoring.")
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
		fmt.Printf("Unknown mode specified: %v, ignoring.\n", modeString)
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
		fmt.Printf("Cannot parse duration: %v, ignoring.\n", str)
		return
	}
	*field += duration
	if *field < time.Second {
		*field = time.Second
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

// Constructs a human-readable string representing the remote host.
func getRemoteHost(r *http.Request) (key string) {
	key = r.RemoteAddr
	host, _, err := net.SplitHostPort(key)
	if err == nil {
		key = host
	}
	return fmt.Sprintf("%s | %s | %s", key, r.Header.Get("X-Forwarded-For"), r.UserAgent())
}

// JsonRequest struct represents the body of an HTTP POST request.
type JsonRequest struct {
	Mode string
	Work string
	Rest string
}

// Returns JsonRequest for the HTTP request body, or nil in case of errors.
func parseRequestBody(r *http.Request) (*JsonRequest, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Error reading request body: %v", err)
	}

	var result JsonRequest
	err = json.Unmarshal(bodyBytes, &result)
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
		state.changeMode(jsonRequest.Mode)
		clients.broadcast(state.toJson())
	}
}

// Starts an HTTP server which:
//   - Responds with a full HTML page to HTTP GET requests. The page will
//     represent the current State.
//   - Accepts requests for mode changes via HTTP POST and responds with a JSON
//     encoded current State.
func main() {
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := getRemoteHost(r)
		if remoteHosts.add(key) {
			fmt.Printf("New remote host:\t%s\n", key)
		}

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
				fmt.Fprintf(w, state.toJson())
			}()

			jsonRequest, err := parseRequestBody(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			handleJsonRequest(jsonRequest)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			fmt.Printf("websocket upgrade failed, closing: %s\n", err)
			return
		}

		client := WsClient{conn: c}
		clients.add(&client)

		defer func() {
			clients.remove(&client)
			c.Close()
		}()

		fmt.Println("websocket connection established, looping...")
		err = client.send(state.toJson())
		if err == nil {
			fmt.Printf("sent the current state: %s\n", state.String())
		}

		for {
			mtype, message, err := c.ReadMessage()
			if err != nil {
				fmt.Printf("websocket read error, closing: %s\n", err)
				break
			}

			fmt.Printf("websocket message received (type %d): %s\n", mtype, message)
			switch mtype {
			case websocket.TextMessage:
				var jsonRequest JsonRequest
				err = json.Unmarshal(message, &jsonRequest)
				if err != nil {
					fmt.Printf("Error while unmarshalling, ignoring: %s\n", err)
					break
				}
				handleJsonRequest(&jsonRequest)
			case websocket.CloseMessage:
				fmt.Printf("websocket close received, closing.\n")
				return
			default:
				// no-op
			}
		}
	})

	fmt.Printf("Server listening on port %d...\n", *portFlag)
	http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), nil)
}
