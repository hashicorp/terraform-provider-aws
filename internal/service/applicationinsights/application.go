package applicationinsights

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceApplicationCreate,
		Read:   resourceApplicationRead,
		Update: resourceApplicationUpdate,
		Delete: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"auto_config_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_create": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cwe_monitor_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"grouping_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(applicationinsights.GroupingType_Values(), false),
			},
			"ops_center_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ops_item_sns_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &applicationinsights.CreateApplicationInput{
		AutoConfigEnabled: aws.Bool(d.Get("auto_config_enabled").(bool)),
		AutoCreate:        aws.Bool(d.Get("auto_create").(bool)),
		CWEMonitorEnabled: aws.Bool(d.Get("cwe_monitor_enabled").(bool)),
		OpsCenterEnabled:  aws.Bool(d.Get("ops_center_enabled").(bool)),
		ResourceGroupName: aws.String(d.Get("resource_group_name").(string)),
		Tags:              Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("grouping_type"); ok {
		input.GroupingType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ops_item_sns_topic_arn"); ok {
		input.OpsItemSNSTopicArn = aws.String(v.(string))
	}

	out, err := conn.CreateApplication(input)
	if err != nil {
		return fmt.Errorf("Error creating ApplicationInsights Application: %w", err)
	}

	d.SetId(aws.StringValue(out.ApplicationInfo.ResourceGroupName))

	if _, err := waitApplicationCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for ApplicationInsights Application (%s) create: %w", d.Id(), err)
	}

	return resourceApplicationRead(d, meta)
}

func resourceApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	application, err := FindApplicationByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ApplicationInsights Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ApplicationInsights Application (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/resource-group/%s", aws.StringValue(application.ResourceGroupName)),
		Service:   "applicationinsights",
	}.String()

	d.Set("arn", arn)
	d.Set("resource_group_name", application.ResourceGroupName)
	d.Set("auto_config_enabled", application.AutoConfigEnabled)
	d.Set("cwe_monitor_enabled", application.CWEMonitorEnabled)
	d.Set("ops_center_enabled", application.OpsCenterEnabled)
	d.Set("ops_item_sns_topic_arn", application.OpsItemSNSTopicArn)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for ApplicationInsights Application (%s): %w", arn, err)
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

func resourceApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &applicationinsights.UpdateApplicationInput{
			ResourceGroupName: aws.String(d.Id()),
		}

		if d.HasChange("auto_config_enabled") {
			input.AutoConfigEnabled = aws.Bool(d.Get("auto_config_enabled").(bool))
		}

		if d.HasChange("cwe_monitor_enabled") {
			input.CWEMonitorEnabled = aws.Bool(d.Get("cwe_monitor_enabled").(bool))
		}

		if d.HasChange("ops_center_enabled") {
			input.OpsCenterEnabled = aws.Bool(d.Get("ops_center_enabled").(bool))
		}

		if d.HasChange("ops_item_sns_topic_arn") {
			_, n := d.GetChange("ops_item_sns_topic_arn")
			if n != nil {
				input.OpsItemSNSTopicArn = aws.String(n.(string))
			} else {
				input.RemoveSNSTopic = aws.Bool(true)
			}

		}

		log.Printf("[DEBUG] Updating ApplicationInsights Application: %s", d.Id())
		_, err := conn.UpdateApplication(input)
		if err != nil {
			return fmt.Errorf("Error Updating ApplicationInsights Application: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating ApplicationInsights Application (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceApplicationRead(d, meta)
}

func resourceApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn

	input := &applicationinsights.DeleteApplicationInput{
		ResourceGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting ApplicationInsights Application: %s", d.Id())
	_, err := conn.DeleteApplication(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, applicationinsights.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting ApplicationInsights Application: %w", err)
	}

	if _, err := waitApplicationTerminated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for ApplicationInsights Application (%s) delete: %w", d.Id(), err)
	}

	return nil
}
