// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/go-uuid"
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

// @SDKResource("aws_connect_phone_number", name="Phone Number")
// @Tags(identifierAttribute="arn")
func resourcePhoneNumber() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePhoneNumberCreate,
		ReadWithoutTimeout:   resourcePhoneNumberRead,
		UpdateWithoutTimeout: resourcePhoneNumberUpdate,
		DeleteWithoutTimeout: resourcePhoneNumberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"country_code": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PhoneNumberCountryCode](),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 500),
			},
			"phone_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPrefix: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validPhoneNumberPrefix,
			},
			names.AttrStatus: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTargetARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PhoneNumberType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePhoneNumberCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	var phoneNumber string
	targetARN := d.Get(names.AttrTargetARN).(string)

	{
		phoneNumberType := d.Get(names.AttrType).(string)
		input := &connect.SearchAvailablePhoneNumbersInput{
			MaxResults:             aws.Int32(1),
			PhoneNumberCountryCode: awstypes.PhoneNumberCountryCode(d.Get("country_code").(string)),
			PhoneNumberType:        awstypes.PhoneNumberType(phoneNumberType),
			TargetArn:              aws.String(targetARN),
		}

		if v, ok := d.GetOk(names.AttrPrefix); ok {
			input.PhoneNumberPrefix = aws.String(v.(string))
		}

		output, err := conn.SearchAvailablePhoneNumbers(ctx, input)

		if err == nil && (output == nil || len(output.AvailableNumbersList) == 0) {
			err = tfresource.NewEmptyResultError(input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "searching Connect Phone Numbers (%s,%s): %s", targetARN, phoneNumberType, err)
		}

		phoneNumber = aws.ToString(output.AvailableNumbersList[0].PhoneNumber)
	}

	{
		uuid, err := uuid.GenerateUUID()
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &connect.ClaimPhoneNumberInput{
			ClientToken: aws.String(uuid), // can't use aws.String(id.UniqueId()), because it's not a valid uuid
			PhoneNumber: aws.String(phoneNumber),
			Tags:        getTagsIn(ctx),
			TargetArn:   aws.String(targetARN),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.PhoneNumberDescription = aws.String(v.(string))
		}

		output, err := conn.ClaimPhoneNumber(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "claiming Connect Phone Number (%s,%s): %s", targetARN, phoneNumber, err)
		}

		d.SetId(aws.ToString(output.PhoneNumberId))
	}

	if _, err := waitPhoneNumberCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Connect Phone Number (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePhoneNumberRead(ctx, d, meta)...)
}

func resourcePhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	phoneNumberSummary, err := findPhoneNumberByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Phone Number (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Phone Number (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, phoneNumberSummary.PhoneNumberArn)
	d.Set("country_code", phoneNumberSummary.PhoneNumberCountryCode)
	d.Set(names.AttrDescription, phoneNumberSummary.PhoneNumberDescription)
	d.Set("phone_number", phoneNumberSummary.PhoneNumber)
	if err := d.Set(names.AttrStatus, flattenPhoneNumberStatus(phoneNumberSummary.PhoneNumberStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting status: %s", err)
	}
	d.Set(names.AttrTargetARN, phoneNumberSummary.TargetArn)
	d.Set(names.AttrType, phoneNumberSummary.PhoneNumberType)

	setTagsOut(ctx, phoneNumberSummary.Tags)

	return diags
}

func resourcePhoneNumberUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		uuid, err := uuid.GenerateUUID()
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &connect.UpdatePhoneNumberInput{
			ClientToken:   aws.String(uuid),
			PhoneNumberId: aws.String(d.Id()),
			TargetArn:     aws.String(d.Get(names.AttrTargetARN).(string)),
		}

		_, err = conn.UpdatePhoneNumber(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Phone Number (%s): %s", d.Id(), err)
		}

		if _, err := waitPhoneNumberUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Phone Number (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePhoneNumberRead(ctx, d, meta)...)
}

func resourcePhoneNumberDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Phone Number: %s", d.Id())
	input := connect.ReleasePhoneNumberInput{
		ClientToken:   aws.String(uuid),
		PhoneNumberId: aws.String(d.Id()),
	}
	_, err = conn.ReleasePhoneNumber(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "releasing Connect Phone Number (%s): %s", d.Id(), err)
	}

	if _, err := waitPhoneNumberDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Phone Number (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func findPhoneNumberByID(ctx context.Context, conn *connect.Client, id string) (*awstypes.ClaimedPhoneNumberSummary, error) {
	input := &connect.DescribePhoneNumberInput{
		PhoneNumberId: aws.String(id),
	}

	return findPhoneNumber(ctx, conn, input)
}

func findPhoneNumber(ctx context.Context, conn *connect.Client, input *connect.DescribePhoneNumberInput) (*awstypes.ClaimedPhoneNumberSummary, error) {
	output, err := conn.DescribePhoneNumber(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ClaimedPhoneNumberSummary == nil || output.ClaimedPhoneNumberSummary.PhoneNumberStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ClaimedPhoneNumberSummary, nil
}

func flattenPhoneNumberStatus(apiObject *awstypes.PhoneNumberStatus) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrMessage: aws.ToString(apiObject.Message),
		names.AttrStatus:  apiObject.Status,
	}

	return []any{tfMap}
}

func statusPhoneNumber(ctx context.Context, conn *connect.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPhoneNumberByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.PhoneNumberStatus.Status), nil
	}
}

func waitPhoneNumberCreated(ctx context.Context, conn *connect.Client, id string, timeout time.Duration) (*awstypes.ClaimedPhoneNumberSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PhoneNumberWorkflowStatusInProgress),
		Target:  enum.Slice(awstypes.PhoneNumberWorkflowStatusClaimed),
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClaimedPhoneNumberSummary); ok {
		if status := output.PhoneNumberStatus; status.Status == awstypes.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(status.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitPhoneNumberUpdated(ctx context.Context, conn *connect.Client, id string, timeout time.Duration) (*awstypes.ClaimedPhoneNumberSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PhoneNumberWorkflowStatusInProgress),
		Target:  enum.Slice(awstypes.PhoneNumberWorkflowStatusClaimed),
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClaimedPhoneNumberSummary); ok {
		if status := output.PhoneNumberStatus; status.Status == awstypes.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(status.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *connect.Client, id string, timeout time.Duration) (*awstypes.ClaimedPhoneNumberSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PhoneNumberWorkflowStatusInProgress),
		Target:  []string{},
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClaimedPhoneNumberSummary); ok {
		if status := output.PhoneNumberStatus; status.Status == awstypes.PhoneNumberWorkflowStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(status.Message)))
		}

		return output, err
	}

	return nil, err
}
