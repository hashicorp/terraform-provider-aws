// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestDBBackupConfigurationCustomScheduleValidator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		model     dbBackupConfigurationModel
		wantError bool
	}{
		"custom schedule with expression": {
			model: dbBackupConfigurationModel{
				CustomSchedule: types.StringValue("cron(0 0 * * ? *)"),
				Enabled:        types.BoolValue(true),
				RetentionDays:  types.Int32Value(7),
				Type:           fwtypes.StringEnumValue(awstypes.AutomatedDbBackupTypeCustomSchedule),
			},
			wantError: false,
		},
		"custom schedule without expression": {
			model: dbBackupConfigurationModel{
				CustomSchedule: types.StringNull(),
				Enabled:        types.BoolValue(true),
				RetentionDays:  types.Int32Value(7),
				Type:           fwtypes.StringEnumValue(awstypes.AutomatedDbBackupTypeCustomSchedule),
			},
			wantError: true,
		},
		"custom schedule with empty expression": {
			model: dbBackupConfigurationModel{
				CustomSchedule: types.StringValue(""),
				Enabled:        types.BoolValue(true),
				RetentionDays:  types.Int32Value(7),
				Type:           fwtypes.StringEnumValue(awstypes.AutomatedDbBackupTypeCustomSchedule),
			},
			wantError: true,
		},
		"daily without expression": {
			model: dbBackupConfigurationModel{
				CustomSchedule: types.StringNull(),
				Enabled:        types.BoolValue(true),
				RetentionDays:  types.Int32Value(7),
				Type:           fwtypes.StringEnumValue(awstypes.AutomatedDbBackupTypeDaily),
			},
			wantError: false,
		},
		"daily with expression": {
			model: dbBackupConfigurationModel{
				CustomSchedule: types.StringValue("cron(0 0 * * ? *)"),
				Enabled:        types.BoolValue(true),
				RetentionDays:  types.Int32Value(7),
				Type:           fwtypes.StringEnumValue(awstypes.AutomatedDbBackupTypeDaily),
			},
			wantError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			objectValue, diags := fwtypes.NewObjectValueOf(ctx, &testCase.model)
			if diags.HasError() {
				t.Fatalf("building object value: %v", diags)
			}

			req := validator.ObjectRequest{
				Path:        path.Root("db_backup_configuration"),
				ConfigValue: objectValue.ObjectValue,
			}
			resp := &validator.ObjectResponse{}

			dbBackupConfigurationCustomScheduleValidator{}.ValidateObject(ctx, req, resp)

			if got, want := resp.Diagnostics.HasError(), testCase.wantError; got != want {
				t.Errorf("HasError() = %t, want %t: %s", got, want, resp.Diagnostics)
			}
		})
	}
}
