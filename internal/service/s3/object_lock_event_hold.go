// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Shared by aws_s3_object and aws_s3_object_copy.

// attrState reports how configuration supplies an attribute.
type attrState int

const (
	attrAbsent  attrState = iota // not written in configuration
	attrUnknown                  // written, but not resolved during planning
	attrSet                      // written, with a known value
)

// configuredString reports a top-level string attribute as written in
// configuration. d.Get falls back to prior state for Optional+Computed, so it
// cannot tell a configured value from one the provider reported earlier.
func configuredString(raw cty.Value, key string) (string, attrState) {
	if raw.IsNull() || !raw.IsKnown() {
		return "", attrAbsent
	}

	switch v := raw.GetAttr(key); {
	case v.IsNull():
		return "", attrAbsent
	case !v.IsKnown():
		return "", attrUnknown
	default:
		return v.AsString(), attrSet
	}
}

// configuredInt32 is configuredString for a top-level integer attribute.
func configuredInt32(raw cty.Value, key string) (int32, attrState) {
	if raw.IsNull() || !raw.IsKnown() {
		return 0, attrAbsent
	}

	switch v := raw.GetAttr(key); {
	case v.IsNull():
		return 0, attrAbsent
	case !v.IsKnown():
		return 0, attrUnknown
	default:
		i, _ := v.AsBigFloat().Int64()
		return int32(i), attrSet
	}
}

// expandObjectEventHoldDuration prefers whichever unit configuration names; a
// days/years switch leaves the old one in state, where d.GetOk would find it.
// With neither configured it falls back to state, preserving an inherited
// duration - S3 rejects EventHold:ON without one.
func expandObjectEventHoldDuration(d *schema.ResourceData) *types.EventHoldDuration {
	raw := d.GetRawConfig()

	if v, state := configuredInt32(raw, "object_lock_event_hold_duration_days"); state == attrSet {
		return &types.EventHoldDuration{Days: aws.Int32(v)}
	}
	if v, state := configuredInt32(raw, "object_lock_event_hold_duration_years"); state == attrSet {
		return &types.EventHoldDuration{Years: aws.Int32(v)}
	}

	if v, ok := d.GetOk("object_lock_event_hold_duration_days"); ok {
		return &types.EventHoldDuration{Days: aws.Int32(int32(v.(int)))}
	}
	if v, ok := d.GetOk("object_lock_event_hold_duration_years"); ok {
		return &types.EventHoldDuration{Years: aws.Int32(int32(v.(int)))}
	}

	return nil
}

func validateObjectEventHold(d *schema.ResourceDiff) error {
	raw := d.GetRawConfig()

	_, days := configuredInt32(raw, "object_lock_event_hold_duration_days")
	_, years := configuredInt32(raw, "object_lock_event_hold_duration_years")
	if days == attrUnknown || years == attrUnknown {
		// Nothing to check against until the duration itself resolves.
		return nil
	}

	hold, state := configuredString(raw, "object_lock_event_hold")

	return validateEventHoldDuration(days == attrSet || years == attrSet, hold, state, d.Get("object_lock_event_hold").(string))
}

// validateEventHoldDuration rejects a duration the provider would not send. A
// duration is only ever forwarded while the effective hold is ON, so any other
// state would accept configuration and silently drop it.
func validateEventHoldDuration(hasDuration bool, configured string, state attrState, effective string) error {
	const durationAttrs = "object_lock_event_hold_duration_days and object_lock_event_hold_duration_years"

	if !hasDuration {
		return nil
	}

	switch state {
	case attrUnknown:
		// Not resolved yet; Terraform plans again once it is.
		return nil
	case attrSet:
		if types.ObjectLockEventHold(configured) != types.ObjectLockEventHoldOn {
			return fmt.Errorf("%s cannot be set when object_lock_event_hold is %q", durationAttrs, configured)
		}
		return nil
	}

	// Not configured, so fall back to the hold already in effect. A released
	// hold still reads as OFF and must not pass.
	if types.ObjectLockEventHold(effective) == types.ObjectLockEventHoldOn {
		return nil
	}
	if effective == "" {
		return fmt.Errorf("%s require object_lock_event_hold to be ON", durationAttrs)
	}

	return fmt.Errorf("%s require object_lock_event_hold to be ON, but it is %q", durationAttrs, effective)
}

// suppressEventHoldRetainUntilDrift ignores the retain-until date moving ahead
// of the configured value. S3 reports max(configured, now+duration) under a
// hold and advances it, so a configured date is a minimum, not an exact value.
func suppressEventHoldRetainUntilDrift(_, old, new string, d *schema.ResourceData) bool {
	// Applies after release too: S3 freezes the date at release time plus the
	// duration, which can stay later than the configured minimum.
	if d.Get("object_lock_event_hold").(string) == "" {
		return false
	}
	if old == "" || new == "" {
		return false
	}

	o, err := time.Parse(time.RFC3339, old)
	if err != nil {
		return false
	}
	n, err := time.Parse(time.RFC3339, new)
	if err != nil {
		return false
	}

	return !o.Before(n)
}
