// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import (
	"context"
	"testing"
)

func TestCommitmentStringSemanticEquals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		oldValue string
		newValue string
		want     bool
	}{
		{
			name:     "treats equivalent decimals as equal",
			oldValue: "1.158",
			newValue: "1.15800000",
			want:     true,
		},
		{
			name:     "treats different decimals as unequal",
			oldValue: "1.158",
			newValue: "1.159",
			want:     false,
		},
		{
			name:     "treats values with different representation as equal",
			oldValue: "0001.2300",
			newValue: "1.23",
			want:     true,
		},
		{
			name:     "falls back to exact string equality for non-decimal values",
			oldValue: "abc",
			newValue: "abc",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			oldValue := CommitmentStringValue(tt.oldValue)
			got, diags := oldValue.StringSemanticEquals(context.Background(), CommitmentStringValue(tt.newValue))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != tt.want {
				t.Errorf("StringSemanticEquals(%q, %q) = %t, want %t", tt.oldValue, tt.newValue, got, tt.want)
			}
		})
	}
}
