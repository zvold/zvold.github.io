A simple time tracker, in the spirit of time management approaches like:

-  [Pomodoro Technique](https://en.wikipedia.org/wiki/Pomodoro_Technique)
-  [52/17 rule](https://en.wikipedia.org/wiki/52/17_rule)
-  [Third Time](https://www.lesswrong.com/posts/RWu8eZqbwgB9zaerh/third-time-a-better-way-to-work)
## Usage

1. Start the server (see below for flags).

   ```
   go run github.com/zvold/zvold.github.io/src/time3@v0.8.0 --db=time3.db
   ```

2. Start the client(s).

    Open `http://hostname:37177` in the browser. Use the optional URL parameter `?t=` to set the "target" work/rest ratio.

    If your server has `gnuplot` installed, the client should show a graph with historical values from the database.

3. Toggle the mode (`work` / `rest` / `off the clock`) appropriately.

    This can be done from any client. The state is maintained on the server, and clients are eventually consistent.

    Refreshing the page will get the up-to-date state from the server.

## Command-line flags

- `-port=<num>` : Change the HTTP port the server will listen on.

- `-https` : Start an _additional_ HTTPs server on `port+1`. This expects SSL certificate files `server.crt` and `server.key` to be present.

- `-db=<path>` : Set sqlite database file name for recording the daily work/rest totals. The file will be created if it doesn't exist.

- `-v` : Enable more verbose server logs.
