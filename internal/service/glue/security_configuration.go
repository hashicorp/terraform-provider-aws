package glue

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceSecurityConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecurityConfigurationCreate,
		Read:   resourceSecurityConfigurationRead,
		Delete: resourceSecurityConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"encryption_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_encryption": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch_encryption_mode": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  glue.CloudWatchEncryptionModeDisabled,
										ValidateFunc: validation.StringInSlice([]string{
											glue.CloudWatchEncryptionModeDisabled,
											glue.CloudWatchEncryptionModeSseKms,
										}, false),
									},
									"kms_key_arn": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"job_bookmarks_encryption": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"job_bookmarks_encryption_mode": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  glue.JobBookmarksEncryptionModeDisabled,
										ValidateFunc: validation.StringInSlice([]string{
											glue.JobBookmarksEncryptionModeCseKms,
											glue.JobBookmarksEncryptionModeDisabled,
										}, false),
									},
									"kms_key_arn": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"s3_encryption": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_key_arn": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"s3_encryption_mode": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  glue.S3EncryptionModeDisabled,
										ValidateFunc: validation.StringInSlice([]string{
											glue.S3EncryptionModeDisabled,
											glue.S3EncryptionModeSseKms,
											glue.S3EncryptionModeSseS3,
										}, false),
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceSecurityConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	name := d.Get("name").(string)

	input := &glue.CreateSecurityConfigurationInput{
		EncryptionConfiguration: expandEncryptionConfiguration(d.Get("encryption_configuration").([]interface{})),
		Name:                    aws.String(name),
	}

	log.Printf("[DEBUG] Creating Glue Security Configuration: %s", input)
	_, err := conn.CreateSecurityConfiguration(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Security Configuration (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceSecurityConfigurationRead(d, meta)
}

func resourceSecurityConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	input := &glue.GetSecurityConfigurationInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Security Configuration: %s", input)
	output, err := conn.GetSecurityConfiguration(input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Glue Security Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Glue Security Configuration (%s): %s", d.Id(), err)
	}

	securityConfiguration := output.SecurityConfiguration
	if securityConfiguration == nil {
		log.Printf("[WARN] Glue Security Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("encryption_configuration", flattenEncryptionConfiguration(securityConfiguration.EncryptionConfiguration)); err != nil {
		return fmt.Errorf("error setting encryption_configuration: %s", err)
	}

	d.Set("name", securityConfiguration.Name)

	return nil
}

func resourceSecurityConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	log.Printf("[DEBUG] Deleting Glue Security Configuration: %s", d.Id())
	err := DeleteSecurityConfiguration(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Security Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func DeleteSecurityConfiguration(conn *glue.Glue, name string) error {
	input := &glue.DeleteSecurityConfigurationInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteSecurityConfiguration(input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func expandCloudWatchEncryption(l []interface{}) *glue.CloudWatchEncryption {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cloudwatchEncryption := &glue.CloudWatchEncryption{
		CloudWatchEncryptionMode: aws.String(m["cloudwatch_encryption_mode"].(string)),
	}

	if v, ok := m["kms_key_arn"]; ok && v.(string) != "" {
		cloudwatchEncryption.KmsKeyArn = aws.String(v.(string))
	}

	return cloudwatchEncryption
}

func expandEncryptionConfiguration(l []interface{}) *glue.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	encryptionConfiguration := &glue.EncryptionConfiguration{
		CloudWatchEncryption:   expandCloudWatchEncryption(m["cloudwatch_encryption"].([]interface{})),
		JobBookmarksEncryption: expandJobBookmarksEncryption(m["job_bookmarks_encryption"].([]interface{})),
		S3Encryption:           expandS3Encryptions(m["s3_encryption"].([]interface{})),
	}

	return encryptionConfiguration
}

func expandJobBookmarksEncryption(l []interface{}) *glue.JobBookmarksEncryption {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	jobBookmarksEncryption := &glue.JobBookmarksEncryption{
		JobBookmarksEncryptionMode: aws.String(m["job_bookmarks_encryption_mode"].(string)),
	}

	if v, ok := m["kms_key_arn"]; ok && v.(string) != "" {
		jobBookmarksEncryption.KmsKeyArn = aws.String(v.(string))
	}

	return jobBookmarksEncryption
}

func expandS3Encryptions(l []interface{}) []*glue.S3Encryption {
	s3Encryptions := make([]*glue.S3Encryption, 0)

	for _, s3Encryption := range l {
		if s3Encryption == nil {
			continue
		}
		s3Encryptions = append(s3Encryptions, expandS3Encryption(s3Encryption.(map[string]interface{})))
	}

	return s3Encryptions
}

func expandS3Encryption(m map[string]interface{}) *glue.S3Encryption {
	s3Encryption := &glue.S3Encryption{
		S3EncryptionMode: aws.String(m["s3_encryption_mode"].(string)),
	}

	if v, ok := m["kms_key_arn"]; ok && v.(string) != "" {
		s3Encryption.KmsKeyArn = aws.String(v.(string))
	}

	return s3Encryption
}

func flattenCloudWatchEncryption(cloudwatchEncryption *glue.CloudWatchEncryption) []interface{} {
	if cloudwatchEncryption == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_encryption_mode": aws.StringValue(cloudwatchEncryption.CloudWatchEncryptionMode),
		"kms_key_arn":                aws.StringValue(cloudwatchEncryption.KmsKeyArn),
	}

	return []interface{}{m}
}

func flattenEncryptionConfiguration(encryptionConfiguration *glue.EncryptionConfiguration) []interface{} {
	if encryptionConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_encryption":    flattenCloudWatchEncryption(encryptionConfiguration.CloudWatchEncryption),
		"job_bookmarks_encryption": flattenJobBookmarksEncryption(encryptionConfiguration.JobBookmarksEncryption),
		"s3_encryption":            flattenS3Encryptions(encryptionConfiguration.S3Encryption),
	}

	return []interface{}{m}
}

func flattenJobBookmarksEncryption(jobBookmarksEncryption *glue.JobBookmarksEncryption) []interface{} {
	if jobBookmarksEncryption == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"job_bookmarks_encryption_mode": aws.StringValue(jobBookmarksEncryption.JobBookmarksEncryptionMode),
		"kms_key_arn":                   aws.StringValue(jobBookmarksEncryption.KmsKeyArn),
	}

	return []interface{}{m}
}

func flattenS3Encryptions(s3Encryptions []*glue.S3Encryption) []interface{} {
	l := make([]interface{}, 0)

	for _, s3Encryption := range s3Encryptions {
		if s3Encryption == nil {
			continue
		}
		l = append(l, flattenS3Encryption(s3Encryption))
	}

	return l
}

func flattenS3Encryption(s3Encryption *glue.S3Encryption) map[string]interface{} {
	m := map[string]interface{}{
		"kms_key_arn":        aws.StringValue(s3Encryption.KmsKeyArn),
		"s3_encryption_mode": aws.StringValue(s3Encryption.S3EncryptionMode),
	}

	return m
}
