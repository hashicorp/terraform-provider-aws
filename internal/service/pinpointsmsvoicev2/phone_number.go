// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpointsmsvoicev2_phone_number", name="Phone Number")
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

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"iso_country_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"message_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MessageType](),
			},
			"monthly_leasing_price": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_capabilities": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.NumberCapability](),
				},
			},
			"number_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RequestableNumberType](),
			},
			"opt_out_list_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Default",
			},
			"phone_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registration_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"self_managed_opt_outs_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"two_way_channel_arn": {
				Type:     schema.TypeString,
				Optional: true,
				RequiredWith: []string{
					"two_way_channel_enabled",
				},
			},
			"two_way_channel_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				RequiredWith: []string{
					"two_way_channel_arn",
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePhoneNumberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.RequestPhoneNumberInput{
		IsoCountryCode:     aws.String(d.Get("iso_country_code").(string)),
		MessageType:        awstypes.MessageType(d.Get("message_type").(string)),
		NumberCapabilities: flex.ExpandStringyValueSet[awstypes.NumberCapability](d.Get("number_capabilities").(*schema.Set)),
		NumberType:         awstypes.RequestableNumberType(d.Get("number_type").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("deletion_protection_enabled"); ok {
		input.DeletionProtectionEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("opt_out_list_name"); ok {
		input.OptOutListName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("registration_id"); ok {
		input.RegistrationId = aws.String(v.(string))
	}

	output, err := conn.RequestPhoneNumber(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting End User Messaging SMS Phone Number: %s", err)
	}

	d.SetId(aws.ToString(output.PhoneNumberId))

	if _, err := waitPhoneNumberActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for End User Messaging SMS Phone Number (%s) create: %s", d.Id(), err)
	}

	if sdkv2.HasNonZeroValues(d, "self_managed_opt_outs_enabled", "two_way_channel_arn", "two_way_channel_enabled") {
		input := &pinpointsmsvoicev2.UpdatePhoneNumberInput{
			PhoneNumberId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("self_managed_opt_outs_enabled"); ok {
			input.SelfManagedOptOutsEnabled = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("two_way_channel_arn"); ok {
			input.TwoWayChannelArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("two_way_channel_enabled"); ok {
			input.TwoWayEnabled = aws.Bool(v.(bool))
		}

		_, err := conn.UpdatePhoneNumber(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating End User Messaging SMS Phone Number (%s): %s", d.Id(), err)
		}

		if _, err := waitPhoneNumberActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for End User Messaging SMS Phone Number (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePhoneNumberRead(ctx, d, meta)...)
}

func resourcePhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	out, err := findPhoneNumberByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] End User Messaging SMS Phone Number (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading End User Messaging SMS Phone Number (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.PhoneNumberArn)
	d.Set("deletion_protection_enabled", out.DeletionProtectionEnabled)
	d.Set("iso_country_code", out.IsoCountryCode)
	d.Set("message_type", out.MessageType)
	d.Set("monthly_leasing_price", out.MonthlyLeasingPrice)
	d.Set("number_capabilities", out.NumberCapabilities)
	d.Set("number_type", out.NumberType)
	d.Set("opt_out_list_name", out.OptOutListName)
	d.Set("phone_number", out.PhoneNumber)
	d.Set("self_managed_opt_outs_enabled", out.SelfManagedOptOutsEnabled)
	d.Set("two_way_channel_arn", out.TwoWayChannelArn)
	d.Set("two_way_channel_enabled", out.TwoWayEnabled)

	return diags
}

func resourcePhoneNumberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &pinpointsmsvoicev2.UpdatePhoneNumberInput{
			PhoneNumberId: aws.String(d.Id()),
		}

		if d.HasChanges("deletion_protection_enabled") {
			input.DeletionProtectionEnabled = aws.Bool(d.Get("deletion_protection_enabled").(bool))
		}

		if d.HasChanges("opt_out_list_name") {
			input.OptOutListName = aws.String(d.Get("opt_out_list_name").(string))
		}

		if d.HasChanges("self_managed_opt_outs_enabled") {
			input.SelfManagedOptOutsEnabled = aws.Bool(d.Get("self_managed_opt_outs_enabled").(bool))
		}

		if d.HasChanges("two_way_channel_arn") {
			input.TwoWayChannelArn = aws.String(d.Get("two_way_channel_arn").(string))
		}

		if d.HasChanges("two_way_channel_enabled") {
			input.TwoWayEnabled = aws.Bool(d.Get("two_way_channel_enabled").(bool))
		}

		_, err := conn.UpdatePhoneNumber(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating End User Messaging SMS Phone Number (%s): %s", d.Id(), err)
		}

		if _, err := waitPhoneNumberActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for End User Messaging SMS Phone Number (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePhoneNumberRead(ctx, d, meta)...)
}

func resourcePhoneNumberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	log.Printf("[INFO] Deleting End User Messaging SMS Phone Number: %s", d.Id())
	_, err := conn.ReleasePhoneNumber(ctx, &pinpointsmsvoicev2.ReleasePhoneNumberInput{
		PhoneNumberId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "releasing End User Messaging SMS Phone Number (%s): %s", d.Id(), err)
	}

	if _, err := waitPhoneNumberDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for End User Messaging SMS Phone Number (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findPhoneNumberByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.PhoneNumberInformation, error) {
	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
		PhoneNumberIds: []string{id},
	}

	output, err := findPhoneNumber(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.NumberStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findPhoneNumber(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribePhoneNumbersInput) (*awstypes.PhoneNumberInformation, error) {
	output, err := findPhoneNumbers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPhoneNumbers(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribePhoneNumbersInput) ([]awstypes.PhoneNumberInformation, error) {
	var output []awstypes.PhoneNumberInformation

	pages := pinpointsmsvoicev2.NewDescribePhoneNumbersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PhoneNumbers...)
	}

	return output, nil
}

func statusPhoneNumber(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPhoneNumberByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPhoneNumberActive(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, timeout time.Duration) (*awstypes.PhoneNumberInformation, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NumberStatusPending, awstypes.NumberStatusAssociating),
		Target:  enum.Slice(awstypes.NumberStatusActive),
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PhoneNumberInformation); ok {
		return output, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, timeout time.Duration) (*awstypes.PhoneNumberInformation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NumberStatusDisassociating),
		Target:  []string{},
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PhoneNumberInformation); ok {
		return output, err
	}

	return nil, err
}
