package codebuild

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceReportGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceReportGroupCreate,
		Read:   resourceReportGroupRead,
		Update: resourceReportGroupUpdate,
		Delete: resourceReportGroupDelete,

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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codebuild.ReportType_Values(), false),
			},
			"export_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ReportExportConfigType_Values(), false),
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"packaging": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      codebuild.ReportPackagingTypeNone,
										ValidateFunc: validation.StringInSlice(codebuild.ReportPackagingType_Values(), false),
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
			"delete_reports": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReportGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	createOpts := &codebuild.CreateReportGroupInput{
		Name:         aws.String(d.Get("name").(string)),
		Type:         aws.String(d.Get("type").(string)),
		ExportConfig: expandReportGroupExportConfig(d.Get("export_config").([]interface{})),
		Tags:         Tags(tags.IgnoreAWS()),
	}

	resp, err := conn.CreateReportGroup(createOpts)
	if err != nil {
		return fmt.Errorf("error creating CodeBuild Report Group: %w", err)
	}

	d.SetId(aws.StringValue(resp.ReportGroup.Arn))

	return resourceReportGroupRead(d, meta)
}

func resourceReportGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	reportGroup, err := FindReportGroupByARN(conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CodeBuild, names.ErrActionReading, ResReportGroup, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodeBuild, names.ErrActionReading, ResReportGroup, d.Id(), err)
	}

	if !d.IsNewResource() && reportGroup == nil {
		names.LogNotFoundRemoveState(names.CodeBuild, names.ErrActionReading, ResReportGroup, d.Id())
		d.SetId("")
		return nil
	}

	if reportGroup == nil {
		return names.Error(names.CodeBuild, names.ErrActionReading, ResReportGroup, d.Id(), errors.New("not found after creation"))
	}

	d.Set("arn", reportGroup.Arn)
	d.Set("type", reportGroup.Type)
	d.Set("name", reportGroup.Name)

	if err := d.Set("created", reportGroup.Created.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting created: %w", err)
	}

	if err := d.Set("export_config", flattenReportGroupExportConfig(reportGroup.ExportConfig)); err != nil {
		return fmt.Errorf("error setting export config: %w", err)
	}

	tags := KeyValueTags(reportGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceReportGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &codebuild.UpdateReportGroupInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("export_config") {
		input.ExportConfig = expandReportGroupExportConfig(d.Get("export_config").([]interface{}))
	}

	if d.HasChange("tags_all") {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.UpdateReportGroup(input)
	if err != nil {
		return fmt.Errorf("error updating CodeBuild Report Group: %w", err)
	}

	return resourceReportGroupRead(d, meta)
}

func resourceReportGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	deleteOpts := &codebuild.DeleteReportGroupInput{
		Arn:           aws.String(d.Id()),
		DeleteReports: aws.Bool(d.Get("delete_reports").(bool)),
	}

	if _, err := conn.DeleteReportGroup(deleteOpts); err != nil {
		return fmt.Errorf("error deleting CodeBuild Report Group (%s): %w", d.Id(), err)
	}

	if _, err := waitReportGroupDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error while waiting for CodeBuild Report Group (%s) to become deleted: %w", d.Id(), err)
	}

	return nil
}

func expandReportGroupExportConfig(config []interface{}) *codebuild.ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	exportConfig := &codebuild.ReportExportConfig{}

	if v, ok := s["type"]; ok {
		exportConfig.ExportConfigType = aws.String(v.(string))
	}

	if v, ok := s["s3_destination"]; ok {
		exportConfig.S3Destination = expandReportGroupS3ReportExportConfig(v.([]interface{}))
	}

	return exportConfig
}

func flattenReportGroupExportConfig(config *codebuild.ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["s3_destination"] = flattenReportGroupS3ReportExportConfig(config.S3Destination)
	settings["type"] = aws.StringValue(config.ExportConfigType)

	return []map[string]interface{}{settings}
}

func expandReportGroupS3ReportExportConfig(config []interface{}) *codebuild.S3ReportExportConfig {
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

func flattenReportGroupS3ReportExportConfig(config *codebuild.S3ReportExportConfig) []map[string]interface{} {
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
