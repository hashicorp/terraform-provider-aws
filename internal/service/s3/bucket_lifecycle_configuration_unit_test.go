// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestLifecycleConfigEqual(t *testing.T) {
	t.Parallel()

	const (
		allStorageClasses128K = awstypes.TransitionDefaultMinimumObjectSizeAllStorageClasses128k
		variesByStorageClass  = awstypes.TransitionDefaultMinimumObjectSizeVariesByStorageClass
	)

	rulesA := []awstypes.LifecycleRule{{
		ID:     aws.String("rule-a"),
		Status: awstypes.ExpirationStatusEnabled,
	}}
	rulesB := []awstypes.LifecycleRule{{
		ID:     aws.String("rule-b"),
		Status: awstypes.ExpirationStatusEnabled,
	}}

	testCases := map[string]struct {
		transitionMinSize1 awstypes.TransitionDefaultMinimumObjectSize
		rules1             []awstypes.LifecycleRule
		transitionMinSize2 awstypes.TransitionDefaultMinimumObjectSize
		rules2             []awstypes.LifecycleRule
		want               bool
	}{
		// Both sides read back empty: an S3-compatible endpoint that doesn't send
		// the x-amz-transition-default-minimum-object-size header, comparing two
		// consecutive reads.
		"both empty, equal rules": {
			transitionMinSize1: "", rules1: rulesA,
			transitionMinSize2: "", rules2: rulesA,
			want: true,
		},
		// The reported bug: the value read back from an S3-compatible endpoint is
		// empty, but the provider defaults the desired value to all_storage_classes_128K.
		// The rules match, so the configuration has converged.
		"empty read-back vs default, equal rules": {
			transitionMinSize1: "", rules1: rulesA,
			transitionMinSize2: allStorageClasses128K, rules2: rulesA,
			want: true,
		},
		// Same as above with the arguments reversed (the comparison must be symmetric).
		"default vs empty read-back, equal rules": {
			transitionMinSize1: allStorageClasses128K, rules1: rulesA,
			transitionMinSize2: "", rules2: rulesA,
			want: true,
		},
		// AWS happy path: both sides report the same value.
		"both default, equal rules": {
			transitionMinSize1: allStorageClasses128K, rules1: rulesA,
			transitionMinSize2: allStorageClasses128K, rules2: rulesA,
			want: true,
		},
		"both default, different rules": {
			transitionMinSize1: allStorageClasses128K, rules1: rulesA,
			transitionMinSize2: allStorageClasses128K, rules2: rulesB,
			want: false,
		},
		// A genuine difference between two reported values (e.g. mid-propagation on
		// AWS) must not be treated as converged.
		"different reported values, equal rules": {
			transitionMinSize1: allStorageClasses128K, rules1: rulesA,
			transitionMinSize2: variesByStorageClass, rules2: rulesA,
			want: false,
		},
		// Even when the transition size is skipped, differing rules are still compared.
		"empty read-back, different rules": {
			transitionMinSize1: "", rules1: rulesA,
			transitionMinSize2: allStorageClasses128K, rules2: rulesB,
			want: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfs3.LifecycleConfigEqual(tc.transitionMinSize1, tc.rules1, tc.transitionMinSize2, tc.rules2); got != tc.want {
				t.Errorf("LifecycleConfigEqual() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestKeepTransitionDefaultMinimumObjectSize(t *testing.T) {
	t.Parallel()

	allStorageClasses128K := fwtypes.StringEnumValue(awstypes.TransitionDefaultMinimumObjectSizeAllStorageClasses128k)
	variesByStorageClass := fwtypes.StringEnumValue(awstypes.TransitionDefaultMinimumObjectSizeVariesByStorageClass)
	empty := fwtypes.StringEnumValue(awstypes.TransitionDefaultMinimumObjectSize(""))
	null := fwtypes.StringEnumNull[awstypes.TransitionDefaultMinimumObjectSize]()
	unknown := fwtypes.StringEnumUnknown[awstypes.TransitionDefaultMinimumObjectSize]()

	testCases := map[string]struct {
		keep     fwtypes.StringEnum[awstypes.TransitionDefaultMinimumObjectSize]
		readBack fwtypes.StringEnum[awstypes.TransitionDefaultMinimumObjectSize]
		want     fwtypes.StringEnum[awstypes.TransitionDefaultMinimumObjectSize]
	}{
		// The endpoint didn't report the value (empty), so the planned default is
		// preserved to keep the result consistent with the plan.
		"empty read-back keeps planned default": {
			keep: allStorageClasses128K, readBack: empty, want: allStorageClasses128K,
		},
		"null read-back keeps planned default": {
			keep: allStorageClasses128K, readBack: null, want: allStorageClasses128K,
		},
		// The endpoint reported a value: it is authoritative.
		"reported value is authoritative": {
			keep: allStorageClasses128K, readBack: variesByStorageClass, want: variesByStorageClass,
		},
		"equal reported value": {
			keep: allStorageClasses128K, readBack: allStorageClasses128K, want: allStorageClasses128K,
		},
		// A planned value can be unknown (e.g. an S3 Express directory bucket). An
		// unknown value can't be stored, so the (known) empty read-back is used.
		"unknown planned with empty read-back uses read-back": {
			keep: unknown, readBack: empty, want: empty,
		},
		"null planned with empty read-back keeps null": {
			keep: null, readBack: empty, want: null,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfs3.KeepTransitionDefaultMinimumObjectSize(tc.keep, tc.readBack); !got.Equal(tc.want) {
				t.Errorf("KeepTransitionDefaultMinimumObjectSize() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
