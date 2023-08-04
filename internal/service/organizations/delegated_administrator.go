// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_organizations_delegated_administrator")
func ResourceDelegatedAdministrator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDelegatedAdministratorCreate,
		ReadWithoutTimeout:   resourceDelegatedAdministratorRead,
		DeleteWithoutTimeout: resourceDelegatedAdministratorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
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
			"service_principal": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDelegatedAdministratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	accountID := d.Get("account_id").(string)
	servicePrincipal := d.Get("service_principal").(string)
	id := DelegatedAdministratorCreateResourceID(accountID, servicePrincipal)
	input := &organizations.RegisterDelegatedAdministratorInput{
		AccountId:        aws.String(accountID),
		ServicePrincipal: aws.String(servicePrincipal),
	}

	_, err := conn.RegisterDelegatedAdministratorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Delegated Administrator (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceDelegatedAdministratorRead(ctx, d, meta)...)
}

func resourceDelegatedAdministratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	accountID, servicePrincipal, err := DelegatedAdministratorParseResourceID(d.Id())

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

	d.Set("account_id", accountID)
	d.Set("arn", delegatedAccount.Arn)
	d.Set("delegation_enabled_date", aws.TimeValue(delegatedAccount.DelegationEnabledDate).Format(time.RFC3339))
	d.Set("email", delegatedAccount.Email)
	d.Set("joined_method", delegatedAccount.JoinedMethod)
	d.Set("joined_timestamp", aws.TimeValue(delegatedAccount.JoinedTimestamp).Format(time.RFC3339))
	d.Set("name", delegatedAccount.Name)
	d.Set("service_principal", servicePrincipal)
	d.Set("status", delegatedAccount.Status)

	return diags
}

func resourceDelegatedAdministratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	accountID, servicePrincipal, err := DelegatedAdministratorParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Organizations Delegated Administrator: %s", d.Id())
	_, err = conn.DeregisterDelegatedAdministratorWithContext(ctx, &organizations.DeregisterDelegatedAdministratorInput{
		AccountId:        aws.String(accountID),
		ServicePrincipal: aws.String(servicePrincipal),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Delegated Administrator (%s): %s", d.Id(), err)
	}

	return diags
}

func findDelegatedAdministratorByTwoPartKey(ctx context.Context, conn *organizations.Organizations, accountID, servicePrincipal string) (*organizations.DelegatedAdministrator, error) {
	input := &organizations.ListDelegatedAdministratorsInput{
		ServicePrincipal: aws.String(servicePrincipal),
	}

	output, err := findDelegatedAdministrators(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.Id) == accountID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findDelegatedAdministrators(ctx context.Context, conn *organizations.Organizations, input *organizations.ListDelegatedAdministratorsInput) ([]*organizations.DelegatedAdministrator, error) {
	var output []*organizations.DelegatedAdministrator

	err := conn.ListDelegatedAdministratorsPagesWithContext(ctx, input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.DelegatedAdministrators...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

const delegatedAdministratorResourceIDSeparator = "/"

func DelegatedAdministratorCreateResourceID(accountID, servicePrincipal string) string {
	parts := []string{accountID, servicePrincipal}
	id := strings.Join(parts, delegatedAdministratorResourceIDSeparator)

	return id
}

func DelegatedAdministratorParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, delegatedAdministratorResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ACCOUNTID%[2]sSERVICEPRINCIPAL", id, delegatedAdministratorResourceIDSeparator)
}
