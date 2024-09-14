// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_securityhub_standards_control_associations", name="Standards Control Associations")
func dataSourceStandardsControlAssociations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStandardsControlAssociationsRead,
		Schema: map[string]*schema.Schema{
			"security_control_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"standards_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceStandardsControlAssociationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.ListStandardsControlAssociationsInput{
		SecurityControlId: aws.String(d.Get("security_control_id").(string)),
	}

	standardsControlAssociations, err := findStandardsControlAssociations(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Standards Control Associations (%s): %s", d.Id(), err)
	}

	var standardsArns []string

	for _, v := range standardsControlAssociations {
		standardsArns = append(standardsArns, aws.ToString(v.StandardsArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("standards_arns", standardsArns)

	return diags
}

func findStandardsControlAssociations(ctx context.Context, conn *securityhub.Client, input *securityhub.ListStandardsControlAssociationsInput) ([]types.StandardsControlAssociationSummary, error) {
	var output []types.StandardsControlAssociationSummary

	pages := securityhub.NewListStandardsControlAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.StandardsControlAssociationSummaries...)
	}

	return output, nil
}
