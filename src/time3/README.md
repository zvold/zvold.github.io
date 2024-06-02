A simple time tracker, in the spirit of time management approaches like:

-  [Pomodoro Technique](https://en.wikipedia.org/wiki/Pomodoro_Technique)
-  [52/17 rule](https://en.wikipedia.org/wiki/52/17_rule)
-  [Third Time](https://www.lesswrong.com/posts/RWu8eZqbwgB9zaerh/third-time-a-better-way-to-work)
-  etc.

Usage:

1. Start the server.

   ```
   go run github.com/zvold/zvold.github.io/src/time3@latest
   ```

2. Start the client(s).

   Open http://localhost:37177 in the browser. Use the optional URL parameter `?t=` to set the "target" work/rest ratio.

3. Toggle the mode (`work` / `rest` / `off`) appropriately.

   This can be done from any client. The state is maintained on the server, and clients are eventually consistent.

   It's always safe to simply refresh the page to get the up-to-date state.
