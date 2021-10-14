package rds_test

import (
	"context"
	"reflect"
	"testing"

	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestResourceAwsDbInstanceStateUpgradeV0(t *testing.T) {
	testCases := []struct {
		Description   string
		InputState    map[string]interface{}
		ExpectedState map[string]interface{}
	}{
		{
			Description:   "missing state",
			InputState:    nil,
			ExpectedState: nil,
		},
		{
			Description: "adds delete_automated_backups",
			InputState: map[string]interface{}{
				"allocated_storage": 10,
				"engine":            "mariadb",
				"identifier":        "my-test-instance",
				"instance_class":    "db.t2.micro",
				"password":          "avoid-plaintext-passwords",
				"username":          "tfacctest",
				"tags":              map[string]interface{}{"key1": "value1"},
			},
			ExpectedState: map[string]interface{}{
				"allocated_storage":        10,
				"delete_automated_backups": true,
				"engine":                   "mariadb",
				"identifier":               "my-test-instance",
				"instance_class":           "db.t2.micro",
				"password":                 "avoid-plaintext-passwords",
				"username":                 "tfacctest",
				"tags":                     map[string]interface{}{"key1": "value1"},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Description, func(t *testing.T) {
			got, err := tfrds.InstanceStateUpgradeV0(context.Background(), testCase.InputState, nil)

			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			if !reflect.DeepEqual(testCase.ExpectedState, got) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", testCase.ExpectedState, got)
			}
		})
	}
}
