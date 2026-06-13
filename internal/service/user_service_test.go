package service

import (
	"testing"
	"time"
)

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name     string
		dob      string // YYYY-MM-DD
		now      string // YYYY-MM-DD
		wantAge  int
	}{
		{
			name:    "birthday already passed this year",
			dob:     "1990-05-10",
			now:     "2025-06-13",
			wantAge: 35,
		},
		{
			name:    "birthday is today",
			dob:     "1990-06-13",
			now:     "2025-06-13",
			wantAge: 35,
		},
		{
			name:    "birthday has not yet occurred this year",
			dob:     "1990-12-31",
			now:     "2025-06-13",
			wantAge: 34,
		},
		{
			name:    "newborn (dob equals now)",
			dob:     "2025-06-13",
			now:     "2025-06-13",
			wantAge: 0,
		},
		{
			name:    "leap year birthday (Feb 29), checked on Feb 28",
			dob:     "2000-02-29",
			now:     "2025-02-28",
			wantAge: 24,
		},
		{
			name:    "leap year birthday (Feb 29), checked on Mar 1",
			dob:     "2000-02-29",
			now:     "2025-03-01",
			wantAge: 25,
		},
		{
			name:    "exactly one year old",
			dob:     "2000-01-01",
			now:     "2001-01-01",
			wantAge: 1,
		},
	}

	const layout = "2006-01-02"
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dob, err := time.Parse(layout, tc.dob)
			if err != nil {
				t.Fatalf("bad dob %q: %v", tc.dob, err)
			}
			now, err := time.Parse(layout, tc.now)
			if err != nil {
				t.Fatalf("bad now %q: %v", tc.now, err)
			}

			got := CalculateAge(dob, now)
			if got != tc.wantAge {
				t.Errorf("CalculateAge(%s, %s) = %d; want %d", tc.dob, tc.now, got, tc.wantAge)
			}
		})
	}
}
