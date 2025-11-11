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
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identifier of the account.",
			},
			"arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ARN of the account.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the account.",
			},
			"joined_method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Method by which the account joined the organization.",
			},
			"joined_timestamp": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date the account became a part of the organization (RFC3339 format).",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the account.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "State of the account.",
			},
			"principal_org_path": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The full path of the account within the AWS Organization. " +
					"See https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_last-accessed-view-data-orgs.html#access_policies_last-accessed-viewing-orgs-entity-path",
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID := d.Get("id").(string)

	accountOut, err := client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: aws.String(accountID),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading account %s: %s", accountID, err)
	}
	if accountOut == nil || accountOut.Account == nil {
		return sdkdiag.AppendErrorf(diags, "no account information returned for %s", accountID)
	}

	acc := accountOut.Account
	id := aws.ToString(acc.Id)

	d.SetId(id)
	d.Set("id", id)
	d.Set("arn", acc.Arn)
	d.Set("email", acc.Email)
	d.Set("joined_method", acc.JoinedMethod)
	if acc.JoinedTimestamp != nil {
		d.Set("joined_timestamp", acc.JoinedTimestamp.Format(time.RFC3339))
	}
	d.Set("name", acc.Name)
	d.Set("state", acc.State)

	path, err := intOrg.BuildPrincipalOrgPath(ctx, client, id)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "building principal org path for account %s: %s", id, err)
	}
	d.Set("principal_org_path", path)

	return diags
}
