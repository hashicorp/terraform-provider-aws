// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testResourceStreamStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		names.AttrARN:             "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"encryption_type":         "NONE",
		names.AttrKMSKeyID:        "",
		names.AttrName:            "test",
		names.AttrRetentionPeriod: 24,
		"shard_count":             1,
		"shard_level_metrics":     []interface{}{},
		names.AttrTags:            map[string]interface{}{acctest.CtKey1: acctest.CtValue1},
	}
}

func testResourceStreamStateDataV1() map[string]interface{} {
	v0 := testResourceStreamStateDataV0()
	return map[string]interface{}{
		names.AttrARN:               v0[names.AttrARN],
		"encryption_type":           v0["encryption_type"],
		"enforce_consumer_deletion": false,
		names.AttrKMSKeyID:          v0[names.AttrKMSKeyID],
		names.AttrName:              v0[names.AttrName],
		names.AttrRetentionPeriod:   v0[names.AttrRetentionPeriod],
		"shard_count":               v0["shard_count"],
		"shard_level_metrics":       v0["shard_level_metrics"],
		names.AttrTags:              v0[names.AttrTags],
	}
}

func TestStreamStateUpgradeV0(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	expected := testResourceStreamStateDataV1()
	actual, err := tfkinesis.StreamStateUpgradeV0(ctx, testResourceStreamStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
