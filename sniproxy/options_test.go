package sniproxy

import (
	"net/url"
	"testing"
)

func TestDecodeOptions(t *testing.T) {
	for _, test := range []struct {
		s    string
		want Options
	}{
		{s: "", want: Options{}},
		{s: "{}", want: Options{}},
		{s: `{"Siding":true}`, want: Options{Siding: true}},
		{s: `{"DialWithAddr":true}`, want: Options{DialWithAddr: true}},
		{
			s:    `{"Siding":true,"DialWithAddr":true}`,
			want: Options{Siding: true, DialWithAddr: true},
		},
		// Unknown fields are ignored.
		{s: `{"Bogus":123}`, want: Options{}},
	} {
		got, err := decodeOptions(test.s)
		if err != nil {
			t.Errorf("decodeOptions(%q): unexpected error: %s", test.s, err)
			continue
		}
		if got == nil {
			t.Errorf("decodeOptions(%q): got nil options", test.s)
			continue
		}
		if *got != test.want {
			t.Errorf(
				"decodeOptions(%q) = %+v, want %+v",
				test.s, *got, test.want,
			)
		}
	}
}

func TestDecodeOptionsError(t *testing.T) {
	for _, s := range []string{
		"not json",
		"{",
		`{"Siding":"yes"}`, // wrong type for bool field
	} {
		got, err := decodeOptions(s)
		if err == nil {
			t.Errorf("decodeOptions(%q): got nil error, want error", s)
		}
		if got != nil {
			t.Errorf("decodeOptions(%q): got %+v, want nil on error", s, got)
		}
	}
}

func TestOptionsFromQuery(t *testing.T) {
	// No "opt" key yields default options.
	got, err := optionsFromQuery(url.Values{})
	if err != nil {
		t.Fatalf("optionsFromQuery(empty): %s", err)
	}
	if *got != (Options{}) {
		t.Errorf("optionsFromQuery(empty) = %+v, want zero", *got)
	}

	// The "opt" query value is decoded as JSON options.
	q := url.Values{}
	q.Set("opt", `{"Siding":true,"DialWithAddr":true}`)
	got, err = optionsFromQuery(q)
	if err != nil {
		t.Fatalf("optionsFromQuery: %s", err)
	}
	want := Options{Siding: true, DialWithAddr: true}
	if *got != want {
		t.Errorf("optionsFromQuery = %+v, want %+v", *got, want)
	}

	// An invalid "opt" value surfaces the decode error.
	q.Set("opt", "not json")
	if _, err := optionsFromQuery(q); err == nil {
		t.Errorf("optionsFromQuery(invalid): got nil error, want error")
	}
}
