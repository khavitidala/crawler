package main

import (
	"reflect"
	"testing"
)

func TestGetURLsFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<a href="/path/one">
					<span>Boot.dev</span>
				</a>
				<a href="https://other.com/path/one">
					<span>Boot.dev</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:     "have no URL",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<img href="/path/one">
			</body>
		</html>
		`,
			expected: []string{},
		},
		{
			name:     "absolute URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<div class="a">
					<a href="https://other.com/path/one">A</a>
				</div>
			</body>
		</html>
		`,
			expected: []string{"https://other.com/path/one"},
		},
		{
			name:     "relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<div class="a">
					<a href="/path/one">A</a>
				</div>
			</body>
		</html>
		`,
			expected: []string{"https://blog.boot.dev/path/one"},
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL); err != nil {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
			} else if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
