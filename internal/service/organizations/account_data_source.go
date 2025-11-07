package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	intOrg "github.com/hashicorp/terraform-provider-aws/internal/organizations"
)

func DataSourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountRead,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal_org_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID := d.Get("account_id").(string)
	d.SetId(accountID)

	accountOut, err := client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: aws.String(accountID),
	})
	if err == nil && accountOut.Account != nil {
		d.Set("arn", accountOut.Account.Arn)
	}

	path, err := intOrg.BuildPrincipalOrgPath(ctx, client, accountID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "building principal org path for Account %s: %s", accountID, err)
	}
	d.Set("principal_org_path", path)

	return diags
}
