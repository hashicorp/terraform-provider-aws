// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository/types"
	"github.com/google/go-cmp/cmp"
)

func TestFlattenNonDefaultCloudFormationParameters(t *testing.T) {
	for _, tc := range []struct {
		name             string
		cfParams         []cloudformationtypes.Parameter
		parameterDefs    []awstypes.ParameterDefinition
		configuredParams map[string]any
		expected         map[string]any
	}{
		{
			name: "non-default value included",
			cfParams: []cloudformationtypes.Parameter{
				{ParameterKey: aws.String("Param1"), ParameterValue: aws.String("custom")},
			},
			parameterDefs: []awstypes.ParameterDefinition{
				{Name: aws.String("Param1"), DefaultValue: aws.String("default")},
			},
			configuredParams: map[string]any{},
			expected:         map[string]any{"Param1": "custom"},
		},
		{
			name: "NoEcho preserves configured value",
			cfParams: []cloudformationtypes.Parameter{
				{ParameterKey: aws.String("Secret"), ParameterValue: aws.String("****")},
			},
			parameterDefs: []awstypes.ParameterDefinition{
				{Name: aws.String("Secret"), NoEcho: aws.Bool(true)},
			},
			configuredParams: map[string]any{"Secret": "my-secret"},
			expected:         map[string]any{"Secret": "my-secret"},
		},
		{
			name: "NoEcho not in config is excluded",
			cfParams: []cloudformationtypes.Parameter{
				{ParameterKey: aws.String("Secret"), ParameterValue: aws.String("****")},
			},
			parameterDefs: []awstypes.ParameterDefinition{
				{Name: aws.String("Secret"), NoEcho: aws.Bool(true)},
			},
			configuredParams: map[string]any{},
			expected:         map[string]any{},
		},
		{
			name: "configured param matching default is included",
			cfParams: []cloudformationtypes.Parameter{
				{ParameterKey: aws.String("Param1"), ParameterValue: aws.String("default")},
			},
			parameterDefs: []awstypes.ParameterDefinition{
				{Name: aws.String("Param1"), DefaultValue: aws.String("default")},
			},
			configuredParams: map[string]any{"Param1": "default"},
			expected:         map[string]any{"Param1": "default"},
		},
		{
			name: "unconfigured param matching default is excluded",
			cfParams: []cloudformationtypes.Parameter{
				{ParameterKey: aws.String("Param1"), ParameterValue: aws.String("default")},
			},
			parameterDefs: []awstypes.ParameterDefinition{
				{Name: aws.String("Param1"), DefaultValue: aws.String("default")},
			},
			configuredParams: map[string]any{},
			expected:         map[string]any{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenNonDefaultCloudFormationParameters(tc.cfParams, tc.parameterDefs, tc.configuredParams)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
