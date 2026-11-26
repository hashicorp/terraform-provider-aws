// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func TestValidateEventHoldDuration(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		hasDuration bool
		configured  string
		state       attrState
		effective   string
		wantErr     bool
	}{
		"no duration is always fine": {
			hasDuration: false,
			state:       attrAbsent,
			effective:   "",
		},
		"no duration with a released hold": {
			hasDuration: false,
			state:       attrAbsent,
			effective:   "OFF",
		},
		"duration with a configured ON": {
			hasDuration: true,
			configured:  "ON",
			state:       attrSet,
		},
		"duration with a configured OFF": {
			hasDuration: true,
			configured:  "OFF",
			state:       attrSet,
			wantErr:     true,
		},
		// A hold supplied by another resource is not resolved during the first
		// plan. Rejecting it there would fail a configuration that is valid.
		"duration with an unknown hold defers": {
			hasDuration: true,
			state:       attrUnknown,
		},
		"duration with an unknown hold defers even with state": {
			hasDuration: true,
			state:       attrUnknown,
			effective:   "OFF",
		},
		"duration with no hold anywhere": {
			hasDuration: true,
			state:       attrAbsent,
			effective:   "",
			wantErr:     true,
		},
		// Inherited from a bucket default_event_hold.
		"duration with an inherited ON": {
			hasDuration: true,
			state:       attrAbsent,
			effective:   "ON",
		},
		// The hold was released, so the duration would be dropped silently.
		"duration with a released hold in state": {
			hasDuration: true,
			state:       attrAbsent,
			effective:   "OFF",
			wantErr:     true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := validateEventHoldDuration(testCase.hasDuration, testCase.configured, testCase.state, testCase.effective)

			if got, want := err != nil, testCase.wantErr; got != want {
				t.Fatalf("error %t, want %t: %v", got, want, err)
			}
		})
	}
}

type mockEventHoldDiffer struct {
	rawConfig cty.Value
	values    map[string]any
}

var _ sdkv2.ResourceDiffer = (*mockEventHoldDiffer)(nil)

func (d *mockEventHoldDiffer) Get(key string) any {
	if v, ok := d.values[key]; ok {
		return v
	}
	return ""
}

func (d *mockEventHoldDiffer) GetChange(string) (any, any) { return nil, nil }

func (d *mockEventHoldDiffer) GetOk(key string) (any, bool) {
	v, ok := d.values[key]
	if !ok {
		return nil, false
	}
	switch typed := v.(type) {
	case int:
		return typed, typed != 0
	case string:
		return typed, typed != ""
	default:
		return v, v != nil
	}
}

func (d *mockEventHoldDiffer) GetRawConfig() cty.Value {
	if d.rawConfig.Type() == cty.NilType {
		return cty.NullVal(cty.DynamicPseudoType)
	}
	return d.rawConfig
}

func (d *mockEventHoldDiffer) GetRawPlan() cty.Value  { return cty.NullVal(cty.DynamicPseudoType) }
func (d *mockEventHoldDiffer) GetRawState() cty.Value { return cty.NullVal(cty.DynamicPseudoType) }
func (d *mockEventHoldDiffer) HasChange(string) bool  { return false }
func (d *mockEventHoldDiffer) HasChanges(...string) bool {
	return false
}
func (d *mockEventHoldDiffer) Id() string { return "test-object" }

func eventHoldRawConfig(values map[string]cty.Value) cty.Value {
	allValues := map[string]cty.Value{
		"object_lock_event_hold":                cty.NullVal(cty.String),
		"object_lock_event_hold_duration_days":  cty.NullVal(cty.Number),
		"object_lock_event_hold_duration_years": cty.NullVal(cty.Number),
	}

	maps.Copy(allValues, values)

	return cty.ObjectVal(allValues)
}

func TestValidateObjectEventHold(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rawConfig cty.Value
		effective string
		wantErr   bool
	}{
		"ignore_changes all with unset hold": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal(""),
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(0),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(0),
			}),
			effective: "",
		},
		"ignore_changes all with released hold": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal("OFF"),
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(0),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(0),
			}),
			effective: "OFF",
		},
		"ignore_changes all with active days hold": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal("ON"),
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(30),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(0),
			}),
			effective: "ON",
		},
		"ignore_changes all with active years hold": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal("ON"),
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(0),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(2),
			}),
			effective: "ON",
		},
		"configured days without hold in config or state": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold_duration_days": cty.NumberIntVal(30),
			}),
			effective: "",
			wantErr:   true,
		},
		"configured years with empty hold in config": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal(""),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(1),
			}),
			effective: "",
			wantErr:   true,
		},
		"unknown duration defers validation": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal(""),
				"object_lock_event_hold_duration_days":  cty.UnknownVal(cty.Number),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(0),
			}),
			effective: "",
		},
		"null raw config": {
			rawConfig: cty.NullVal(cty.DynamicPseudoType),
			effective: "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := &mockEventHoldDiffer{
				rawConfig: testCase.rawConfig,
				values: map[string]any{
					"object_lock_event_hold": testCase.effective,
				},
			}

			err := validateObjectEventHold(d)

			if got, want := err != nil, testCase.wantErr; got != want {
				t.Fatalf("error %t, want %t: %v", got, want, err)
			}
		})
	}
}

func TestExpandObjectEventHoldDuration(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rawConfig cty.Value
		values    map[string]any
		want      *types.EventHoldDuration
	}{
		"ignore_changes all with zero durations": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal(""),
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(0),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(0),
			}),
			values: map[string]any{
				"object_lock_event_hold_duration_days":  0,
				"object_lock_event_hold_duration_years": 0,
			},
			want: nil,
		},
		"ignore_changes all with zero days and active years": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":                cty.StringVal("ON"),
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(0),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(2),
			}),
			values: map[string]any{
				"object_lock_event_hold_duration_days":  0,
				"object_lock_event_hold_duration_years": 2,
			},
			want: &types.EventHoldDuration{
				Years: aws.Int32(2),
			},
		},
		"configured days": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold":               cty.StringVal("ON"),
				"object_lock_event_hold_duration_days": cty.NumberIntVal(30),
			}),
			want: &types.EventHoldDuration{
				Days: aws.Int32(30),
			},
		},
		"falls back to state days when raw config unset": {
			rawConfig: eventHoldRawConfig(nil),
			values: map[string]any{
				"object_lock_event_hold_duration_days": 45,
			},
			want: &types.EventHoldDuration{
				Days: aws.Int32(45),
			},
		},
		"falls back to state years when raw config has zeroes": {
			rawConfig: eventHoldRawConfig(map[string]cty.Value{
				"object_lock_event_hold_duration_days":  cty.NumberIntVal(0),
				"object_lock_event_hold_duration_years": cty.NumberIntVal(0),
			}),
			values: map[string]any{
				"object_lock_event_hold_duration_years": 3,
			},
			want: &types.EventHoldDuration{
				Years: aws.Int32(3),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := &mockEventHoldDiffer{
				rawConfig: testCase.rawConfig,
				values:    testCase.values,
			}

			got := expandObjectEventHoldDuration(d)

			if diff := cmp.Diff(got, testCase.want, cmpopts.IgnoreUnexported(types.EventHoldDuration{})); diff != "" {
				t.Fatalf("unexpected duration diff (-got +want):\n%s", diff)
			}
		})
	}
}
