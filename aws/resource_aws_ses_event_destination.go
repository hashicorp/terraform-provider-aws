package aws

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsSesEventDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesEventDestinationCreate,
		Read:   resourceAwsSesEventDestinationRead,
		Delete: resourceAwsSesEventDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected NAME/CONFIGURATION_SET_NAME", d.Id())
				}
				d.SetId(idParts[0])
				d.Set("name", idParts[0])
				d.Set("configuration_set_name", idParts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"configuration_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"matching_types": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						ses.EventTypeSend,
						ses.EventTypeReject,
						ses.EventTypeBounce,
						ses.EventTypeComplaint,
						ses.EventTypeDelivery,
						ses.EventTypeOpen,
						ses.EventTypeClick,
						ses.EventTypeRenderingFailure,
					}, false),
				},
			},

			"cloudwatch_destination": {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"kinesis_destination", "sns_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_value": {
							Type:     schema.TypeString,
							Required: true,
						},

						"dimension_name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"value_source": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								ses.DimensionValueSourceMessageTag,
								ses.DimensionValueSourceEmailHeader,
								ses.DimensionValueSourceLinkTag,
							}, false),
						},
					},
				},
			},

			"kinesis_destination": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				MaxItems:      1,
				ConflictsWith: []string{"cloudwatch_destination", "sns_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_arn": {
							Type:     schema.TypeString,
							Required: true,
						},

						"role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"sns_destination": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cloudwatch_destination", "kinesis_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topic_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsSesEventDestinationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configurationSetName := d.Get("configuration_set_name").(string)
	eventDestinationName := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	matchingEventTypes := d.Get("matching_types").(*schema.Set).List()

	createOpts := &ses.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(configurationSetName),
		EventDestination: &ses.EventDestination{
			Name:               aws.String(eventDestinationName),
			Enabled:            aws.Bool(enabled),
			MatchingEventTypes: expandStringList(matchingEventTypes),
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
			DeliveryStreamARN: aws.String(kinesis["stream_arn"].(string)),
			IAMRoleARN:        aws.String(kinesis["role_arn"].(string)),
		}
		log.Printf("[DEBUG] Creating kinesis destination: %#v", kinesis)
	}

	if v, ok := d.GetOk("sns_destination"); ok {
		destination := v.([]interface{})
		sns := destination[0].(map[string]interface{})
		createOpts.EventDestination.SNSDestination = &ses.SNSDestination{
			TopicARN: aws.String(sns["topic_arn"].(string)),
		}
		log.Printf("[DEBUG] Creating sns destination: %#v", sns)
	}

	_, err := conn.CreateConfigurationSetEventDestination(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating SES configuration set event destination: %s", err)
	}

	d.SetId(eventDestinationName)

	log.Printf("[WARN] SES DONE")
	return resourceAwsSesEventDestinationRead(d, meta)
}

func resourceAwsSesEventDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	log.Printf("[DEBUG] SES Read Configuration Set Destination: %s", d.Id())
	configurationSetName := d.Get("configuration_set_name").(string)
	input := &ses.DescribeConfigurationSetInput{
		ConfigurationSetName: aws.String(configurationSetName),
	}
	resp, err := conn.DescribeConfigurationSet(input)
	if err != nil {
		if isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
			log.Printf("[WARN] ")
			d.SetId("")
			return nil
		}
		return err
	}

	var eventDestination *ses.EventDestination
	for _, e := range resp.EventDestinations {
		if aws.StringValue(e.Name) == d.Id() {
			*eventDestination = *e
			break
		}
	}

	if eventDestination == nil {
		log.Printf("[WARN] ")
		d.SetId("")
		return nil
	}
	d.Set("name", aws.StringValue(eventDestination.Name))
	d.Set("enabled", aws.BoolValue(eventDestination.Enabled))
	if err := d.Set("matching_types", flattenStringSet(eventDestination.MatchingEventTypes)); err != nil {
		return fmt.Errorf("error setting matching_types: %s", err)
	}
	if err := d.Set("cloudwatch_destination", flattenCloudWatchDimensionConfigurations(eventDestination.CloudWatchDestination)); err != nil {
		return fmt.Errorf("error setting cloudwatch_destination: %s", err)
	}
	if err := d.Set("kinesis_destination", flattenKinesisFirehoseDestination(eventDestination.KinesisFirehoseDestination)); err != nil {
		return fmt.Errorf("error setting kinesis_destination: %s", err)
	}
	if err := d.Set("sns_destination", flattenSnsDestination(eventDestination.SNSDestination)); err != nil {
		return fmt.Errorf("error setting sns_destination: %s", err)
	}

	return nil
}

func resourceAwsSesEventDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	log.Printf("[DEBUG] SES Delete Configuration Set Destination: %s", d.Id())
	_, err := conn.DeleteConfigurationSetEventDestination(&ses.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		EventDestinationName: aws.String(d.Id()),
	})

	return err
}

func flattenCloudWatchDimensionConfigurations(cwd *ses.CloudWatchDestination) *schema.Set {
	if cwd == nil || len(cwd.DimensionConfigurations) == 0 {
		return nil
	}
	out := make([]interface{}, len(cwd.DimensionConfigurations))
	for _, config := range cwd.DimensionConfigurations {
		c := make(map[string]interface{})
		c["default_value"] = aws.StringValue(config.DefaultDimensionValue)
		c["dimension_name"] = aws.StringValue(config.DimensionName)
		c["value_source"] = aws.StringValue(config.DimensionValueSource)
		out = append(out, c)
	}
	return schema.NewSet(cloudWatchDimensionConfigurationHash, out)
}

func flattenKinesisFirehoseDestination(kfd *ses.KinesisFirehoseDestination) []interface{} {
	if kfd == nil {
		return []interface{}{}
	}
	destination := map[string]interface{}{
		"stream_arn": aws.StringValue(kfd.DeliveryStreamARN),
		"role_arn":   aws.StringValue(kfd.IAMRoleARN),
	}
	return []interface{}{destination}
}

func flattenSnsDestination(s *ses.SNSDestination) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	destination := map[string]interface{}{
		"topic_arn": aws.StringValue(s.TopicARN),
	}
	return []interface{}{destination}
}

func cloudWatchDimensionConfigurationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["default_value"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["dimension_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value_source"].(string)))

	return hashcode.String(buf.String())
}

func generateCloudWatchDestination(v []interface{}) []*ses.CloudWatchDimensionConfiguration {

	b := make([]*ses.CloudWatchDimensionConfiguration, len(v))

	for i, vI := range v {
		cloudwatch := vI.(map[string]interface{})
		b[i] = &ses.CloudWatchDimensionConfiguration{
			DefaultDimensionValue: aws.String(cloudwatch["default_value"].(string)),
			DimensionName:         aws.String(cloudwatch["dimension_name"].(string)),
			DimensionValueSource:  aws.String(cloudwatch["value_source"].(string)),
		}
	}

	return b
}
