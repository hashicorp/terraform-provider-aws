// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFunctionKMSKeyARNValidation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	r, err := tflambdaweb.ResourceFunction(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	revisionConfig, ok := schemaResp.Schema.Blocks["revision_config"].(schema.ListNestedBlock)
	if !ok {
		t.Fatal("revision_config is not a list nested block")
	}
	kmsKeyARN, ok := revisionConfig.NestedObject.Attributes[names.AttrKMSKeyARN].(schema.StringAttribute)
	if !ok {
		t.Fatal("kms_key_arn is not a string attribute")
	}

	testCases := map[string]struct {
		value     string
		wantError bool
	}{
		"key ARN": {
			value: "arn:aws:kms:us-east-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab", //lintignore:AWSAT003,AWSAT005
		},
		"multi-Region key ARN": {
			value: "arn:aws:kms:eu-west-1:123456789012:key/mrk-1234abcd12ab34cd56ef1234567890ab", //lintignore:AWSAT003,AWSAT005
		},
		"other partition": {
			value: "arn:aws-eusc:kms:eusc-de-east-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab", //lintignore:AWSAT003,AWSAT005
		},
		"alias ARN": {
			value:     "arn:aws:kms:us-east-1:123456789012:alias/example", //lintignore:AWSAT003,AWSAT005
			wantError: true,
		},
		"non-KMS ARN": {
			value:     "arn:aws:s3:::example-bucket", //lintignore:AWSAT003,AWSAT005
			wantError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var resp validator.StringResponse
			for _, v := range kmsKeyARN.Validators {
				v.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(testCase.value)}, &resp)
			}

			if got := resp.Diagnostics.HasError(); got != testCase.wantError {
				t.Errorf("value %q: got error %t, want %t (%v)", testCase.value, got, testCase.wantError, resp.Diagnostics)
			}
		})
	}
}
