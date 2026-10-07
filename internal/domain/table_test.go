package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestTableLabelValidate(t *testing.T) {
	cases := []struct {
		name    string
		table   Table
		wantErr error
	}{
		{
			name:    "plain label",
			table:   Table{ID: "t1", Label: "T1", Status: TableFree},
			wantErr: nil,
		},
		{
			name:    "padded label is trimmed",
			table:   Table{ID: "t2", Label: "  T2  ", Status: TableFree},
			wantErr: nil,
		},
		{
			name:    "forty chars fit",
			table:   Table{ID: "t3", Label: strings.Repeat("m", 40), Status: TableFree},
			wantErr: nil,
		},
		{
			name:    "empty label",
			table:   Table{ID: "t4", Label: "", Status: TableFree},
			wantErr: ErrEmptyTableLabel,
		},
		{
			name:    "blank label",
			table:   Table{ID: "t5", Label: "   ", Status: TableFree},
			wantErr: ErrEmptyTableLabel,
		},
		{
			name:    "forty one chars overflow",
			table:   Table{ID: "t6", Label: strings.Repeat("m", 41), Status: TableFree},
			wantErr: ErrTableLabelTooLong,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.table.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want %v", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want error wrapping %v", err, tt.wantErr)
			}
		})
	}
}
