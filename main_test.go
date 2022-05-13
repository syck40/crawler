package main

import (
	"testing"
	"time"
)

func TestGetDaysFromMonth(t *testing.T) {
	want := []string{time.Now().Format(dateFormat)}
	got := getDaysFromMonth(time.Now(), dateFormat)
	if want[0] != got[0] {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestGetRemainingMonths(t *testing.T) {
	got := getRemainingMonths(time.Now())
	t.Log(got)
}