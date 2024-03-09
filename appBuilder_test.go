package main

import (
	"testing"
)

func TestAppBuilder_cleanResponse(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"```hcl\n\nsome code\n\n```", "some code"},
	}

	for _, test := range tests {
		if output := cleanResponse(test.input); output != test.expected {
			t.Errorf("Test failed: input: %s, expected: %s, received: %s", test.input, test.expected, output)
		}
	}
}
