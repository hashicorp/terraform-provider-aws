package backup

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFramework() *schema.Resource {
	return &schema.Resource{
		Read: resourceFrameworkRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_parameter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"scope": {
							// The control scope can include
							// one or more resource types,
							// a combination of a tag key and value,
							// or a combination of one resource type and one resource ID.
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_resource_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"compliance_resource_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									// A maximum of one key-value pair can be provided.
									// The tag value is optional, but it cannot be an empty string
									"tags": tftags.TagsSchema(),
								},
							},
						},
					},
				},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validFrameworkName,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFrameworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeFramework(&backup.DescribeFrameworkInput{
		FrameworkName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Framework (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Backup Framework (%s): %w", d.Id(), err)
	}

	d.Set("arn", resp.FrameworkArn)
	d.Set("deployment_status", resp.DeploymentStatus)
	d.Set("description", resp.FrameworkDescription)
	d.Set("name", resp.FrameworkName)
	d.Set("status", resp.FrameworkStatus)

	if err := d.Set("creation_time", resp.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}

	if err := d.Set("control", flattenFrameworkControls(resp.FrameworkControls)); err != nil {
		return fmt.Errorf("error setting control: %w", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Framework (%s): %w", d.Id(), err)
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


func flattenFrameworkControls(controls []*backup.FrameworkControl) []interface{} {
	if controls == nil {
		return []interface{}{}
	}

	frameworkControls := []interface{}{}
	for _, control := range controls {
		values := map[string]interface{}{}
		values["input_parameter"] = flattenInputParameters(control.ControlInputParameters)
		values["name"] = aws.StringValue(control.ControlName)
		values["scope"] = flattenScope(control.ControlScope)
		frameworkControls = append(frameworkControls, values)
	}
	return frameworkControls
}

func flattenInputParameters(inputParams []*backup.ControlInputParameter) []interface{} {
	if inputParams == nil {
		return []interface{}{}
	}

	controlInputParameters := []interface{}{}
	for _, inputParam := range inputParams {
		values := map[string]interface{}{}
		values["name"] = aws.StringValue(inputParam.ParameterName)
		values["value"] = aws.StringValue(inputParam.ParameterValue)
		controlInputParameters = append(controlInputParameters, values)
	}
	return controlInputParameters
}

func flattenScope(scope *backup.ControlScope) []interface{} {
	if scope == nil {
		return []interface{}{}
	}

	controlScope := map[string]interface{}{
		"compliance_resource_ids":   flex.FlattenStringList(scope.ComplianceResourceIds),
		"compliance_resource_types": flex.FlattenStringList(scope.ComplianceResourceTypes),
	}

	if v := scope.Tags; v != nil {
		controlScope["tags"] = KeyValueTags(v).IgnoreAWS().Map()
	}

	return []interface{}{controlScope}
}
