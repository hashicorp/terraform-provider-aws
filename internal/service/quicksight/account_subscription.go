// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_quicksight_account_subscription", name="Account Subscription")
func ResourceAccountSubscription() *schema.Resource {
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
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(quicksight.AuthenticationMethodOption_Values(), false),
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
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(quicksight.Edition_Values(), false),
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

const (
	ResNameAccountSubscription = "Account Subscription"
)

func resourceAccountSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}

	in := &quicksight.CreateAccountSubscriptionInput{
		AwsAccountId:         aws.String(awsAccountId),
		AccountName:          aws.String(d.Get("account_name").(string)),
		AuthenticationMethod: aws.String(d.Get("authentication_method").(string)),
		Edition:              aws.String(d.Get("edition").(string)),
		NotificationEmail:    aws.String(d.Get("notification_email").(string)),
	}

	if v, ok := d.GetOk("active_directory_name"); ok {
		in.ActiveDirectoryName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("admin_group"); ok && len(v.([]interface{})) > 0 {
		in.AdminGroup = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("author_group"); ok && len(v.([]interface{})) > 0 {
		in.AuthorGroup = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("reader_group"); ok && len(v.([]interface{})) > 0 {
		in.ReaderGroup = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("contact_number"); ok {
		in.ContactNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("directory_id"); ok {
		in.DirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_address"); ok {
		in.EmailAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("first_name"); ok {
		in.FirstName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_identity_center_instance_arn"); ok {
		in.IAMIdentityCenterInstanceArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("last_name"); ok {
		in.LastName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("realm"); ok {
		in.Realm = aws.String(v.(string))
	}

	out, err := conn.CreateAccountSubscriptionWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionCreating, ResNameAccountSubscription, d.Get("account_name").(string), err)
	}

	if out == nil || out.SignupResponse == nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionCreating, ResNameAccountSubscription, d.Get("account_name").(string), errors.New("empty output"))
	}

	d.SetId(awsAccountId)

	if _, err := waitAccountSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForCreation, ResNameAccountSubscription, d.Id(), err)
	}

	return append(diags, resourceAccountSubscriptionRead(ctx, d, meta)...)
}

func resourceAccountSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	out, err := FindAccountSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight AccountSubscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	// Ressource is logically deleted with UNSUBSCRIBED status
	if !d.IsNewResource() && aws.StringValue(out.AccountSubscriptionStatus) == statusUnsuscribed {
		log.Printf("[WARN] QuickSight AccountSubscription (%s) unsuscribed, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionReading, ResNameAccountSubscription, d.Id(), err)
	}

	d.Set("account_name", out.AccountName)
	d.Set("edition", out.Edition)
	d.Set("notification_email", out.NotificationEmail)
	d.Set("account_subscription_status", out.AccountSubscriptionStatus)
	d.Set("iam_identity_center_instance_arn", out.IAMIdentityCenterInstanceArn)

	return diags
}

func resourceAccountSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	log.Printf("[INFO] Deleting QuickSight AccountSubscription %s", d.Id())

	_, err := conn.DeleteAccountSubscriptionWithContext(ctx, &quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionDeleting, ResNameAccountSubscription, d.Id(), err)
	}

	if _, err := waitAccountSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForDeletion, ResNameAccountSubscription, d.Id(), err)
	}

	return diags
}

// Not documented on AWS
const (
	statusCreated                 = "ACCOUNT_CREATED"
	statusOk                      = "OK"
	statusSignupAttemptInProgress = "SIGNUP_ATTEMPT_IN_PROGRESS"
	statusUnsuscribeInProgress    = "UNSUBSCRIBE_IN_PROGRESS"
	statusUnsuscribed             = "UNSUBSCRIBED"
)

func waitAccountSubscriptionCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.AccountInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusSignupAttemptInProgress},
		Target:                    []string{statusCreated, statusOk},
		Refresh:                   statusAccountSubscription(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.AccountInfo); ok {
		return out, err
	}

	return nil, err
}

func waitAccountSubscriptionDeleted(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.AccountInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusCreated, statusOk, statusUnsuscribeInProgress},
		Target:  []string{statusUnsuscribed},
		Refresh: statusAccountSubscription(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.AccountInfo); ok {
		return out, err
	}

	return nil, err
}

func statusAccountSubscription(ctx context.Context, conn *quicksight.QuickSight, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindAccountSubscriptionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, *out.AccountSubscriptionStatus, nil
	}
}

func FindAccountSubscriptionByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.AccountInfo, error) {
	in := &quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: aws.String(id),
	}
	out, err := conn.DescribeAccountSubscriptionWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.AccountInfo.AccountName == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AccountInfo, nil
}
