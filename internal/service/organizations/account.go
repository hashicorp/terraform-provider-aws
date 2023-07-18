// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_account", name="Account")
// @Tags(identifierAttribute="id")
func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountCreate,
		ReadWithoutTimeout:   resourceAccountRead,
		UpdateWithoutTimeout: resourceAccountUpdate,
		DeleteWithoutTimeout: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"close_on_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"create_govcloud": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email": {
				ForceNew: true,
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(6, 64),
					validation.StringMatch(regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`), "must be a valid email address"),
				),
			},
			"govcloud_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_user_access_to_billing": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{organizations.IAMUserAccessToBillingAllow, organizations.IAMUserAccessToBillingDeny}, true),
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
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"parent_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_MoveAccount.html#organizations-MoveAccount-request-DestinationParentId"),
			},
			"role_name": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]{1,64}$`), "must consist of uppercase letters, lowercase letters, digits with no spaces, and any of the following characters"),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	var iamUserAccessToBilling *string

	if v, ok := d.GetOk("iam_user_access_to_billing"); ok {
		iamUserAccessToBilling = aws.String(v.(string))
	}

	var roleName *string

	if v, ok := d.GetOk("role_name"); ok {
		roleName = aws.String(v.(string))
	}

	s, err := createAccount(ctx, conn,
		d.Get("name").(string),
		d.Get("email").(string),
		iamUserAccessToBilling,
		roleName,
		getTagsIn(ctx),
		d.Get("create_govcloud").(bool),
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AWS Organizations Account (%s): %s", d.Get("name").(string), err)
	}

	output, err := waitAccountCreated(ctx, conn, aws.StringValue(s.Id))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AWS Organizations Account (%s) create: %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(output.AccountId))
	d.Set("govcloud_id", output.GovCloudAccountId)

	if v, ok := d.GetOk("parent_id"); ok {
		oldParentAccountID, err := findParentAccountID(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading AWS Organizations Account (%s) parent: %s", d.Id(), err)
		}

		if newParentAccountID := v.(string); newParentAccountID != oldParentAccountID {
			input := &organizations.MoveAccountInput{
				AccountId:           aws.String(d.Id()),
				DestinationParentId: aws.String(newParentAccountID),
				SourceParentId:      aws.String(oldParentAccountID),
			}

			log.Printf("[DEBUG] Moving AWS Organizations Account: %s", input)
			if _, err := conn.MoveAccountWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "moving AWS Organizations Account (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	account, err := FindAccountByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AWS Organizations Account does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AWS Organizations Account (%s): %s", d.Id(), err)
	}

	parentAccountID, err := findParentAccountID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AWS Organizations Account (%s) parent: %s", d.Id(), err)
	}

	d.Set("arn", account.Arn)
	d.Set("email", account.Email)
	d.Set("joined_method", account.JoinedMethod)
	d.Set("joined_timestamp", aws.TimeValue(account.JoinedTimestamp).Format(time.RFC3339))
	d.Set("name", account.Name)
	d.Set("parent_id", parentAccountID)
	d.Set("status", account.Status)

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if d.HasChange("parent_id") {
		o, n := d.GetChange("parent_id")

		input := &organizations.MoveAccountInput{
			AccountId:           aws.String(d.Id()),
			SourceParentId:      aws.String(o.(string)),
			DestinationParentId: aws.String(n.(string)),
		}

		if _, err := conn.MoveAccountWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "moving AWS Organizations Account (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	close := d.Get("close_on_deletion").(bool)
	var err error

	if close {
		log.Printf("[DEBUG] Closing AWS Organizations Account: %s", d.Id())
		_, err = conn.CloseAccountWithContext(ctx, &organizations.CloseAccountInput{
			AccountId: aws.String(d.Id()),
		})
	} else {
		log.Printf("[DEBUG] Removing AWS Organizations Account from organization: %s", d.Id())
		_, err = conn.RemoveAccountFromOrganizationWithContext(ctx, &organizations.RemoveAccountFromOrganizationInput{
			AccountId: aws.String(d.Id()),
		})
	}

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAccountNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AWS Organizations Account (%s): %s", d.Id(), err)
	}

	if close {
		if _, err := waitAccountDeleted(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for AWS Organizations Account (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}

func createAccount(ctx context.Context, conn *organizations.Organizations, name, email string, iamUserAccessToBilling, roleName *string, tags []*organizations.Tag, govCloud bool) (*organizations.CreateAccountStatus, error) {
	if govCloud {
		input := &organizations.CreateGovCloudAccountInput{
			AccountName: aws.String(name),
			Email:       aws.String(email),
		}

		if iamUserAccessToBilling != nil {
			input.IamUserAccessToBilling = iamUserAccessToBilling
		}

		if roleName != nil {
			input.RoleName = roleName
		}

		if len(tags) > 0 {
			input.Tags = tags
		}

		log.Printf("[DEBUG] Creating AWS Organizations Account with GovCloud Account: %s", input)
		outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 4*time.Minute,
			func() (interface{}, error) {
				return conn.CreateGovCloudAccountWithContext(ctx, input)
			},
			organizations.ErrCodeFinalizingOrganizationException,
		)

		if err != nil {
			return nil, err
		}

		return outputRaw.(*organizations.CreateGovCloudAccountOutput).CreateAccountStatus, nil
	}

	input := &organizations.CreateAccountInput{
		AccountName: aws.String(name),
		Email:       aws.String(email),
	}

	if iamUserAccessToBilling != nil {
		input.IamUserAccessToBilling = iamUserAccessToBilling
	}

	if roleName != nil {
		input.RoleName = roleName
	}

	if len(tags) > 0 {
		input.Tags = tags
	}

	log.Printf("[DEBUG] Creating AWS Organizations Account: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 4*time.Minute,
		func() (interface{}, error) {
			return conn.CreateAccountWithContext(ctx, input)
		},
		organizations.ErrCodeFinalizingOrganizationException,
	)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*organizations.CreateAccountOutput).CreateAccountStatus, nil
}

