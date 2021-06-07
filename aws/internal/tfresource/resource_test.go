package tfresource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestDeleteNoError(t *testing.T) {
	theError := errors.New("fail")

	testCases := []struct {
		Name        string
		Resource    *schema.Resource
		ExpectError bool
	}{
		{
			Name: "no error Delete function",
			Resource: &schema.Resource{
				Delete: func(rd *schema.ResourceData, i interface{}) error {
					return nil
				},
			},
		},
		{
			Name: "no error DeleteContext function",
			Resource: &schema.Resource{
				DeleteContext: func(c context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {
					return nil
				},
			},
		},
		{
			Name: "no error DeleteWithoutTimeout function",
			Resource: &schema.Resource{
				DeleteWithoutTimeout: func(c context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {
					return nil
				},
			},
		},
		{
			Name: "error Delete function",
			Resource: &schema.Resource{
				Delete: func(rd *schema.ResourceData, i interface{}) error {
					return theError
				},
			},
			ExpectError: true,
		},
		{
			Name: "error DeleteContext function",
			Resource: &schema.Resource{
				DeleteContext: func(c context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {
					return diag.FromErr(theError)
				},
			},
			ExpectError: true,
		},
		{
			Name: "error DeleteWithoutTimeout function",
			Resource: &schema.Resource{
				DeleteWithoutTimeout: func(c context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {
					return diag.FromErr(theError)
				},
			},
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := testCase.Resource
			d := r.Data(nil)

			err := tfresource.Delete(r, d, nil)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
