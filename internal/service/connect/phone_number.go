// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_phone_number", name="Phone Number")
// @Tags(identifierAttribute="arn")
func ResourcePhoneNumber() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePhoneNumberCreate,
		ReadWithoutTimeout:   resourcePhoneNumberRead,
		UpdateWithoutTimeout: resourcePhoneNumberUpdate,
		DeleteWithoutTimeout: resourcePhoneNumberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(phoneNumberCreatedTimeout),
			Update: schema.DefaultTimeout(phoneNumberUpdatedTimeout),
			Delete: schema.DefaultTimeout(phoneNumberDeletedTimeout),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"country_code": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.PhoneNumberCountryCode_Values(), false),
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.PhoneNumberType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePhoneNumberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	targetArn := d.Get(names.AttrTargetARN).(string)
	phoneNumberType := d.Get(names.AttrType).(string)
	input := &connect.SearchAvailablePhoneNumbersInput{
		MaxResults:             aws.Int64(1),
		PhoneNumberCountryCode: aws.String(d.Get("country_code").(string)),
		PhoneNumberType:        aws.String(phoneNumberType),
		TargetArn:              aws.String(targetArn),
	}

	if v, ok := d.GetOk(names.AttrPrefix); ok {
		input.PhoneNumberPrefix = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Searching for Connect Available Phone Numbers %s", input)
	output, err := conn.SearchAvailablePhoneNumbersWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "searching Connect Phone Number for Connect Instance (%s,%s): %s", targetArn, phoneNumberType, err)
	}

	if output == nil || output.AvailableNumbersList == nil || len(output.AvailableNumbersList) == 0 {
		return sdkdiag.AppendErrorf(diags, "searching Connect Phone Number for Connect Instance (%s,%s): empty output", targetArn, phoneNumberType)
	}

	phoneNumber := output.AvailableNumbersList[0].PhoneNumber

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "generating uuid for ClientToken for Connect Instance (%s,%s): %s", targetArn, aws.StringValue(phoneNumber), err)
	}

	input2 := &connect.ClaimPhoneNumberInput{
		ClientToken: aws.String(uuid), // can't use aws.String(id.UniqueId()), because it's not a valid uuid
		PhoneNumber: phoneNumber,
		Tags:        getTagsIn(ctx),
		TargetArn:   aws.String(targetArn),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input2.PhoneNumberDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Claiming Connect Phone Number %s", input2)
	output2, err2 := conn.ClaimPhoneNumberWithContext(ctx, input2)

	if err2 != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Phone Number for Connect Instance (%s,%s): %s", targetArn, aws.StringValue(phoneNumber), err2)
	}

	if output2 == nil || output2.PhoneNumberId == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Phone Number for Connect Instance (%s,%s): empty output", targetArn, aws.StringValue(phoneNumber))
	}

	phoneNumberId := output2.PhoneNumberId
	d.SetId(aws.StringValue(phoneNumberId))

	if _, err := waitPhoneNumberCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Phone Number (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourcePhoneNumberRead(ctx, d, meta)...)
}

func resourcePhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	phoneNumberId := d.Id()

	resp, err := conn.DescribePhoneNumberWithContext(ctx, &connect.DescribePhoneNumberInput{
		PhoneNumberId: aws.String(phoneNumberId),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Phone Number (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Phone Number (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.ClaimedPhoneNumberSummary == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Phone Number (%s): empty response", d.Id())
	}

	phoneNumberSummary := resp.ClaimedPhoneNumberSummary

	d.Set(names.AttrARN, phoneNumberSummary.PhoneNumberArn)
	d.Set("country_code", phoneNumberSummary.PhoneNumberCountryCode)
	d.Set(names.AttrDescription, phoneNumberSummary.PhoneNumberDescription)
	d.Set("phone_number", phoneNumberSummary.PhoneNumber)
	d.Set(names.AttrType, phoneNumberSummary.PhoneNumberType)
	d.Set(names.AttrTargetARN, phoneNumberSummary.TargetArn)

	if err := d.Set(names.AttrStatus, flattenPhoneNumberStatus(phoneNumberSummary.PhoneNumberStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting status: %s", err)
	}

	setTagsOut(ctx, resp.ClaimedPhoneNumberSummary.Tags)

	return diags
}

func resourcePhoneNumberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	phoneNumberId := d.Id()

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "generating uuid for ClientToken for Phone Number %s: %s", phoneNumberId, err)
	}

	if d.HasChange(names.AttrTargetARN) {
		_, err := conn.UpdatePhoneNumberWithContext(ctx, &connect.UpdatePhoneNumberInput{
			ClientToken:   aws.String(uuid),
			PhoneNumberId: aws.String(phoneNumberId),
			TargetArn:     aws.String(d.Get(names.AttrTargetARN).(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Phone Number (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitPhoneNumberUpdated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Phone Number (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourcePhoneNumberRead(ctx, d, meta)...)
}

func resourcePhoneNumberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	phoneNumberId := d.Id()

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "generating uuid for ClientToken for Phone Number %s: %s", phoneNumberId, err)
	}

	_, err = conn.ReleasePhoneNumberWithContext(ctx, &connect.ReleasePhoneNumberInput{
		ClientToken:   aws.String(uuid),
		PhoneNumberId: aws.String(phoneNumberId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting PhoneNumber (%s): %s", d.Id(), err)
	}

	if _, err := waitPhoneNumberDeleted(ctx, conn, d.Timeout(schema.TimeoutCreate), phoneNumberId); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Phone Number (%s) deletion: %s", phoneNumberId, err)
	}

	return diags
}

func flattenPhoneNumberStatus(apiObject *connect.PhoneNumberStatus) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		names.AttrMessage: aws.StringValue(apiObject.Message),
		names.AttrStatus:  aws.StringValue(apiObject.Status),
	}

	return []interface{}{values}
}
