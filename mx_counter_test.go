package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnique(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  []string
	}{
		"simple":    {input: []string{"test@test.com", "test@test.com", "other@other.com"}, want: []string{"test@test.com", "other@other.com"}},
		"identical": {input: []string{"test@test.com", "other@other.com"}, want: []string{"test@test.com", "other@other.com"}},
		"empty":     {input: []string{}, want: []string{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := unique(tc.input)
			diff := cmp.Diff(tc.want, got)

			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestGetDomain(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
		err   error
	}{
		"standard":      {input: "hey@gmail.com", want: "gmail.com"},
		"short domain":  {input: "hey@g", want: "g"},
		"subdomains":    {input: "hey@subdomain.domain.com", want: "subdomain.domain.com"},
		"invalid email": {input: "hey", want: "subdomain.domain.com", err: invalidEmailErr},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := getDomain(tc.input)
			diff := cmp.Diff(tc.want, got)

			if err != tc.err {
				t.Fatalf(err.Error())
			}

			if err == nil && diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestUniqueDomains(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  []string
		err   error
	}{
		"simple":        {input: []string{"test@test.com", "other@test.com", "test@another.com"}, want: []string{"test.com", "another.com"}},
		"only one":      {input: []string{"test@test.com"}, want: []string{"test.com"}},
		"invalid email": {input: []string{"test.com"}, want: []string{"test.com"}, err: invalidEmailErr},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := uniqueDomains(tc.input)
			diff := cmp.Diff(tc.want, got)

			if err != tc.err {
				t.Fatalf(err.Error())
			}

			if err == nil && diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestGetOrderedCounts(t *testing.T) {
	tests := map[string]struct {
		input map[string]int
		want  []kv
	}{
		"simple":       {input: map[string]int{"google.com": 3, "yahoo.ca": 4}, want: []kv{{Key: "yahoo.ca", Value: 4}, {Key: "google.com", Value: 3}}},
		"removes zero": {input: map[string]int{"google.com": 3, "yahoo.ca": 0}, want: []kv{{Key: "google.com", Value: 3}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := getOrderedCounts(tc.input)
			diff := cmp.Diff(tc.want, got)

			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
