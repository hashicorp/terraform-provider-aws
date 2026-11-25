// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestApplicationStatusCheckAssociationImportIDParse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		id      string
		wantID  string
		want    map[string]any
		wantErr bool
	}{
		"instance": {
			id:     "asc-1234567890abcdef0,instance-id,i-0123456789abcdef0",
			wantID: "asc-1234567890abcdef0,instance-id,i-0123456789abcdef0",
			want: map[string]any{
				"application_status_check_id": "asc-1234567890abcdef0",
				names.AttrInstanceID:          "i-0123456789abcdef0",
			},
		},
		"tag with reserved characters": {
			id:     "asc-1234567890abcdef0,tag,team%2Cname,platform%3Dcore+services%2B",
			wantID: "asc-1234567890abcdef0,tag,team%2Cname,platform%3Dcore+services%2B",
			want: map[string]any{
				"application_status_check_id": "asc-1234567890abcdef0",
				"target_tag_key":              "team,name",
				"target_tag_value":            "platform=core services+",
			},
		},
		"empty tag value": {
			id:     "asc-1234567890abcdef0,tag,Environment,",
			wantID: "asc-1234567890abcdef0,tag,Environment,",
			want: map[string]any{
				"application_status_check_id": "asc-1234567890abcdef0",
				"target_tag_key":              "Environment",
				"target_tag_value":            "",
			},
		},
		"missing type": {
			id:      "asc-1234567890abcdef0",
			wantErr: true,
		},
		"unknown type": {
			id:      "asc-1234567890abcdef0,unknown,value",
			wantErr: true,
		},
		"instance extra part": {
			id:      "asc-1234567890abcdef0,instance-id,i-0123456789abcdef0,extra",
			wantErr: true,
		},
		"empty instance": {
			id:      "asc-1234567890abcdef0,instance-id,",
			wantErr: true,
		},
		"empty tag key": {
			id:      "asc-1234567890abcdef0,tag,,value",
			wantErr: true,
		},
		"invalid encoding": {
			id:      "asc-1234567890abcdef0,tag,%ZZ,value",
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotID, got, err := (applicationStatusCheckAssociationImportID{}).Parse(testCase.id)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected error, got ID %q and attributes %#v", gotID, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if gotID != testCase.wantID {
				t.Errorf("ID = %q, want %q", gotID, testCase.wantID)
			}
			if len(got) != len(testCase.want) {
				t.Fatalf("attributes = %#v, want %#v", got, testCase.want)
			}
			for key, want := range testCase.want {
				if got[key] != want {
					t.Errorf("attribute %q = %#v, want %#v", key, got[key], want)
				}
			}
		})
	}
}

func TestExpandApplicationStatusCheckAssociationTarget(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		model        applicationStatusCheckAssociationResourceModel
		wantInstance string
		wantTagKey   string
		wantTagValue string
		wantErr      bool
	}{
		"instance": {
			model: applicationStatusCheckAssociationResourceModel{
				InstanceID:     types.StringValue("i-0123456789abcdef0"),
				TargetTagKey:   types.StringNull(),
				TargetTagValue: types.StringNull(),
			},
			wantInstance: "i-0123456789abcdef0",
		},
		"tag": {
			model: applicationStatusCheckAssociationResourceModel{
				InstanceID:     types.StringNull(),
				TargetTagKey:   types.StringValue("Environment"),
				TargetTagValue: types.StringValue(""),
			},
			wantTagKey: "Environment",
		},
		"no target": {
			model: applicationStatusCheckAssociationResourceModel{
				InstanceID:     types.StringNull(),
				TargetTagKey:   types.StringNull(),
				TargetTagValue: types.StringNull(),
			},
			wantErr: true,
		},
		"multiple targets": {
			model: applicationStatusCheckAssociationResourceModel{
				InstanceID:     types.StringValue("i-0123456789abcdef0"),
				TargetTagKey:   types.StringValue("Environment"),
				TargetTagValue: types.StringValue("production"),
			},
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			instances, tags, err := expandApplicationStatusCheckAssociationTarget(testCase.model)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected error, got instances %#v and tags %#v", instances, tags)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if testCase.wantInstance != "" {
				if len(instances) != 1 || instances[0] != testCase.wantInstance || len(tags) != 0 {
					t.Fatalf("instances = %#v, tags = %#v", instances, tags)
				}
			} else if len(instances) != 0 || len(tags) != 1 || aws.ToString(tags[0].Key) != testCase.wantTagKey || aws.ToString(tags[0].Value) != testCase.wantTagValue {
				t.Fatalf("instances = %#v, tags = %#v", instances, tags)
			}
		})
	}
}

func TestApplicationStatusCheckAssociationResultsError(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		err := applicationStatusCheckAssociationResultsError(
			[]awstypes.SuccessfulAssociationResponseObject{{AssociationValue: aws.String("i-0123456789abcdef0")}},
			nil,
		)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		err := applicationStatusCheckAssociationResultsError(nil, nil)
		if err == nil || !strings.Contains(err.Error(), "expected one successful result") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("partial failures", func(t *testing.T) {
		t.Parallel()

		err := applicationStatusCheckAssociationResultsError(
			[]awstypes.SuccessfulAssociationResponseObject{{AssociationValue: aws.String("i-success")}},
			[]awstypes.UnsuccessfulAssociationResponseObject{
				{AssociationType: aws.String("INSTANCE_ID"), AssociationValue: aws.String("i-failed-1"), Reason: aws.String("first reason")},
				{AssociationType: aws.String("INSTANCE_ID"), AssociationValue: aws.String("i-failed-2"), Reason: aws.String("second reason")},
			},
		)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "i-failed-1") || !strings.Contains(err.Error(), "first reason") || !strings.Contains(err.Error(), "i-failed-2") || !strings.Contains(err.Error(), "second reason") {
			t.Fatalf("error = %s", err)
		}
	})
}
