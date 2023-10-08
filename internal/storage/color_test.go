package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseHexColor(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr error
		want    HexColor
	}{
		{
			"normal usage",
			"#fe9303",
			nil,
			"#fe9303",
		},
		{
			"case is irrelevant",
			"#FE9303",
			nil,
			"#FE9303",
		},
		{
			"3-character is accepted the same as 6-character",
			"#eee",
			nil,
			"#eee",
		},
		{
			"hash character must be included",
			"fe9303",
			ErrInvalidHexColor,
			"#cccccc",
		},
		{
			"named color values are not recognized",
			"aliceblue",
			ErrInvalidHexColor,
			"#cccccc",
		},
		{
			"non-hex string is invalid color",
			"zzddhh",
			ErrInvalidHexColor,
			"#cccccc",
		},
		{
			"empty string is invalid color",
			"",
			ErrInvalidHexColor,
			"#cccccc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHexColor(tt.s)
			assert.Equal(t, tt.want, got)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
