package distance

import "testing"

func TestLocationUsableForImpossibleTravel(t *testing.T) {
	tests := []struct {
		loc  string
		want bool
	}{
		{"", false},
		{"   ", false},
		{"UnknownCity", false},
		{"istanbul", true},
		{" Istanbul ", true},
	}
	for _, tt := range tests {
		if got := LocationUsableForImpossibleTravel(tt.loc); got != tt.want {
			t.Errorf("LocationUsableForImpossibleTravel(%q) = %v, want %v", tt.loc, got, tt.want)
		}
	}
}
