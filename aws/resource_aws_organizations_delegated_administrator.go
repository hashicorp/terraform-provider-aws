package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsOrganizationsDelegatedAdministrator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsOrganizationsDelegatedAdministratorCreate,
		ReadWithoutTimeout:   resourceAwsOrganizationsDelegatedAdministratorRead,
		DeleteWithoutTimeout: resourceAwsOrganizationsDelegatedAdministratorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"service_principal": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsOrganizationsDelegatedAdministratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)
	input := &organizations.RegisterDelegatedAdministratorInput{
		AccountId:        aws.String(accountID),
		ServicePrincipal: aws.String(d.Get("service_principal").(string)),
	}

	_, err := conn.RegisterDelegatedAdministratorWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Organizations DelegatedAdministrator (%s): %w", name, err))
	}

	d.SetId(accountID)

	return resourceAwsOrganizationsDelegatedAdministratorRead(ctx, d, meta)
}

func resourceAwsOrganizationsDelegatedAdministratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn
	input := &organizations.ListDelegatedAdministratorsInput{
		ServicePrincipal: aws.String(d.Get("service_principal").(string)),
	}
	var delegatedAccount *organizations.DelegatedAdministrator
	err := conn.ListDelegatedAdministratorsPagesWithContext(ctx, input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
		for _, delegated := range page.DelegatedAdministrators {
			if aws.StringValue(delegated.Id) != d.Id() {
				delegatedAccount = delegated
			}
		}

		return !lastPage
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing AWS Organization (%s) DelegatedAdministrators: %w", d.Id(), err))
	}

	d.Set("arn", delegatedAccount.Arn)
	d.Set("delegation_enabled_date", aws.TimeValue(delegatedAccount.DelegationEnabledDate).Format(time.RFC3339))
	d.Set("email", delegatedAccount.Email)
	d.Set("joined_method", delegatedAccount.Status)
	d.Set("joined_timestamp", aws.TimeValue(delegatedAccount.JoinedTimestamp).Format(time.RFC3339))
	d.Set("name", delegatedAccount.Name)
	d.Set("status", delegatedAccount.Status)

	return nil
}

func resourceAwsOrganizationsDelegatedAdministratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.DeregisterDelegatedAdministratorInput{
		AccountId:        aws.String(d.Id()),
		ServicePrincipal: aws.String(d.Get("service_principal").(string)),
	}

	_, err := conn.DeregisterDelegatedAdministratorWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Organizations DelegatedAdministrator (%s): %w", d.Id(), err))
	}
	return nil
}
