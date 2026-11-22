// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
)

func embeddedSettingsKey(m map[string]any) []any {
	v, _ := m["embedded_destination_settings"].([]any)
	return v
}

func TestFlattenCaptionDestinationSettingsEmbedded(t *testing.T) {
	// Non-nil embedded settings round-trips as a single empty element.
	in := &types.CaptionDestinationSettings{
		EmbeddedDestinationSettings: &types.EmbeddedDestinationSettings{},
	}
	out := flattenCaptionDescriptionsCaptionDestinationSettings(in)
	if len(out) != 1 {
		t.Fatalf("expected one destination settings map, got %d", len(out))
	}
	m, ok := out[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", out[0])
	}
	got := embeddedSettingsKey(m)
	if len(got) != 1 {
		t.Fatalf("expected embedded_destination_settings length 1, got %d", len(got))
	}

	// Nil embedded settings emit an absent (empty) list.
	inNil := &types.CaptionDestinationSettings{}
	outNil := flattenCaptionDescriptionsCaptionDestinationSettings(inNil)
	mNil, _ := outNil[0].(map[string]any)
	if gotNil := embeddedSettingsKey(mNil); len(gotNil) != 0 {
		t.Fatalf("expected embedded_destination_settings length 0, got %d", len(gotNil))
	}

	// Fully nil input returns nil.
	if got := flattenCaptionDescriptionsCaptionDestinationSettings(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %v", got)
	}
}

func TestFlattenExpandCaptionDestinationSettingsEmbeddedRoundTrip(t *testing.T) {
	// Non-nil embedded settings survive a flatten/expand round trip.
	in := &types.CaptionDestinationSettings{
		EmbeddedDestinationSettings: &types.EmbeddedDestinationSettings{},
	}
	expanded := expandChannelEncoderSettingsCaptionDescriptionsDestinationSettings(
		flattenCaptionDescriptionsCaptionDestinationSettings(in),
	)
	if expanded == nil || expanded.EmbeddedDestinationSettings == nil {
		t.Fatalf("expected EmbeddedDestinationSettings to survive round trip, got %v", expanded)
	}

	// Nil embedded settings stay nil after a round trip.
	inNil := &types.CaptionDestinationSettings{}
	expandedNil := expandChannelEncoderSettingsCaptionDescriptionsDestinationSettings(
		flattenCaptionDescriptionsCaptionDestinationSettings(inNil),
	)
	if expandedNil != nil && expandedNil.EmbeddedDestinationSettings != nil {
		t.Fatalf("expected EmbeddedDestinationSettings to stay nil, got %v", expandedNil)
	}
}
