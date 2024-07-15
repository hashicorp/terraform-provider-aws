// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_delegated_administrator", name="Delegated Administrator")
func resourceDelegatedAdministrator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDelegatedAdministratorCreate,
		ReadWithoutTimeout:   resourceDelegatedAdministratorRead,
		DeleteWithoutTimeout: resourceDelegatedAdministratorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delegation_enabled_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEmail: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_principal": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDelegatedAdministratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	servicePrincipal := d.Get("service_principal").(string)
	id := delegatedAdministratorCreateResourceID(accountID, servicePrincipal)
	input := &organizations.RegisterDelegatedAdministratorInput{
		AccountId:        aws.String(accountID),
		ServicePrincipal: aws.String(servicePrincipal),
	}

	_, err := conn.RegisterDelegatedAdministrator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Delegated Administrator (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceDelegatedAdministratorRead(ctx, d, meta)...)
}

func resourceDelegatedAdministratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID, servicePrincipal, err := delegatedAdministratorParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	delegatedAccount, err := findDelegatedAdministratorByTwoPartKey(ctx, conn, accountID, servicePrincipal)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Delegated Administrator %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Delegated Administrator (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrARN, delegatedAccount.Arn)
	d.Set("delegation_enabled_date", aws.ToTime(delegatedAccount.DelegationEnabledDate).Format(time.RFC3339))
	d.Set(names.AttrEmail, delegatedAccount.Email)
	d.Set("joined_method", delegatedAccount.JoinedMethod)
	d.Set("joined_timestamp", aws.ToTime(delegatedAccount.JoinedTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, delegatedAccount.Name)
	d.Set("service_principal", servicePrincipal)
	d.Set(names.AttrStatus, delegatedAccount.Status)

	return diags
}

func resourceDelegatedAdministratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID, servicePrincipal, err := delegatedAdministratorParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Organizations Delegated Administrator: %s", d.Id())
	_, err = conn.DeregisterDelegatedAdministrator(ctx, &organizations.DeregisterDelegatedAdministratorInput{
		AccountId:        aws.String(accountID),
		ServicePrincipal: aws.String(servicePrincipal),
	})

	if errs.IsA[*awstypes.AccountNotRegisteredException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Delegated Administrator (%s): %s", d.Id(), err)
	}

	return diags
}

func findDelegatedAdministratorByTwoPartKey(ctx context.Context, conn *organizations.Client, accountID, servicePrincipal string) (*awstypes.DelegatedAdministrator, error) {
	input := &organizations.ListDelegatedAdministratorsInput{
		ServicePrincipal: aws.String(servicePrincipal),
	}

	return findDelegatedAdministrator(ctx, conn, input, func(v *awstypes.DelegatedAdministrator) bool {
		return aws.ToString(v.Id) == accountID
	})
}

func findDelegatedAdministrator(ctx context.Context, conn *organizations.Client, input *organizations.ListDelegatedAdministratorsInput, filter tfslices.Predicate[*awstypes.DelegatedAdministrator]) (*awstypes.DelegatedAdministrator, error) {
	output, err := findDelegatedAdministrators(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDelegatedAdministrators(ctx context.Context, conn *organizations.Client, input *organizations.ListDelegatedAdministratorsInput, filter tfslices.Predicate[*awstypes.DelegatedAdministrator]) ([]awstypes.DelegatedAdministrator, error) {
	var output []awstypes.DelegatedAdministrator

	pages := organizations.NewListDelegatedAdministratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DelegatedAdministrators {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

const delegatedAdministratorResourceIDSeparator = "/"

func delegatedAdministratorCreateResourceID(accountID, servicePrincipal string) string {
	parts := []string{accountID, servicePrincipal}
	id := strings.Join(parts, delegatedAdministratorResourceIDSeparator)

	return id
}

func delegatedAdministratorParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, delegatedAdministratorResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ACCOUNTID%[2]sSERVICEPRINCIPAL", id, delegatedAdministratorResourceIDSeparator)
}
