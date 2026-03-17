// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of special types:
// - timestamptypes.RFC3339
// - types.ARN
// - types.JSON and fwtypes.SmithyJSON

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	smithydocument "github.com/aws/smithy-go/document"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

type tfRFC3339Time struct {
	CreationDateTime timetypes.RFC3339 `tfsdk:"creation_date_time"`
}

type awsRFC3339TimePointer struct {
	CreationDateTime *time.Time
}

type awsRFC3339TimeValue struct {
	CreationDateTime time.Time
}

type tfSingleARNField struct {
	Field1 fwtypes.ARN `tfsdk:"field1"`
}

var _ tfsmithy.JSONStringer = (*testJSONDocument)(nil)
var _ smithydocument.Marshaler = (*testJSONDocument)(nil)

type testJSONDocument struct {
	Value any
}

func newTestJSONDocument(v any) tfsmithy.JSONStringer {
	return &testJSONDocument{Value: v}
}

func (m *testJSONDocument) UnmarshalSmithyDocument(v any) error {
	data, err := tfjson.EncodeToBytes(m.Value)
	if err != nil {
		return err
	}
	return tfjson.DecodeFromBytes(data, v)
}

func (m *testJSONDocument) MarshalSmithyDocument() ([]byte, error) {
	return tfjson.EncodeToBytes(m.Value)
}

var _ tfsmithy.JSONStringer = &testJSONDocumentError{}

type testJSONDocumentError struct{}

func (m *testJSONDocumentError) UnmarshalSmithyDocument(v any) error {
	return errUnmarshallSmithyDocument
}

func (m *testJSONDocumentError) MarshalSmithyDocument() ([]byte, error) {
	return nil, errMarshallSmithyDocument
}

var (
	errUnmarshallSmithyDocument = errors.New("test unmarshal error")
	errMarshallSmithyDocument   = errors.New("test marshal error")
)

type awsJSONStringer struct {
	Field1 tfsmithy.JSONStringer `json:"field1"`
}

type tfJSONStringer struct {
	Field1 fwtypes.SmithyJSON[tfsmithy.JSONStringer] `tfsdk:"field1"`
}

func TestExpandSpecialTypes(t *testing.T) {
	t.Parallel()

	testARN := "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1" //lintignore:AWSAT003,AWSAT005

	testTimeStr := "2013-09-25T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))

	testCases := map[string]autoFlexTestCases{
		"timestamp": {
			"timestamp pointer": {
				Source: &tfRFC3339Time{
					CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
				},
				Target: &awsRFC3339TimePointer{},
				WantTarget: &awsRFC3339TimePointer{
					CreationDateTime: &testTimeTime,
				},
			},
			"timestamp": {
				Source: &tfRFC3339Time{
					CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
				},
				Target: &awsRFC3339TimeValue{},
				WantTarget: &awsRFC3339TimeValue{
					CreationDateTime: testTimeTime,
				},
			},
		},

		"single ARN": {
			"single ARN Source and single string Target": {
				Source:     &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
				Target:     &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{Field1: testARN},
			},
			"single ARN Source and single *string Target": {
				Source:     &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
				Target:     &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{Field1: aws.String(testARN)},
			},
		},

		"json": {
			"JSONValue Source to json interface Target": {
				Source: &tfJSONStringer{Field1: fwtypes.NewSmithyJSONValue(`{"field1": "a"}`, newTestJSONDocument)},
				Target: &awsJSONStringer{},
				WantTarget: &awsJSONStringer{
					Field1: &testJSONDocument{
						Value: map[string]any{
							"field1": "a",
						},
					},
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenSpecialTypes(t *testing.T) {
	t.Parallel()

	testARN := "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1" //lintignore:AWSAT003,AWSAT005

	testTimeStr := "2013-09-25T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))
	var zeroTime time.Time

	testCases := map[string]autoFlexTestCases{
		"single ARN": {
			"single string Source and single ARN Target": {
				Source:     &awsSingleStringValue{Field1: testARN},
				Target:     &tfSingleARNField{},
				WantTarget: &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
			},
			"single *string Source and single ARN Target": {
				Source:     &awsSingleStringPointer{Field1: aws.String(testARN)},
				Target:     &tfSingleARNField{},
				WantTarget: &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
			},
			"single nil *string Source and single ARN Target": {
				Source:     &awsSingleStringPointer{},
				Target:     &tfSingleARNField{},
				WantTarget: &tfSingleARNField{Field1: fwtypes.ARNNull()},
			},
		},
		"timestamp": {
			"timestamp": {
				Source: &awsRFC3339TimeValue{
					CreationDateTime: testTimeTime,
				},
				Target: &tfRFC3339Time{},
				WantTarget: &tfRFC3339Time{
					CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
				},
			},
			"timestamp pointer": {
				Source: &awsRFC3339TimePointer{
					CreationDateTime: &testTimeTime,
				},
				Target: &tfRFC3339Time{},
				WantTarget: &tfRFC3339Time{
					CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
				},
			},
			"timestamp nil": {
				Source: &awsRFC3339TimePointer{},
				Target: &tfRFC3339Time{},
				WantTarget: &tfRFC3339Time{
					CreationDateTime: timetypes.NewRFC3339Null(),
				},
			},
			"timestamp empty": {
				Source: &awsRFC3339TimeValue{},
				Target: &tfRFC3339Time{},
				WantTarget: &tfRFC3339Time{
					CreationDateTime: timetypes.NewRFC3339TimeValue(zeroTime),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenJSONInterfaceToStringTypable(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"json interface Source string Target": {
			Source: &awsJSONStringer{
				Field1: &testJSONDocument{
					Value: &struct {
						Test string `json:"test"`
					}{
						Test: "a",
					},
				},
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringValue(`{"test":"a"}`),
			},
		},
		"null json interface Source string Target": {
			Source: &awsJSONStringer{
				Field1: nil,
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringNull(),
			},
		},

		"json interface Source JSONValue Target": {
			Source: &awsJSONStringer{
				Field1: &testJSONDocument{
					Value: &struct {
						Test string `json:"test"`
					}{
						Test: "a",
					},
				},
			},
			Target: &tfJSONStringer{},
			WantTarget: &tfJSONStringer{
				Field1: fwtypes.NewSmithyJSONValue(`{"test":"a"}`, newTestJSONDocument),
			},
		},
		"null json interface Source JSONValue Target": {
			Source: &awsJSONStringer{
				Field1: nil,
			},
			Target: &tfJSONStringer{},
			WantTarget: &tfJSONStringer{
				Field1: fwtypes.NewSmithyJSONNull[tfsmithy.JSONStringer](),
			},
		},

		"json interface Source marshal error": {
			Source: &awsJSONStringer{
				Field1: &testJSONDocumentError{},
			},
			Target:        &tfSingleStringField{},
			ExpectedDiags: diagAFTypeErr[*testJSONDocumentError](diagFlatteningMarshalSmithyDocument, errMarshallSmithyDocument),
		},

		"non-json interface Source string Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringNull(),
			},
		},

		"null non-json interface Source string Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringNull(),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}
