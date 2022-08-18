package connect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePhoneNumber() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourcePhoneNumberRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	tags := KeyValueTags(phoneNumberSummary.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
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
