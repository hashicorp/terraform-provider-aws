// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_account_subscription", name="Account Subscription")
func resourceAccountSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountSubscriptionCreate,
		ReadWithoutTimeout:   resourceAccountSubscriptionRead,
		DeleteWithoutTimeout: resourceAccountSubscriptionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"account_name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"account_subscription_status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"active_directory_name": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"admin_group": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
					ForceNew: true,
				},
				"authentication_method": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.AuthenticationMethodOption](),
				},
				"author_group": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
					ForceNew: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				"contact_number": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"directory_id": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"edition": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Edition](),
				},
				"email_address": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"first_name": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"iam_identity_center_instance_arn": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"last_name": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"notification_email": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"reader_group": {
					Type:     schema.TypeList,
					Optional: true,
					ForceNew: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"realm": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
			}
		},
	}
}

func resourceAccountSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	accountName := d.Get("account_name").(string)
	input := &quicksight.CreateAccountSubscriptionInput{
		AccountName:          aws.String(accountName),
		AwsAccountId:         aws.String(awsAccountID),
		AuthenticationMethod: awstypes.AuthenticationMethodOption(d.Get("authentication_method").(string)),
		Edition:              awstypes.Edition(d.Get("edition").(string)),
		NotificationEmail:    aws.String(d.Get("notification_email").(string)),
	}

	if v, ok := d.GetOk("active_directory_name"); ok {
		input.ActiveDirectoryName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("admin_group"); ok && len(v.([]any)) > 0 {
		input.AdminGroup = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("author_group"); ok && len(v.([]any)) > 0 {
		input.AuthorGroup = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("reader_group"); ok && len(v.([]any)) > 0 {
		input.ReaderGroup = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("contact_number"); ok {
		input.ContactNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("directory_id"); ok {
		input.DirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_address"); ok {
		input.EmailAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("first_name"); ok {
		input.FirstName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_identity_center_instance_arn"); ok {
		input.IAMIdentityCenterInstanceArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("last_name"); ok {
		input.LastName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("realm"); ok {
		input.Realm = aws.String(v.(string))
	}

	_, err := conn.CreateAccountSubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Account Subscription (%s): %s", accountName, err)
	}

	d.SetId(awsAccountID)

	if _, err := waitAccountSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Account Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAccountSubscriptionRead(ctx, d, meta)...)
}

func resourceAccountSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	out, err := findAccountSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Account Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Account Subscription (%s): %s", d.Id(), err)
	}

	d.Set("account_name", out.AccountName)
	d.Set("account_subscription_status", out.AccountSubscriptionStatus)
	d.Set("edition", out.Edition)
	d.Set("iam_identity_center_instance_arn", out.IAMIdentityCenterInstanceArn)
	d.Set("notification_email", out.NotificationEmail)

	return diags
}

func resourceAccountSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	log.Printf("[INFO] Deleting QuickSight Account Subscription: %s", d.Id())
	_, err := conn.DeleteAccountSubscription(ctx, &quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Account Subscription (%s): %s", d.Id(), err)
	}

	if _, err := waitAccountSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Account Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}

// Not documented on AWS
const (
	accountSubscriptionStatusCreated                 = "ACCOUNT_CREATED"
	accountSubscriptionStatusOK                      = "OK"
	accountSubscriptionStatusSignupAttemptInProgress = "SIGNUP_ATTEMPT_IN_PROGRESS"
	accountSubscriptionStatusUnsuscribeInProgress    = "UNSUBSCRIBE_IN_PROGRESS"
	accountSubscriptionStatusUnsuscribed             = "UNSUBSCRIBED"
)

func waitAccountSubscriptionCreated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*awstypes.AccountInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{accountSubscriptionStatusSignupAttemptInProgress},
		Target:  []string{accountSubscriptionStatusCreated, accountSubscriptionStatusOK},
		Refresh: statusAccountSubscription(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccountInfo); ok {
		return output, err
	}

	return nil, err
}

func waitAccountSubscriptionDeleted(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*awstypes.AccountInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{accountSubscriptionStatusCreated, accountSubscriptionStatusOK, accountSubscriptionStatusUnsuscribeInProgress},
		Target:  []string{},
		Refresh: statusAccountSubscription(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccountInfo); ok {
		return output, err
	}

	return nil, err
}

func statusAccountSubscription(ctx context.Context, conn *quicksight.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findAccountSubscriptionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.AccountSubscriptionStatus), nil
	}
}

func findAccountSubscriptionByID(ctx context.Context, conn *quicksight.Client, id string) (*awstypes.AccountInfo, error) {
	input := &quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: aws.String(id),
	}

	output, err := findAccountSubscription(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.AccountSubscriptionStatus); status == accountSubscriptionStatusUnsuscribed {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findAccountSubscription(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeAccountSubscriptionInput) (*awstypes.AccountInfo, error) {
	output, err := conn.DescribeAccountSubscription(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccountInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccountInfo, nil
}
