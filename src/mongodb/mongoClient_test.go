package mongodb

import "testing"

func Test_dbConnect(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"connect"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConnect()
		})
	}
}
