// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

// TestLifecycleConfigEqual exercises the transition_default_minimum_object_size
// comparison, which must tolerate an empty *observed* value so that S3-compatible
// backends that omit the x-amz-transition-default-minimum-object-size field (e.g.
// Ceph RADOS Gateway) can still reach convergence in waitLifecycleConfigEquals.
//
// The first argument is the observed (backend GET) value; the second is the
// requested value. The empty-tolerance is asymmetric by design: an empty
// observed value matches any requested value, but a non-empty observed value
// must match exactly. Callers must pass the observed value first.
func TestLifecycleConfigEqual(t *testing.T) {
	t.Parallel()

	rules := []awstypes.LifecycleRule{
		{
			ID:     aws.String("rule-1"),
			Status: awstypes.ExpirationStatusEnabled,
			Filter: &awstypes.LifecycleRuleFilter{Prefix: aws.String("/")},
			NoncurrentVersionExpiration: &awstypes.NoncurrentVersionExpiration{
				NoncurrentDays: aws.Int32(1),
			},
		},
	}
	differentRules := []awstypes.LifecycleRule{
		{
			ID:     aws.String("rule-1"),
			Status: awstypes.ExpirationStatusEnabled,
			Filter: &awstypes.LifecycleRuleFilter{Prefix: aws.String("/")},
			NoncurrentVersionExpiration: &awstypes.NoncurrentVersionExpiration{
				NoncurrentDays: aws.Int32(5),
			},
		},
	}

	const (
		size128k awstypes.TransitionDefaultMinimumObjectSize = awstypes.TransitionDefaultMinimumObjectSizeAllStorageClasses128k
		sizeVary awstypes.TransitionDefaultMinimumObjectSize = awstypes.TransitionDefaultMinimumObjectSizeVariesByStorageClass
		empty    awstypes.TransitionDefaultMinimumObjectSize = ""
	)

	testCases := map[string]struct {
		observedSize awstypes.TransitionDefaultMinimumObjectSize
		observedRule []awstypes.LifecycleRule
		requestSize  awstypes.TransitionDefaultMinimumObjectSize
		requestRule  []awstypes.LifecycleRule
		want         bool
	}{
		// The Ceph case: the backend omits the field, so the observed value is
		// empty. It must be treated as equal to any requested value, otherwise the
		// convergence wait loops forever. This is the whole point of the guard.
		"observed empty, requested non-empty, rules equal": {
			observedSize: empty,
			observedRule: rules,
			requestSize:  size128k,
			requestRule:  rules,
			want:         true,
		},
		"observed empty, requested empty, rules equal": {
			observedSize: empty,
			observedRule: rules,
			requestSize:  empty,
			requestRule:  rules,
			want:         true,
		},
		// The AWS regression guard: a non-empty observed value that differs from
		// the requested value must NOT be treated as equal, so genuine mismatches
		// on real AWS are still detected.
		"observed non-empty differs from requested": {
			observedSize: size128k,
			observedRule: rules,
			requestSize:  sizeVary,
			requestRule:  rules,
			want:         false,
		},
		"observed non-empty equals requested": {
			observedSize: size128k,
			observedRule: rules,
			requestSize:  size128k,
			requestRule:  rules,
			want:         true,
		},
		// Rule differences are still detected regardless of the size handling.
		"observed empty, rules differ": {
			observedSize: empty,
			observedRule: rules,
			requestSize:  size128k,
			requestRule:  differentRules,
			want:         false,
		},
		"sizes equal, rules differ": {
			observedSize: size128k,
			observedRule: rules,
			requestSize:  size128k,
			requestRule:  differentRules,
			want:         false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfs3.LifecycleConfigEqual(tc.observedSize, tc.observedRule, tc.requestSize, tc.requestRule)
			if got != tc.want {
				t.Errorf("LifecycleConfigEqual(%q, _, %q, _) = %t, want %t", tc.observedSize, tc.requestSize, got, tc.want)
			}
		})
	}
}
