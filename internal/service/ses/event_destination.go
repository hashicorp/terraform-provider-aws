// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_event_destination")
func ResourceEventDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventDestinationCreate,
		ReadWithoutTimeout:   resourceEventDestinationRead,
		DeleteWithoutTimeout: resourceEventDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEventDestinationImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_destination": {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"kinesis_destination", "sns_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDefaultValue: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.@-]+$`), "must contain only alphanumeric, underscore, hyphen, period, and at signs characters"),
							),
						},
						"dimension_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_:-]+$`), "must contain only alphanumeric, underscore, and hyphen characters"),
							),
						},
						"value_source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ses.DimensionValueSource_Values(), false),
						},
					},
				},
			},
			"configuration_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"kinesis_destination": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				MaxItems:      1,
				ConflictsWith: []string{"cloudwatch_destination", "sns_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrStreamARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"matching_types": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ses.EventType_Values(), false),
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric, underscore, and hyphen characters"),
				),
			},
			"sns_destination": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cloudwatch_destination", "kinesis_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
		},
	}
}

func resourceEventDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	configurationSetName := d.Get("configuration_set_name").(string)
	eventDestinationName := d.Get(names.AttrName).(string)
	enabled := d.Get(names.AttrEnabled).(bool)
	matchingEventTypes := d.Get("matching_types").(*schema.Set)

	createOpts := &ses.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(configurationSetName),
		EventDestination: &ses.EventDestination{
			Name:               aws.String(eventDestinationName),
			Enabled:            aws.Bool(enabled),
			MatchingEventTypes: flex.ExpandStringSet(matchingEventTypes),
		},
	}

	if v, ok := d.GetOk("cloudwatch_destination"); ok {
		destination := v.(*schema.Set).List()
		createOpts.EventDestination.CloudWatchDestination = &ses.CloudWatchDestination{
			DimensionConfigurations: generateCloudWatchDestination(destination),
		}
		log.Printf("[DEBUG] Creating cloudwatch destination: %#v", destination)
	}

	if v, ok := d.GetOk("kinesis_destination"); ok {
		destination := v.([]interface{})

		kinesis := destination[0].(map[string]interface{})
		createOpts.EventDestination.KinesisFirehoseDestination = &ses.KinesisFirehoseDestination{
			DeliveryStreamARN: aws.String(kinesis[names.AttrStreamARN].(string)),
			IAMRoleARN:        aws.String(kinesis[names.AttrRoleARN].(string)),
		}
		log.Printf("[DEBUG] Creating kinesis destination: %#v", kinesis)
	}

	if v, ok := d.GetOk("sns_destination"); ok {
		destination := v.([]interface{})
		sns := destination[0].(map[string]interface{})
		createOpts.EventDestination.SNSDestination = &ses.SNSDestination{
			TopicARN: aws.String(sns[names.AttrTopicARN].(string)),
		}
		log.Printf("[DEBUG] Creating sns destination: %#v", sns)
	}

	_, err := conn.CreateConfigurationSetEventDestinationWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES configuration set event destination: %s", err)
	}

	d.SetId(eventDestinationName)

	log.Printf("[WARN] SES DONE")
	return append(diags, resourceEventDestinationRead(ctx, d, meta)...)
}

func resourceEventDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	configurationSetName := d.Get("configuration_set_name").(string)
	input := &ses.DescribeConfigurationSetInput{
		ConfigurationSetAttributeNames: aws.StringSlice([]string{ses.ConfigurationSetAttributeEventDestinations}),
		ConfigurationSetName:           aws.String(configurationSetName),
	}

	output, err := conn.DescribeConfigurationSetWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, ses.ErrCodeConfigurationSetDoesNotExistException) {
		log.Printf("[WARN] SES Configuration Set (%s) not found, removing from state", configurationSetName)
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Configuration Set Event Destination (%s): %s", d.Id(), err)
	}

	var thisEventDestination *ses.EventDestination
	for _, eventDestination := range output.EventDestinations {
		if aws.StringValue(eventDestination.Name) == d.Id() {
			thisEventDestination = eventDestination
			break
		}
	}
	if thisEventDestination == nil {
		log.Printf("[WARN] SES Configuration Set Event Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("configuration_set_name", output.ConfigurationSet.Name)
	d.Set(names.AttrEnabled, thisEventDestination.Enabled)
	d.Set(names.AttrName, thisEventDestination.Name)
	if err := d.Set("cloudwatch_destination", flattenCloudWatchDestination(thisEventDestination.CloudWatchDestination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_destination: %s", err)
	}
	if err := d.Set("kinesis_destination", flattenKinesisFirehoseDestination(thisEventDestination.KinesisFirehoseDestination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kinesis_destination: %s", err)
	}
	if err := d.Set("matching_types", flex.FlattenStringSet(thisEventDestination.MatchingEventTypes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting matching_types: %s", err)
	}
	if err := d.Set("sns_destination", flattenSNSDestination(thisEventDestination.SNSDestination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sns_destination: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("configuration-set/%s:event-destination/%s", configurationSetName, d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceEventDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	log.Printf("[DEBUG] SES Delete Configuration Set Destination: %s", d.Id())
	_, err := conn.DeleteConfigurationSetEventDestinationWithContext(ctx, &ses.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		EventDestinationName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Event Destination (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceEventDestinationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'configuration-set-name/event-destination-name'", d.Id())
	}

	configurationSetName := parts[0]
	eventDestinationName := parts[1]
	log.Printf("[DEBUG] Importing SES event destination %s from configuration set %s", eventDestinationName, configurationSetName)

	d.SetId(eventDestinationName)
	d.Set("configuration_set_name", configurationSetName)

	return []*schema.ResourceData{d}, nil
}

func generateCloudWatchDestination(v []interface{}) []*ses.CloudWatchDimensionConfiguration {
	b := make([]*ses.CloudWatchDimensionConfiguration, len(v))

	for i, vI := range v {
		cloudwatch := vI.(map[string]interface{})
		b[i] = &ses.CloudWatchDimensionConfiguration{
			DefaultDimensionValue: aws.String(cloudwatch[names.AttrDefaultValue].(string)),
			DimensionName:         aws.String(cloudwatch["dimension_name"].(string)),
			DimensionValueSource:  aws.String(cloudwatch["value_source"].(string)),
		}
	}

	return b
}

func flattenCloudWatchDestination(destination *ses.CloudWatchDestination) []interface{} {
	if destination == nil {
		return []interface{}{}
	}

	vDimensionConfigurations := []interface{}{}

	for _, dimensionConfiguration := range destination.DimensionConfigurations {
		mDimensionConfiguration := map[string]interface{}{
			names.AttrDefaultValue: aws.StringValue(dimensionConfiguration.DefaultDimensionValue),
			"dimension_name":       aws.StringValue(dimensionConfiguration.DimensionName),
			"value_source":         aws.StringValue(dimensionConfiguration.DimensionValueSource),
		}

		vDimensionConfigurations = append(vDimensionConfigurations, mDimensionConfiguration)
	}

	return vDimensionConfigurations
}

func flattenKinesisFirehoseDestination(destination *ses.KinesisFirehoseDestination) []interface{} {
	if destination == nil {
		return []interface{}{}
	}

	mDestination := map[string]interface{}{
		names.AttrRoleARN:   aws.StringValue(destination.IAMRoleARN),
		names.AttrStreamARN: aws.StringValue(destination.DeliveryStreamARN),
	}

	return []interface{}{mDestination}
}

func flattenSNSDestination(destination *ses.SNSDestination) []interface{} {
	if destination == nil {
		return []interface{}{}
	}

	mDestination := map[string]interface{}{
		names.AttrTopicARN: aws.StringValue(destination.TopicARN),
	}

	return []interface{}{mDestination}
}
