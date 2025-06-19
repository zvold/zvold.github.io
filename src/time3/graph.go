package main

import (
	"bytes"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

//go:embed daily_totals_template.gnuplot
//go:embed dummy_graph.png
var f2 embed.FS

type options struct {
	date   string
	days   int
	width  int
	height int
}

var datePattern2 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Construct valid 'options' or returns an error for invalid options.
func newOptions(date string, days, width, height int) (*options, error) {
	if !datePattern2.MatchString(date) {
		return nil, fmt.Errorf("Invalid date: '%s'.", date)
	}
	if days < 1 || days > 31 {
		return nil, fmt.Errorf("Invalid days: '%d'.", days)
	}
	if width < 200 || width > 2000 {
		return nil, fmt.Errorf("Invalid width: '%d'.", width)
	}
	if height < 100 || height > 1000 {
		return nil, fmt.Errorf("Invalid width: '%d'.", height)
	}
	return &options{
		date:   date,
		days:   days,
		width:  width,
		height: height,
	}, nil
}

func graphPageHandler(db *Database) func(a http.ResponseWriter, b *http.Request) {
	dummyImage, _ := f2.ReadFile("dummy_graph.png")
	scriptTemplate, _ := f2.ReadFile("daily_totals_template.gnuplot")

	return func(w http.ResponseWriter, r *http.Request) {
		logNewPeer(r)

		if db == nil {
			if dummyImage == nil {
				http.Error(w, "Dummy graph not found.", http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "image/png")
				if _, err := w.Write(dummyImage); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			return
		}

		if scriptTemplate == nil {
			http.Error(w, "Gnuplot script not found.", http.StatusInternalServerError)
		}

		// The request has to specify:
		//   - 'date', the latest date to plot
		//   - 'n', optional number of historical days to plot, defaults to 7
		//   - 'w', optional width of the image in pixels, defaults to 1200
		//   - 'h', optional height of the image in pixels, defaults to 600

		params := r.URL.Query()
		if len(params) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		opts, err := newOptions(
			params.Get("date"),
			parseInt(params.Get("n"), 7),
			parseInt(params.Get("w"), 1200),
			parseInt(params.Get("h"), 600),
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("querying the database.", "opts", opts)

		png, err2 := plotGraph(db, opts, string(scriptTemplate))
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		if _, err3 := w.Write(png); err3 != nil {
			http.Error(w, err3.Error(), http.StatusInternalServerError)
		}
	}
}

func formatDate(d time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d", d.Year(), d.Month(), d.Day())
}

// Execs 'gnuplot' with data, returns the resulting png image.
func plotGraph(db *Database, opts *options, script string) ([]byte, error) {
	t2, _ := time.Parse(time.RFC3339, opts.date+"T00:00:01Z")
	t1 := t2.AddDate(0, 0, -opts.days+1)

	data := db.ReadTotals(formatDate(t1), opts.date)
	hasPrefix := func(p string) func(s string) bool {
		return func(s string) bool {
			return strings.HasPrefix(s, p)
		}
	}
	// Add explicit 0/0 data on the boundary days, so they're plotted regardless.
	if !slices.ContainsFunc(data, hasPrefix(formatDate(t1))) {
		data = append(data, formatDate(t1)+" 0 0")
	}
	if !slices.ContainsFunc(data, hasPrefix(opts.date)) {
		data = append(data, opts.date+" 0 0")
	}

	slog.Debug("read rows:", "data", data)

	dataStr := dataToStr(data)

	scriptStr := script
	scriptStr = strings.ReplaceAll(scriptStr, "%WIDTH%", fmt.Sprintf("%d", opts.width))
	scriptStr = strings.ReplaceAll(scriptStr, "%HEIGHT%", fmt.Sprintf("%d", opts.height))
	scriptStr = strings.ReplaceAll(scriptStr, "%XMIN%", fmt.Sprintf("%d", t1.UnixMilli()/1000-42200))
	scriptStr = strings.ReplaceAll(scriptStr, "%XMAX%", fmt.Sprintf("%d", t2.UnixMilli()/1000+42200))

	cmd := exec.Command("gnuplot")

	cmd.Stdin = strings.NewReader(scriptStr + dataStr)

	var buffer bytes.Buffer
	cmd.Stdout = &buffer

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func dataToStr(data []string) string {
	result := "date work rest\n"
	for _, s := range data {
		result += s + "\n"
	}
	result += "e\n"
	return result + result
}

func parseInt(s string, d int) int {
	r, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return r
}
