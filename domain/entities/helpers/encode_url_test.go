package helpers

import (
	"testing"
)

func Test_Helper_Base62Encode(t *testing.T) {
	t.Parallel()

	salt := "salt"
	tests := []struct {
		id   int64
		want string
	}{
		{1, "wwwE"},
		{2, "wwwf"},
		{3, "www2"},
		{4, "wwwQ"},
		{5, "wwwS"},
		{6, "wwwV"},
		{7, "wwwa"},
		{8, "wwwm"},
		{9, "wwwD"},
	}

	for _, test := range tests {
		got := Base62Encode(test.id, salt)
		if got != test.want {
			t.Errorf("Base62Encode(%d, %s) = %s, want %s", test.id, salt, got, test.want)
		}
	}
}
