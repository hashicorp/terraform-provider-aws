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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"country_code": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.PhoneNumberCountryCode_Values(), false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 500),
			},
			"phone_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validPhoneNumberPrefix,
			},
			"status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"target_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"type": {
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
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	targetArn := d.Get("target_arn").(string)
	phoneNumberType := d.Get("type").(string)
	input := &connect.SearchAvailablePhoneNumbersInput{
		MaxResults:             aws.Int64(1),
		PhoneNumberCountryCode: aws.String(d.Get("country_code").(string)),
		PhoneNumberType:        aws.String(phoneNumberType),
		TargetArn:              aws.String(targetArn),
	}

	if v, ok := d.GetOk("prefix"); ok {
		input.PhoneNumberPrefix = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Searching for Connect Available Phone Numbers %s", input)
	output, err := conn.SearchAvailablePhoneNumbersWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("searching Connect Phone Number for Connect Instance (%s,%s): %s", targetArn, phoneNumberType, err)
	}

	if output == nil || output.AvailableNumbersList == nil || len(output.AvailableNumbersList) == 0 {
		return diag.Errorf("searching Connect Phone Number for Connect Instance (%s,%s): empty output", targetArn, phoneNumberType)
	}

	phoneNumber := output.AvailableNumbersList[0].PhoneNumber

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("generating uuid for ClientToken for Connect Instance (%s,%s): %s", targetArn, aws.StringValue(phoneNumber), err)
	}

	input2 := &connect.ClaimPhoneNumberInput{
		ClientToken: aws.String(uuid), // can't use aws.String(id.UniqueId()), because it's not a valid uuid
		PhoneNumber: phoneNumber,
		Tags:        getTagsIn(ctx),
		TargetArn:   aws.String(targetArn),
	}

	if v, ok := d.GetOk("description"); ok {
		input2.PhoneNumberDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Claiming Connect Phone Number %s", input2)
	output2, err2 := conn.ClaimPhoneNumberWithContext(ctx, input2)

	if err2 != nil {
		return diag.Errorf("creating Connect Phone Number for Connect Instance (%s,%s): %s", targetArn, aws.StringValue(phoneNumber), err2)
	}

	if output2 == nil || output2.PhoneNumberId == nil {
		return diag.Errorf("creating Connect Phone Number for Connect Instance (%s,%s): empty output", targetArn, aws.StringValue(phoneNumber))
	}

	phoneNumberId := output2.PhoneNumberId
	d.SetId(aws.StringValue(phoneNumberId))

	if _, err := waitPhoneNumberCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return diag.Errorf("waiting for Phone Number (%s) creation: %s", d.Id(), err)
	}

	return resourcePhoneNumberRead(ctx, d, meta)
}

func resourcePhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	phoneNumberId := d.Id()

	resp, err := conn.DescribePhoneNumberWithContext(ctx, &connect.DescribePhoneNumberInput{
		PhoneNumberId: aws.String(phoneNumberId),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Phone Number (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Connect Phone Number (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.ClaimedPhoneNumberSummary == nil {
		return diag.Errorf("getting Connect Phone Number (%s): empty response", d.Id())
	}

	phoneNumberSummary := resp.ClaimedPhoneNumberSummary

	d.Set("arn", phoneNumberSummary.PhoneNumberArn)
	d.Set("country_code", phoneNumberSummary.PhoneNumberCountryCode)
	d.Set("description", phoneNumberSummary.PhoneNumberDescription)
	d.Set("phone_number", phoneNumberSummary.PhoneNumber)
	d.Set("type", phoneNumberSummary.PhoneNumberType)
	d.Set("target_arn", phoneNumberSummary.TargetArn)

	if err := d.Set("status", flattenPhoneNumberStatus(phoneNumberSummary.PhoneNumberStatus)); err != nil {
		return diag.Errorf("setting status: %s", err)
	}

	setTagsOut(ctx, resp.ClaimedPhoneNumberSummary.Tags)

	return nil
}

func resourcePhoneNumberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	phoneNumberId := d.Id()

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("generating uuid for ClientToken for Phone Number %s: %s", phoneNumberId, err)
	}

	if d.HasChange("target_arn") {
		_, err := conn.UpdatePhoneNumberWithContext(ctx, &connect.UpdatePhoneNumberInput{
			ClientToken:   aws.String(uuid),
			PhoneNumberId: aws.String(phoneNumberId),
			TargetArn:     aws.String(d.Get("target_arn").(string)),
		})

		if err != nil {
			return diag.Errorf("updating Phone Number (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitPhoneNumberUpdated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return diag.Errorf("waiting for Phone Number (%s) update: %s", d.Id(), err)
	}

	return resourcePhoneNumberRead(ctx, d, meta)
}

func resourcePhoneNumberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	phoneNumberId := d.Id()

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("generating uuid for ClientToken for Phone Number %s: %s", phoneNumberId, err)
	}

	_, err = conn.ReleasePhoneNumberWithContext(ctx, &connect.ReleasePhoneNumberInput{
		ClientToken:   aws.String(uuid),
		PhoneNumberId: aws.String(phoneNumberId),
	})

	if err != nil {
		return diag.Errorf("deleting PhoneNumber (%s): %s", d.Id(), err)
	}

	if _, err := waitPhoneNumberDeleted(ctx, conn, d.Timeout(schema.TimeoutCreate), phoneNumberId); err != nil {
		return diag.Errorf("waiting for Phone Number (%s) deletion: %s", phoneNumberId, err)
	}

	return nil
}

func flattenPhoneNumberStatus(apiObject *connect.PhoneNumberStatus) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"message": aws.StringValue(apiObject.Message),
		"status":  aws.StringValue(apiObject.Status),
	}

	return []interface{}{values}
}
