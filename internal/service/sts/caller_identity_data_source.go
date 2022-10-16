package sts

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCallerIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCallerIdentityRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCallerIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).STSConn

	output, err := FindCallerIdentity(ctx, conn)

	if err != nil {
		return diag.Errorf("reading STS Caller Identity: %s", err)
	}

	accountID := aws.StringValue(output.Account)
	d.SetId(accountID)
	d.Set("account_id", accountID)
	d.Set("arn", output.Arn)
	d.Set("user_id", output.UserId)

	return nil
}
