// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listresource

import (
	"testing"
)

type mockResourceData struct {
	id     string
	values map[string]any
}

func (d mockResourceData) Id() string {
	return d.id
}

func (d mockResourceData) GetOk(name string) (any, bool) {
	v, ok := d.values[name]
	if !ok {
		return nil, false
	}
	// Mimic helper/schema.ResourceData.GetOk, which reports zero values as unset.
	switch tv := v.(type) {
	case string:
		return v, tv != ""
	case bool:
		return v, tv
	case int:
		return v, tv != 0
	default:
		return v, true
	}
}

func TestGetAttributeOk(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		d         mockResourceData
		name      string
		wantValue any
		wantOk    bool
	}{
		"id": {
			d:         mockResourceData{id: "some-id"},
			name:      "id",
			wantValue: "some-id",
			wantOk:    true,
		},
		"set string": {
			d:         mockResourceData{values: map[string]any{"name": "value"}},
			name:      "name",
			wantValue: "value",
			wantOk:    true,
		},
		"unset string": {
			d:         mockResourceData{values: map[string]any{"name": ""}},
			name:      "name",
			wantValue: nil,
			wantOk:    false,
		},
		"zero value bool is valid": {
			d:         mockResourceData{values: map[string]any{"egress": false}},
			name:      "egress",
			wantValue: false,
			wantOk:    true,
		},
		"true bool": {
			d:         mockResourceData{values: map[string]any{"egress": true}},
			name:      "egress",
			wantValue: true,
			wantOk:    true,
		},
		"set int": {
			d:         mockResourceData{values: map[string]any{"rule_number": 100}},
			name:      "rule_number",
			wantValue: 100,
			wantOk:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotValue, gotOk := getAttributeOk(tc.d, tc.name)
			if gotOk != tc.wantOk {
				t.Errorf("getAttributeOk(%q) ok = %v, want %v", tc.name, gotOk, tc.wantOk)
			}
			if gotValue != tc.wantValue {
				t.Errorf("getAttributeOk(%q) value = %v, want %v", tc.name, gotValue, tc.wantValue)
			}
		})
	}
}
