package main

import "testing"

func TestUnlockedPolicy(t *testing.T) {
	for _, tc := range []struct{name, text string; want bool}{
		{"unlocked", " showing=false\n interactiveState=INTERACTIVE_STATE_AWAKE\n mIsShowing=false", true},
		{"locked", "showing=true\ninteractiveState=INTERACTIVE_STATE_AWAKE", false},
		{"asleep", "showing=false\ninteractiveState=INTERACTIVE_STATE_ASLEEP", false},
		{"transition", "showing=false\ninteractiveState=INTERACTIVE_STATE_AWAKE\nmIsShowing=true", false},
		{"empty", "", false},
		{"incomplete", "showing=false", false},
	} {
		t.Run(tc.name, func(t *testing.T) { if got := unlockedPolicy(tc.text); got != tc.want { t.Fatalf("got %v, want %v", got, tc.want) } })
	}
}
