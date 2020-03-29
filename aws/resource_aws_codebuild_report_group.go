package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsCodeBuildReportGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildReportGroupCreate,
		Read:   resourceAwsCodeBuildReportGroupRead,
		Update: resourceAwsCodeBuildReportGroupUpdate,
		Delete: resourceAwsCodeBuildReportGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codebuild.ReportTypeTest,
				}, false),
			},
			"export_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ReportExportConfigTypeNoExport,
								codebuild.ReportExportConfigTypeS3,
							}, false),
						},
						"s3_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Required: true,
									},
									"encryption_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"encryption_key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"packaging": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  codebuild.ReportPackagingTypeNone,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.ReportPackagingTypeNone,
											codebuild.ReportPackagingTypeZip,
										}, false),
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeBuildReportGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn
	createOpts := &codebuild.CreateReportGroupInput{
		Name:         aws.String(d.Get("name").(string)),
		Type:         aws.String(d.Get("type").(string)),
		ExportConfig: expandAwsCodeBuildReportGroupExportConfig(d.Get("export_config").([]interface{})),
	}

	resp, err := conn.CreateReportGroup(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating CodeBuild Report Groups: %s", err)
	}

	d.SetId(aws.StringValue(resp.ReportGroup.Arn))

	return resourceAwsCodeBuildReportGroupRead(d, meta)
}

func resourceAwsCodeBuildReportGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	resp, err := conn.BatchGetReportGroups(&codebuild.BatchGetReportGroupsInput{
		ReportGroupArns: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		return fmt.Errorf("Error Listing CodeBuild Report Groups: %s", err)
	}

	if len(resp.ReportGroups) == 0 {
		return fmt.Errorf("no matches found for CodeBuild Report Groups: %s", d.Id())
	}

	if len(resp.ReportGroups) > 1 {
		return fmt.Errorf("multiple matches found for CodeBuild Report Groups: %s", d.Id())
	}

	reportGroups := resp.ReportGroups[0]

	if reportGroups == nil {
		log.Printf("[WARN] CodeBuild Report Groups (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", reportGroups.Arn)
	d.Set("type", reportGroups.Type)
	d.Set("name", reportGroups.Name)

	if err := d.Set("created", reportGroups.Created.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting created: %s", err)
	}

	if err := d.Set("export_config", flattenAwsCodeBuildReportGroupExportConfig(reportGroups.ExportConfig)); err != nil {
		return fmt.Errorf("error setting export config: %s", err)
	}

	return nil
}

func resourceAwsCodeBuildReportGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	if d.HasChange("export_config") {
		input := &codebuild.UpdateReportGroupInput{
			Arn:          aws.String(d.Id()),
			ExportConfig: expandAwsCodeBuildReportGroupExportConfig(d.Get("export_config").([]interface{})),
		}

		_, err := conn.UpdateReportGroup(input)
		if err != nil {
			return fmt.Errorf("Error updating CodeBuild Report Groups: %s", err)
		}
	}

	return resourceAwsCodeBuildReportGroupRead(d, meta)
}

func resourceAwsCodeBuildReportGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	deleteOpts := &codebuild.DeleteReportGroupInput{
		Arn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteReportGroup(deleteOpts); err != nil {
		return fmt.Errorf("Error deleting CodeBuild Report Groups(%s): %s", d.Id(), err)
	}

	return nil
}

func expandAwsCodeBuildReportGroupExportConfig(config []interface{}) *codebuild.ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	exportConfig := &codebuild.ReportExportConfig{}

	if v, ok := s["type"]; ok {
		exportConfig.ExportConfigType = aws.String(v.(string))
	}

	if v, ok := s["s3_destination"]; ok {
		exportConfig.S3Destination = expandAwsCodeBuildReportGroupS3ReportExportConfig(v.([]interface{}))
	}

	return exportConfig
}

func flattenAwsCodeBuildReportGroupExportConfig(config *codebuild.ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["s3_destination"] = flattenAwsCodeBuildReportGroupS3ReportExportConfig(config.S3Destination)
	settings["type"] = aws.StringValue(config.ExportConfigType)

	return []map[string]interface{}{settings}
}

func expandAwsCodeBuildReportGroupS3ReportExportConfig(config []interface{}) *codebuild.S3ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	s3ReportExportConfig := &codebuild.S3ReportExportConfig{}

	if v, ok := s["bucket"]; ok {
		s3ReportExportConfig.Bucket = aws.String(v.(string))
	}
	if v, ok := s["encryption_disabled"]; ok {
		s3ReportExportConfig.EncryptionDisabled = aws.Bool(v.(bool))
	}

	if v, ok := s["encryption_key"]; ok {
		s3ReportExportConfig.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := s["packaging"]; ok {
		s3ReportExportConfig.Packaging = aws.String(v.(string))
	}

	if v, ok := s["path"]; ok {
		s3ReportExportConfig.Path = aws.String(v.(string))
	}

	return s3ReportExportConfig
}

func flattenAwsCodeBuildReportGroupS3ReportExportConfig(config *codebuild.S3ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["path"] = aws.StringValue(config.Path)
	settings["bucket"] = aws.StringValue(config.Bucket)
	settings["packaging"] = aws.StringValue(config.Packaging)
	settings["encryption_disabled"] = aws.BoolValue(config.EncryptionDisabled)

	if config.EncryptionKey != nil {
		settings["encryption_key"] = aws.StringValue(config.EncryptionKey)
	}

	return []map[string]interface{}{settings}
}
