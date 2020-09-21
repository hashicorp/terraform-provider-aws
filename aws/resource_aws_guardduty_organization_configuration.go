package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsGuardDutyOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyOrganizationConfigurationUpdate,
		Read:   resourceAwsGuardDutyOrganizationConfigurationRead,
		Update: resourceAwsGuardDutyOrganizationConfigurationUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"detector_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
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
		},
	}
}

func resourceAwsGuardDutyOrganizationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)

	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnable:  aws.Bool(d.Get("auto_enable").(bool)),
		DetectorId:  aws.String(detectorID),
		DataSources: expandOrganizationDatasourceConfig(d),
	}

	_, err := conn.UpdateOrganizationConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating GuardDuty Organization Configuration (%s): %w", detectorID, err)
	}

	d.SetId(detectorID)

	return resourceAwsGuardDutyOrganizationConfigurationRead(d, meta)
}

func resourceAwsGuardDutyOrganizationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.DescribeOrganizationConfigurationInput{
		DetectorId: aws.String(d.Id()),
	}

	output, err := conn.DescribeOrganizationConfiguration(input)

	if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
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

	if err := d.Set("datasources", flattenOrganizationDatasourceConfig(output.DataSources)); err != nil {
		return fmt.Errorf("error setting datasources: %s", err)
	}

	d.Set("detector_id", d.Id())
	d.Set("auto_enable", output.AutoEnable)

	return nil
}

func flattenOrganizationDatasourceConfig(dsConfig *guardduty.OrganizationDataSourceConfigurationsResult) []interface{} {
	if dsConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := dsConfig.S3Logs; v != nil {
		values["s3_logs"] = flattenOrganizationS3LogsConfig(v)
	}

	return []interface{}{values}
}

func flattenOrganizationS3LogsConfig(s3LogsConfig *guardduty.OrganizationS3LogsConfigurationResult) []interface{} {
	values := map[string]interface{}{}

	if s3LogsConfig == nil {
		values["auto_enable"] = false
	} else {
		values["auto_enable"] = aws.BoolValue(s3LogsConfig.AutoEnable)
	}

	return []interface{}{values}
}

func expandOrganizationDatasourceConfig(d *schema.ResourceData) *guardduty.OrganizationDataSourceConfigurations {
	dsConfig := &guardduty.OrganizationDataSourceConfigurations{}

	if v, ok := d.GetOk("datasources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		configList := v.([]interface{})
		data := configList[0].(map[string]interface{})

		if v, ok := data["s3_logs"]; ok {
			dsConfig.S3Logs = expandOrganizationS3LogsConfig(v.([]interface{}))
		}
	}

	return dsConfig
}

func expandOrganizationS3LogsConfig(configList []interface{}) *guardduty.OrganizationS3LogsConfiguration {
	if len(configList) == 0 || configList[0] == nil {
		return nil
	}

	data := configList[0].(map[string]interface{})

	autoEnable := data["auto_enable"].(bool)

	s3LogsConfig := &guardduty.OrganizationS3LogsConfiguration{
		AutoEnable: aws.Bool(autoEnable),
	}

	return s3LogsConfig
}
