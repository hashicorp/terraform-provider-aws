// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"testing"
)

func TestIPSetAddressesEquivalent(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		old  []any
		new  []any
		want bool
	}{
		"both empty": {
			old:  []any{},
			new:  []any{},
			want: true,
		},
		"identical": {
			old:  []any{"10.0.0.0/8", "192.0.2.0/24"},
			new:  []any{"192.0.2.0/24", "10.0.0.0/8"},
			want: true,
		},
		"equivalent IPv6 notation": {
			old:  []any{"2001:db8::/32"},
			new:  []any{"2001:0DB8:0000:0000:0000:0000:0000:0000/32"},
			want: true,
		},
		"different length": {
			old:  []any{"10.0.0.0/8"},
			new:  []any{"10.0.0.0/8", "192.0.2.0/24"},
			want: false,
		},
		"same length, one differs": {
			old:  []any{"10.0.0.0/8", "192.0.2.0/24"},
			new:  []any{"10.0.0.0/8", "198.51.100.0/24"},
			want: false,
		},
		"host bits differ": {
			old:  []any{"10.0.0.1/8"},
			new:  []any{"10.0.0.0/8"},
			want: false,
		},
		"unparsable old": {
			old:  []any{"not-a-cidr"},
			new:  []any{"not-a-cidr"},
			want: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := ipSetAddressesEquivalent(testCase.old, testCase.new); got != testCase.want {
				t.Errorf("ipSetAddressesEquivalent() = %t, want %t", got, testCase.want)
			}
		})
	}
}
