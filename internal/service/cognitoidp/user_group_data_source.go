// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cognito_user_group", name="User Group")
func DataSourceUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
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
	}
}

const (
	DSNameUserGroup = "User Group Data Source"
)

func dataSourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	name := d.Get("name").(string)
	userPoolId := d.Get("user_pool_id").(string)
	params := &cognitoidentityprovider.GetGroupInput{
		GroupName:  aws.String(name),
		UserPoolId: aws.String(userPoolId),
	}

	log.Print("[DEBUG] Reading Cognito User Group")

	resp, err := conn.GetGroupWithContext(ctx, params)
	if tfresource.NotFound(err) || err != nil {
		return append(diags, create.DiagError(names.CognitoIDP, create.ErrActionReading, DSNameUserGroup, name, err)...)
	}

	d.SetId(
		fmt.Sprintf("%s/%s", name, userPoolId),
	)

	d.Set("description", resp.Group.Description)
	d.Set("precedence", resp.Group.Precedence)
	d.Set("role_arn", resp.Group.RoleArn)
	return diags
}
