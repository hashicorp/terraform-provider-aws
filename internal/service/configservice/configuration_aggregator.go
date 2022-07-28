package configservice

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceConfigurationAggregator() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationAggregatorPut,
		Read:   resourceConfigurationAggregatorRead,
		Update: resourceConfigurationAggregatorPut,
		Delete: resourceConfigurationAggregatorDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			// This is to prevent this error:
			// All fields are ForceNew or Computed w/out Optional, Update is superfluous
			customdiff.ForceNewIfChange("account_aggregation_source", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0
			}),
			customdiff.ForceNewIfChange("organization_aggregation_source", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"account_aggregation_source": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"organization_aggregation_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_ids": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidAccountID,
							},
						},
						"all_regions": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"organization_aggregation_source": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"account_aggregation_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_regions": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConfigurationAggregatorPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &configservice.PutConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Get("name").(string)),
		Tags:                        Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("account_aggregation_source"); ok && len(v.([]interface{})) > 0 {
		req.AccountAggregationSources = expandAccountAggregationSources(v.([]interface{}))
	}

	if v, ok := d.GetOk("organization_aggregation_source"); ok && len(v.([]interface{})) > 0 {
		req.OrganizationAggregationSource = expandOrganizationAggregationSource(v.([]interface{})[0].(map[string]interface{}))
	}

	resp, err := conn.PutConfigurationAggregator(req)
	if err != nil {
		return fmt.Errorf("error creating aggregator: %w", err)
	}

	configAgg := resp.ConfigurationAggregator
	d.SetId(aws.StringValue(configAgg.ConfigurationAggregatorName))

	if !d.IsNewResource() && d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		arn := aws.StringValue(configAgg.ConfigurationAggregatorArn)
		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Config Configuration Aggregator (%s) tags: %w", arn, err)
		}
	}

	return resourceConfigurationAggregatorRead(d, meta)
}

func resourceConfigurationAggregatorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &configservice.DescribeConfigurationAggregatorsInput{
		ConfigurationAggregatorNames: []*string{aws.String(d.Id())},
	}

	res, err := conn.DescribeConfigurationAggregators(req)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationAggregatorException) {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Configuration Aggregator", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.ConfigService, names.ErrActionReading, "Configuration Aggregator", d.Id(), err)
	}

	if !d.IsNewResource() && (res == nil || len(res.ConfigurationAggregators) == 0) {
		log.Printf("[WARN] No aggregators returned (%s), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && (res == nil || len(res.ConfigurationAggregators) == 0) {
		return names.Error(names.ConfigService, names.ErrActionReading, "Configuration Aggregator", d.Id(), errors.New("not found after creation"))
	}

	aggregator := res.ConfigurationAggregators[0]
	arn := aws.StringValue(aggregator.ConfigurationAggregatorArn)
	d.Set("arn", arn)
	d.Set("name", aggregator.ConfigurationAggregatorName)

	if err := d.Set("account_aggregation_source", flattenAccountAggregationSources(aggregator.AccountAggregationSources)); err != nil {
		return fmt.Errorf("error setting account_aggregation_source: %s", err)
	}

	if err := d.Set("organization_aggregation_source", flattenOrganizationAggregationSource(aggregator.OrganizationAggregationSource)); err != nil {
		return fmt.Errorf("error setting organization_aggregation_source: %s", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Config Configuration Aggregator (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceConfigurationAggregatorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	req := &configservice.DeleteConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Id()),
	}
	_, err := conn.DeleteConfigurationAggregator(req)

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationAggregatorException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Config Configuration Aggregator (%s): %w", d.Id(), err)
	}

	return nil
}
