package helpers_test

import (
	"testing"

	"lnk/domain/entities/helpers"
)

func Test_Helper_Base62Encode(t *testing.T) {
	t.Parallel()

	salt := "salt"
	tests := []struct {
		want string
		id   int64
	}{
		{"wwwE", 1},
		{"wwwf", 2},
		{"www2", 3},
		{"wwwQ", 4},
		{"wwwS", 5},
		{"wwwV", 6},
		{"wwwa", 7},
		{"wwwm", 8},
		{"wwwD", 9},
	}

	for _, test := range tests {
		got := helpers.Base62Encode(test.id, salt)
		if got != test.want {
			t.Errorf("Base62Encode(%d, %s) = %s, want %s", test.id, salt, got, test.want)
		}
	}
}
