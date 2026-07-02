package domain

import "testing"

func TestReviewTransitions(t *testing.T) {
	cases := []struct {
		from ReviewStatus
		to   ReviewStatus
		ok   bool
	}{
		{ReviewPending, ReviewApproved, true},
		{ReviewPending, ReviewRejected, true},
		// Illegal: skipping review or re-reviewing a terminal version.
		{ReviewPending, ReviewPending, false},
		{ReviewApproved, ReviewRejected, false},
		{ReviewApproved, ReviewApproved, false},
		{ReviewRejected, ReviewApproved, false},
		{ReviewRejected, ReviewPending, false},
	}
	for _, tc := range cases {
		if got := tc.from.CanTransitionTo(tc.to); got != tc.ok {
			t.Errorf("%s -> %s: got %v, want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}
