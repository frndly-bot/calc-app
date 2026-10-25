package main

import "testing"

func TestServerPort(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{name: "default", value: "", want: 34115},
		{name: "override", value: "34116", want: 34116},
		{name: "zero falls back", value: "0", want: 34115},
		{name: "out of range falls back", value: "70000", want: 34115},
		{name: "invalid falls back", value: "abc", want: 34115},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("CALC_PORT", test.value)
			if got := serverPort(); got != test.want {
				t.Fatalf("serverPort() = %d, want %d", got, test.want)
			}
		})
	}
}
