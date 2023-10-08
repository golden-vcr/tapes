package sheets

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseTagHeading(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"Instructional?", "instructional"},
		{"Arts + Crafts?", "arts+crafts"},
		{"Runtime", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q yields %q", tt.s, tt.want), func(t *testing.T) {
			got := parseTagHeading(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}
