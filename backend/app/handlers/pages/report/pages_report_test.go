package report

import "testing"

func TestMatchesSearch(t *testing.T) {
	cases := []struct {
		name   string
		term   string
		fields []string
		want   bool
	}{
		{"empty matches", "", []string{"Demo Candidate"}, true},
		{"case insensitive", "demo", []string{"Demo Candidate"}, true},
		{"matches later field", "golang", []string{"Demo", "Golang Developer"}, true},
		{"no match", "missing", []string{"Demo", "Golang Developer"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := matchesSearch(tc.term, tc.fields...); got != tc.want {
				t.Fatalf("matchesSearch(%q, %v) = %v, want %v", tc.term, tc.fields, got, tc.want)
			}
		})
	}
}

func TestMatchesDecision(t *testing.T) {
	cases := []struct {
		filter, recommendation string
		want                   bool
	}{
		{"", "pending", true},
		{"pending", "pending", true},
		{"PENDING", "pending", true},
		{"hire", "maybe", false},
	}
	for _, tc := range cases {
		if got := matchesDecision(tc.filter, tc.recommendation); got != tc.want {
			t.Fatalf("matchesDecision(%q, %q) = %v, want %v", tc.filter, tc.recommendation, got, tc.want)
		}
	}
}

func TestMatchesHiringDecision(t *testing.T) {
	cases := []struct {
		filter, status string
		want           bool
	}{
		{"", "hired", true},
		{"hire", "hired", true},
		{"reject", "rejected", true},
		{"hired", "hired", true},
		{"pending", "hired", false},
	}
	for _, tc := range cases {
		if got := matchesHiringDecision(tc.filter, tc.status); got != tc.want {
			t.Fatalf("matchesHiringDecision(%q, %q) = %v, want %v", tc.filter, tc.status, got, tc.want)
		}
	}
}
