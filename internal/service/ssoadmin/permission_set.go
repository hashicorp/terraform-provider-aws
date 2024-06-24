// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
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

// @SDKResource("aws_ssoadmin_permission_set", name="Permission Set")
// @Tags
func ResourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionSetCreate,
		ReadWithoutTimeout:   resourcePermissionSetRead,
		UpdateWithoutTimeout: resourcePermissionSetUpdate,
		DeleteWithoutTimeout: resourcePermissionSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 700),
					validation.StringMatch(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "must match [\\p{L}\\p{M}\\p{Z}\\p{S}\\p{N}\\p{P}]"),
				),
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexache.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
			},
			"relay_state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 240),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z&$@#\\\/%?=~\-_'"|!:,.;*+\[\]\ \(\)\{\}]+`), "must match [0-9A-Za-z&$@#\\\\\\/%?=~\\-_'\"|!:,.;*+\\[\\]\\(\\)\\{\\}]"),
				),
			},
			"session_duration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Default:      "PT1H",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePermissionSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	instanceARN := d.Get("instance_arn").(string)
	name := d.Get(names.AttrName).(string)
	input := &ssoadmin.CreatePermissionSetInput{
		InstanceArn: aws.String(instanceARN),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("relay_state"); ok {
		input.RelayState = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_duration"); ok {
		input.SessionDuration = aws.String(v.(string))
	}

	output, err := conn.CreatePermissionSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Permission Set (%s): %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.ToString(output.PermissionSet.PermissionSetArn), instanceARN))

	return append(diags, resourcePermissionSetRead(ctx, d, meta)...)
}

func resourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	permissionSetARN, instanceARN, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	permissionSet, err := FindPermissionSet(ctx, conn, permissionSetARN, instanceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Permission Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permission Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, permissionSet.PermissionSetArn)
	d.Set(names.AttrCreatedDate, permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, permissionSet.Description)
	d.Set("instance_arn", instanceARN)
	d.Set(names.AttrName, permissionSet.Name)
	d.Set("relay_state", permissionSet.RelayState)
	d.Set("session_duration", permissionSet.SessionDuration)

	tags, err := listTags(ctx, conn, permissionSetARN, instanceARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SSO Permission Set (%s): %s", permissionSetARN, err)
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourcePermissionSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	permissionSetARN, instanceARN, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges(names.AttrDescription, "relay_state", "session_duration") {
		input := &ssoadmin.UpdatePermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
		}

		// The AWS SSO API requires we send the RelayState value regardless if it's unchanged
		// else the existing Permission Set's RelayState value will be cleared;
		// for consistency, we'll check for the "presence of" instead of "if changed" for all input fields
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("relay_state"); ok {
			input.RelayState = aws.String(v.(string))
		}

		if v, ok := d.GetOk("session_duration"); ok {
			input.SessionDuration = aws.String(v.(string))
		}

		_, err := conn.UpdatePermissionSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSO Permission Set (%s): %s", d.Id(), err)
		}

		// Re-provision ALL accounts after making the above changes
		if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)
		if err := updateTags(ctx, conn, permissionSetARN, instanceARN, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourcePermissionSetRead(ctx, d, meta)...)
}

func resourcePermissionSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	permissionSetARN, instanceARN, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting SSO Permission Set: %s", d.Id())
	_, err = conn.DeletePermissionSet(ctx, &ssoadmin.DeletePermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Permission Set (%s): %s", permissionSetARN, err)
	}

	return diags
}

func ParseResourceID(id string) (string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%q), expected PERMISSION_SET_ARN,INSTANCE_ARN", id)
	}
	return idParts[0], idParts[1], nil
}

func FindPermissionSet(ctx context.Context, conn *ssoadmin.Client, permissionSetARN, instanceARN string) (*awstypes.PermissionSet, error) {
	input := &ssoadmin.DescribePermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	output, err := conn.DescribePermissionSet(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PermissionSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PermissionSet, nil
}

func provisionPermissionSet(ctx context.Context, conn *ssoadmin.Client, permissionSetARN, instanceARN string, timeout time.Duration) error {
	input := &ssoadmin.ProvisionPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
		TargetType:       awstypes.ProvisionTargetTypeAllProvisionedAccounts,
	}

	output, err := conn.ProvisionPermissionSet(ctx, input)

	if err != nil {
		return fmt.Errorf("provisioning SSO Permission Set (%s): %w", permissionSetARN, err)
	}

	if _, err := waitPermissionSetProvisioned(ctx, conn, instanceARN, aws.ToString(output.PermissionSetProvisioningStatus.RequestId), timeout); err != nil {
		return fmt.Errorf("waiting for SSO Permission Set (%s) provision: %w", permissionSetARN, err)
	}

	return nil
}

func findPermissionSetProvisioningStatus(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string) (*awstypes.PermissionSetProvisioningStatus, error) {
	input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
		InstanceArn:                     aws.String(instanceARN),
		ProvisionPermissionSetRequestId: aws.String(requestID),
	}

	output, err := conn.DescribePermissionSetProvisioningStatus(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PermissionSetProvisioningStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PermissionSetProvisioningStatus, nil
}

func statusPermissionSetProvisioning(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPermissionSetProvisioningStatus(ctx, conn, instanceARN, requestID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPermissionSetProvisioned(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string, timeout time.Duration) (*awstypes.PermissionSetProvisioningStatus, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusValuesInProgress),
		Target:  enum.Slice(awstypes.StatusValuesSucceeded),
		Refresh: statusPermissionSetProvisioning(ctx, conn, instanceARN, requestID),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PermissionSetProvisioningStatus); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))

		return output, err
	}

	return nil, err
}
