// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_organizations_delegated_administrators")
func DataSourceDelegatedAdministrators() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDelegatedAdministratorsRead,

		Schema: map[string]*schema.Schema{
			"delegated_administrators": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delegation_enabled_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"service_principal": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func dataSourceDelegatedAdministratorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	input := &organizations.ListDelegatedAdministratorsInput{}

	if v, ok := d.GetOk("service_principal"); ok {
		input.ServicePrincipal = aws.String(v.(string))
	}

	output, err := findDelegatedAdministrators(ctx, conn, input)

	if err != nil {
		return diag.Errorf("reading Organizations Delegated Administrators: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	if err = d.Set("delegated_administrators", flattenDelegatedAdministrators(output)); err != nil {
		return diag.Errorf("setting delegated_administrators: %s", err)
	}

	return nil
}

func flattenDelegatedAdministrators(apiObjects []*organizations.DelegatedAdministrator) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"arn":                     aws.StringValue(apiObject.Arn),
			"delegation_enabled_date": aws.TimeValue(apiObject.DelegationEnabledDate).Format(time.RFC3339),
			"email":                   aws.StringValue(apiObject.Email),
			"id":                      aws.StringValue(apiObject.Id),
			"joined_method":           aws.StringValue(apiObject.JoinedMethod),
			"joined_timestamp":        aws.TimeValue(apiObject.JoinedTimestamp).Format(time.RFC3339),
			"name":                    aws.StringValue(apiObject.Name),
			"status":                  aws.StringValue(apiObject.Status),
		})
	}

	return tfList
}
