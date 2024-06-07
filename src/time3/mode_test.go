package main

import "testing"

func Test_ModeType_toString(t *testing.T) {
	assertToString(Work, "work", t)
	assertToString(Rest, "rest", t)
	assertToString(Off, "off", t)
}

func Test_ModeType_toString_invalid(t *testing.T) {
	assertToString(-1, "unknown", t)
	assertToString(3, "unknown", t)
}

func Test_ModeType_modeFromString(t *testing.T) {
	assertFromString("work", Work, t)
	assertFromString("rest", Rest, t)
	assertFromString("off", Off, t)
}

func Test_ModeType_modeFromString_invalid(t *testing.T) {
	got := modeFromString("argh")
	if got != nil {
		t.Errorf(`modeFromString("argh"), want: nil, got: (ModeType = %v).`, *got)
	}
}

func assertToString(m ModeType, want string, t *testing.T) {
	got := m.toString()
	if got != want {
		t.Errorf(`(ModeType = %v).toString(), want: %s, got: %s.`, m, want, got)
	}
}

func assertFromString(s string, want ModeType, t *testing.T) {
	got := modeFromString(s)
	if *got != want {
		t.Errorf(`modeFromString(%s), want: (ModeType = %v), got: (ModeType = %v).`,
			s, want, *got)
	}
}
