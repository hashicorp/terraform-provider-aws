package guardduty

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationConfigurationUpdate,
		Read:   resourceOrganizationConfigurationRead,
		Update: resourceOrganizationConfigurationUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"datasources": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_enable": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"detector_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn

	detectorID := d.Get("detector_id").(string)

	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
		DetectorId: aws.String(detectorID),
	}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSources = expandOrganizationDataSourceConfigurations(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateOrganizationConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating GuardDuty Organization Configuration (%s): %w", detectorID, err)
	}

	d.SetId(detectorID)

	return resourceOrganizationConfigurationRead(d, meta)
}

func resourceOrganizationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn

	input := &guardduty.DescribeOrganizationConfigurationInput{
		DetectorId: aws.String(d.Id()),
	}

	output, err := conn.DescribeOrganizationConfiguration(input)

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		log.Printf("[WARN] GuardDuty Organization Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading GuardDuty Organization Configuration (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading GuardDuty Organization Configuration (%s): empty response", d.Id())
	}

	d.Set("auto_enable", output.AutoEnable)

	if output.DataSources != nil {
		if err := d.Set("datasources", []interface{}{flattenOrganizationDataSourceConfigurationsResult(output.DataSources)}); err != nil {
			return fmt.Errorf("error setting datasources: %w", err)
		}
	} else {
		d.Set("datasources", nil)
	}

	d.Set("detector_id", d.Id())

	return nil
}

func expandOrganizationDataSourceConfigurations(tfMap map[string]interface{}) *guardduty.OrganizationDataSourceConfigurations {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationDataSourceConfigurations{}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Logs = expandOrganizationS3LogsConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandOrganizationS3LogsConfiguration(tfMap map[string]interface{}) *guardduty.OrganizationS3LogsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &guardduty.OrganizationS3LogsConfiguration{}

	if v, ok := tfMap["auto_enable"].(bool); ok {
		apiObject.AutoEnable = aws.Bool(v)
	}

	return apiObject
}

func flattenOrganizationDataSourceConfigurationsResult(apiObject *guardduty.OrganizationDataSourceConfigurationsResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenOrganizationS3LogsConfigurationResult(v)}
	}

	return tfMap
}

func flattenOrganizationS3LogsConfigurationResult(apiObject *guardduty.OrganizationS3LogsConfigurationResult) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoEnable; v != nil {
		tfMap["auto_enable"] = aws.BoolValue(v)
	}

	return tfMap
}
