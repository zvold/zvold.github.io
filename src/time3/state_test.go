package main

import (
	"testing"
	"time"
)

// *MockClock satisfies the Clock interface.
type MockClock struct {
	now time.Time
}

func (c MockClock) Now() time.Time {
	return c.now
}

var mockClock = MockClock{
	now: time.UnixMilli(1_000_000_000_000),
}

func init() {
	// Override the global variable in the main file.
	clock = &mockClock
}

func Test_State_toJson(t *testing.T) {
	state := State{
		work:      12_345 * time.Millisecond,
		rest:      67_890 * time.Millisecond,
		mode:      Off,
		modeStart: clock.Now(),
	}

	mockClock.now = mockClock.now.Add(20_000 * time.Millisecond)
	want := `{"mode": "off", "work": 12.35, "rest": 67.89, "modeStart": 20000}`
	got := state.toJson()
	if got != want {
		t.Errorf("state.toJson(), want: %v, got: %v", want, got)
	}
}

func Test_State_resetModeStart(t *testing.T) {
	state := State{
		work:      10_000 * time.Millisecond,
		rest:      20_000 * time.Millisecond,
		mode:      Work,
		modeStart: clock.Now(),
	}

	// Reset to 10_500 millis later, 'work' duration should increase by 10_500.
	mockClock.now = mockClock.now.Add(10_500 * time.Millisecond)
	state.resetModeStart()

	want := State{
		work:      20_500 * time.Millisecond, // work: +10_500 ms.
		rest:      20_000 * time.Millisecond,
		mode:      Work,
		modeStart: clock.Now(), // modeStart: now
	}

	if state != want {
		t.Errorf("state.resetModeStart(), want: %s, got: %s", &want, &state)
	}
}

func Test_State_resetModeStart_backwards(t *testing.T) {
	state := State{
		work:      10_000 * time.Millisecond,
		rest:      20_000 * time.Millisecond,
		mode:      Work,
		modeStart: clock.Now(),
	}

	want := State{
		work:      state.work,
		rest:      state.rest,
		mode:      state.mode,
		modeStart: state.modeStart,
	}

	// Reset backwards in time, should be no-op.
	mockClock.now = mockClock.now.Add(-10 * time.Hour)
	state.resetModeStart()

	if state != want {
		t.Errorf("state.resetModeStart(), want: %s, got: %s", &want, &state)
	}
}

func Test_State_resetModeStart_off(t *testing.T) {
	state := State{
		work:      10_000 * time.Millisecond,
		rest:      20_000 * time.Millisecond,
		mode:      Off,
		modeStart: clock.Now(),
	}

	mockClock.now = mockClock.now.Add(10 * time.Hour)

	want := State{
		work:      state.work,
		rest:      state.rest,
		mode:      state.mode,
		modeStart: clock.Now(),
	}

	state.resetModeStart()

	if state != want {
		t.Errorf("state.resetModeStart(), want: %s, got: %s", &want, &state)
	}
}

func Test_patchDuration_positive(t *testing.T) {
	field := 45 * time.Second

	want := field + 65*time.Second
	patchDuration(&field, "+1m5s")
	if field != want {
		t.Errorf("patchDuration(), want: %+v, got: %+v", want, field)
	}

	want += 10 * time.Second
	patchDuration(&field, "10s")
	if field != want {
		t.Errorf("patchDuration(), want: %+v, got: %+v", want, field)
	}
}

func Test_patchDuration_negative(t *testing.T) {
	field := 45 * time.Second

	want := field - 25*time.Second
	patchDuration(&field, "-25s")
	if field != want {
		t.Errorf("patchDuration(), want: %+v, got: %+v", want, field)
	}

	// Minimum resulting value is capped at 1 second.
	want = 1 * time.Second
	patchDuration(&field, "-1h")
	if field != want {
		t.Errorf("patchDuration(), want: %+v, got: %+v", want, field)
	}
}

func Test_patchDuration_invalid(t *testing.T) {
	field := 45 * time.Second

	want := field
	patchDuration(&field, "aghrbwll")
	if field != want {
		t.Errorf("patchDuration(), want: %+v, got: %+v", want, field)
	}
}

func Test_patchDurations(t *testing.T) {
	state := State{
		work:      10 * time.Second,
		rest:      20 * time.Second,
		mode:      Work,
		modeStart: clock.Now(),
	}

	mockClock.now = mockClock.now.Add(50 * time.Second)
	want := State{
		work:      state.work + 50*time.Second - 20*time.Second,
		rest:      state.rest + 40*time.Second,
		mode:      state.mode,
		modeStart: clock.Now(),
	}

	state.patchDurations( /*work=*/ "-20s" /*rest=*/, "40s")

	if state != want {
		t.Errorf("patchDurations(), want: %s, got: %s", &want, &state)
	}
}

func Test_changeMode(t *testing.T) {
	state := State{
		work:      10 * time.Second,
		rest:      20 * time.Second,
		mode:      Work,
		modeStart: clock.Now(),
	}

	mockClock.now = mockClock.now.Add(50 * time.Second)
	want := State{
		work:      state.work + 50*time.Second,
		rest:      state.rest,
		mode:      Rest,
		modeStart: clock.Now(),
	}

	state.changeMode("rest")
	if state != want {
		t.Errorf("patchDurations(), want: %s, got: %s", &want, &state)
	}

	state.changeMode("blarrhgh")
	if state != want {
		t.Errorf("patchDurations(), want: %s, got: %s", &want, &state)
	}
}
