// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_account", name="Account")
// @Tags(identifierAttribute="id")
func resourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountCreate,
		ReadWithoutTimeout:   resourceAccountRead,
		UpdateWithoutTimeout: resourceAccountUpdate,
		DeleteWithoutTimeout: resourceAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountImportState,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrEmail: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(6, 64),
					validation.StringMatch(regexache.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`), "must be a valid email address"),
				),
			},
			"govcloud_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_user_access_to_billing": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IAMUserAccessToBilling](),
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"parent_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_MoveAccount.html#organizations-MoveAccount-request-DestinationParentId"),
			},
			"role_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[\w+=,.@-]{1,64}$`), "must consist of uppercase letters, lowercase letters, digits with no spaces, and any of the following characters"),
			},
			names.AttrStatus: {
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
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	var status *awstypes.CreateAccountStatus

	if d.Get("create_govcloud").(bool) {
		input := &organizations.CreateGovCloudAccountInput{
			AccountName: aws.String(name),
			Email:       aws.String(d.Get(names.AttrEmail).(string)),
			Tags:        getTagsIn(ctx),
		}

		if v, ok := d.GetOk("iam_user_access_to_billing"); ok {
			input.IamUserAccessToBilling = awstypes.IAMUserAccessToBilling(v.(string))
		}

		if v, ok := d.GetOk("role_name"); ok {
			input.RoleName = aws.String(v.(string))
		}

		outputRaw, err := tfresource.RetryWhenIsA[*awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout,
			func() (interface{}, error) {
				return conn.CreateGovCloudAccount(ctx, input)
			})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating AWS Organizations Account (%s) with GovCloud Account: %s", name, err)
		}

		status = outputRaw.(*organizations.CreateGovCloudAccountOutput).CreateAccountStatus
	} else {
		input := &organizations.CreateAccountInput{
			AccountName: aws.String(name),
			Email:       aws.String(d.Get(names.AttrEmail).(string)),
			Tags:        getTagsIn(ctx),
		}

		if v, ok := d.GetOk("iam_user_access_to_billing"); ok {
			input.IamUserAccessToBilling = awstypes.IAMUserAccessToBilling(v.(string))
		}

		if v, ok := d.GetOk("role_name"); ok {
			input.RoleName = aws.String(v.(string))
		}

		outputRaw, err := tfresource.RetryWhenIsA[*awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout,
			func() (interface{}, error) {
				return conn.CreateAccount(ctx, input)
			})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating AWS Organizations Account (%s): %s", name, err)
		}

		status = outputRaw.(*organizations.CreateAccountOutput).CreateAccountStatus
	}

	output, err := waitAccountCreated(ctx, conn, aws.ToString(status.Id))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AWS Organizations Account (%s) create: %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.ToString(output.AccountId))
	d.Set("govcloud_id", output.GovCloudAccountId)

	if v, ok := d.GetOk("parent_id"); ok {
		oldParentAccountID, err := findParentAccountID(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading AWS Organizations Account (%s) parent: %s", d.Id(), err)
		}

		if newParentAccountID, oldParentAccountID := v.(string), aws.ToString(oldParentAccountID); newParentAccountID != oldParentAccountID {
			input := &organizations.MoveAccountInput{
				AccountId:           aws.String(d.Id()),
				DestinationParentId: aws.String(newParentAccountID),
				SourceParentId:      aws.String(oldParentAccountID),
			}

			_, err := conn.MoveAccount(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "moving AWS Organizations Account (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	account, err := findAccountByID(ctx, conn, d.Id())

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

	d.Set(names.AttrARN, account.Arn)
	d.Set(names.AttrEmail, account.Email)
	d.Set("joined_method", account.JoinedMethod)
	d.Set("joined_timestamp", aws.ToTime(account.JoinedTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, account.Name)
	d.Set("parent_id", parentAccountID)
	d.Set(names.AttrStatus, account.Status)

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if d.HasChange("parent_id") {
		o, n := d.GetChange("parent_id")

		input := &organizations.MoveAccountInput{
			AccountId:           aws.String(d.Id()),
			SourceParentId:      aws.String(o.(string)),
			DestinationParentId: aws.String(n.(string)),
		}

		_, err := conn.MoveAccount(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "moving AWS Organizations Account (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	close := d.Get("close_on_deletion").(bool)
	var err error

	if close {
		log.Printf("[DEBUG] Closing AWS Organizations Account: %s", d.Id())
		_, err = conn.CloseAccount(ctx, &organizations.CloseAccountInput{
			AccountId: aws.String(d.Id()),
		})
	} else {
		log.Printf("[DEBUG] Removing AWS Organizations Account from organization: %s", d.Id())
		_, err = conn.RemoveAccountFromOrganization(ctx, &organizations.RemoveAccountFromOrganizationInput{
			AccountId: aws.String(d.Id()),
		})
	}

	if errs.IsA[*awstypes.AccountNotFoundException](err) {
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

func resourceAccountImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if strings.Contains(d.Id(), "_") {
		parts := strings.Split(d.Id(), "_")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%s), expected <account_id>_<IAM User Access Status> or <account_id>", d.Id())
		}

		d.SetId(parts[0])
		d.Set("iam_user_access_to_billing", parts[1])
	} else {
		d.SetId(d.Id())
	}

	return []*schema.ResourceData{d}, nil
}

func findAccountByID(ctx context.Context, conn *organizations.Client, id string) (*awstypes.Account, error) {
	input := &organizations.DescribeAccountInput{
		AccountId: aws.String(id),
	}

	output, err := conn.DescribeAccount(ctx, input)

	if errs.IsA[*awstypes.AccountNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Account == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.Account.Status; status == awstypes.AccountStatusSuspended {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.Account, nil
}

func findParentAccountID(ctx context.Context, conn *organizations.Client, id string) (*string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(id),
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	output, err := findParent(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output.Id, nil
}

func findParent(ctx context.Context, conn *organizations.Client, input *organizations.ListParentsInput) (*awstypes.Parent, error) {
	output, err := findParents(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findParents(ctx context.Context, conn *organizations.Client, input *organizations.ListParentsInput) ([]awstypes.Parent, error) {
	var output []awstypes.Parent

	pages := organizations.NewListParentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Parents...)
	}

	return output, nil
}

func findCreateAccountStatusByID(ctx context.Context, conn *organizations.Client, id string) (*awstypes.CreateAccountStatus, error) {
	input := &organizations.DescribeCreateAccountStatusInput{
		CreateAccountRequestId: aws.String(id),
	}

	output, err := conn.DescribeCreateAccountStatus(ctx, input)

	if errs.IsA[*awstypes.CreateAccountStatusNotFoundException](err) {
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

func statusCreateAccountState(ctx context.Context, conn *organizations.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCreateAccountStatusByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitAccountCreated(ctx context.Context, conn *organizations.Client, id string) (*awstypes.CreateAccountStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.CreateAccountStateInProgress),
		Target:       enum.Slice(awstypes.CreateAccountStateSucceeded),
		Refresh:      statusCreateAccountState(ctx, conn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CreateAccountStatus); ok {
		if state := output.State; state == awstypes.CreateAccountStateFailed {
			tfresource.SetLastError(err, errors.New(string(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func statusAccountStatus(ctx context.Context, conn *organizations.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccountByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAccountDeleted(ctx context.Context, conn *organizations.Client, id string) (*awstypes.Account, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.AccountStatusPendingClosure, awstypes.AccountStatusActive),
		Target:       []string{},
		Refresh:      statusAccountStatus(ctx, conn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Account); ok {
		return output, err
	}

	return nil, err
}