func findParentAccountID(ctx context.Context, conn *organizations.Organizations, id string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(id),
	}
	var output []*organizations.Parent

	err := conn.ListParentsPagesWithContext(ctx, input, func(page *organizations.ListParentsOutput, lastPage bool) bool {
		output = append(output, page.Parents...)

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if len(output) == 0 || output[0] == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	if count := len(output); count > 1 {
		return "", tfresource.NewTooManyResultsError(count, input)
	}

	return aws.StringValue(output[0].Id), nil
}

func findCreateAccountStatusByID(ctx context.Context, conn *organizations.Organizations, id string) (*organizations.CreateAccountStatus, error) {
	input := &organizations.DescribeCreateAccountStatusInput{
		CreateAccountRequestId: aws.String(id),
	}

	output, err := conn.DescribeCreateAccountStatusWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeCreateAccountStatusNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CreateAccountStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CreateAccountStatus, nil
}

func statusCreateAccountState(ctx context.Context, conn *organizations.Organizations, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCreateAccountStatusByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitAccountCreated(ctx context.Context, conn *organizations.Organizations, id string) (*organizations.CreateAccountStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{organizations.CreateAccountStateInProgress},
		Target:       []string{organizations.CreateAccountStateSucceeded},
		Refresh:      statusCreateAccountState(ctx, conn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*organizations.CreateAccountStatus); ok {
		if state := aws.StringValue(output.State); state == organizations.CreateAccountStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func statusAccountStatus(ctx context.Context, conn *organizations.Organizations, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAccountByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitAccountDeleted(ctx context.Context, conn *organizations.Organizations, id string) (*organizations.Account, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{organizations.AccountStatusPendingClosure, organizations.AccountStatusActive},
		Target:       []string{},
		Refresh:      statusAccountStatus(ctx, conn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*organizations.Account); ok {
		return output, err
	}

	return nil, err
}
