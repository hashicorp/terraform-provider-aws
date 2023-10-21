// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cognito_user_groups", name="User Groups")
func DataSourceUserGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserGroupsRead,
		Schema: map[string]*schema.Schema{
			"user_pool_id": {
				Type:         schema.TypeString,
				ValidateFunc: validUserPoolID,
				Required:     true,
				ForceNew:     true,
			},
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"precedence": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

const (
	DSNameUserGroups = "User Groups Data Source"
)

func dataSourceUserGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	userPoolId := d.Get("user_pool_id").(string)

	params := &cognitoidentityprovider.ListGroupsInput{
		UserPoolId: aws.String(userPoolId),
	}
	resp, err := conn.ListGroupsWithContext(ctx, params)
	if err != nil {
		return append(diags, create.DiagError(names.CognitoIDP, create.ErrActionReading, DSNameUserGroups, userPoolId, err)...)
	}

	d.SetId(userPoolId)
	if err := d.Set("groups", flattenUserGroups(ctx, resp.Groups)); err != nil {
		return append(diags, create.DiagError(names.CognitoIDP, create.ErrActionSetting, DSNameUserGroups, d.Id(), err)...)
	}
	return diags
}

func flattenUserGroups(ctx context.Context, groups []*cognitoidentityprovider.GroupType) []interface{} {
	results := make([]interface{}, 0, len(groups))
	for _, group := range groups {
		results = append(results, map[string]interface{}{
			"name":        group.GroupName,
			"description": group.Description,
			"precedence":  group.Precedence,
			"role_arn":    group.RoleArn,
		})
	}
	return results
}
