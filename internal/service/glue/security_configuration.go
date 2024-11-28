// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_security_configuration")
func ResourceSecurityConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityConfigurationCreate,
		ReadWithoutTimeout:   resourceSecurityConfigurationRead,
		DeleteWithoutTimeout: resourceSecurityConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrEncryptionConfiguration: {
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
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.CloudWatchEncryptionModeDisabled,
										ValidateDiagFunc: enum.Validate[awstypes.CloudWatchEncryptionMode](),
									},
									names.AttrKMSKeyARN: {
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
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.JobBookmarksEncryptionModeDisabled,
										ValidateDiagFunc: enum.Validate[awstypes.JobBookmarksEncryptionMode](),
									},
									names.AttrKMSKeyARN: {
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
									names.AttrKMSKeyARN: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"s3_encryption_mode": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Default:          awstypes.S3EncryptionModeDisabled,
										ValidateDiagFunc: enum.Validate[awstypes.S3EncryptionMode](),
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceSecurityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)
	name := d.Get(names.AttrName).(string)

	input := &glue.CreateSecurityConfigurationInput{
		EncryptionConfiguration: expandEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]interface{})),
		Name:                    aws.String(name),
	}

	log.Printf("[DEBUG] Creating Glue Security Configuration: %+v", input)
	_, err := conn.CreateSecurityConfiguration(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Security Configuration (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceSecurityConfigurationRead(ctx, d, meta)...)
}

func resourceSecurityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.GetSecurityConfigurationInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Security Configuration: %+v", input)
	output, err := conn.GetSecurityConfiguration(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		log.Printf("[WARN] Glue Security Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Security Configuration (%s): %s", d.Id(), err)
	}

	securityConfiguration := output.SecurityConfiguration
	if securityConfiguration == nil {
		log.Printf("[WARN] Glue Security Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set(names.AttrEncryptionConfiguration, flattenEncryptionConfiguration(securityConfiguration.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}

	d.Set(names.AttrName, securityConfiguration.Name)

	return diags
}

func resourceSecurityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Security Configuration: %s", d.Id())
	err := DeleteSecurityConfiguration(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Security Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func DeleteSecurityConfiguration(ctx context.Context, conn *glue.Client, name string) error {
	input := &glue.DeleteSecurityConfigurationInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteSecurityConfiguration(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil
	}

	return err
}

func expandCloudWatchEncryption(l []interface{}) *awstypes.CloudWatchEncryption {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cloudwatchEncryption := &awstypes.CloudWatchEncryption{
		CloudWatchEncryptionMode: awstypes.CloudWatchEncryptionMode(m["cloudwatch_encryption_mode"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyARN]; ok && v.(string) != "" {
		cloudwatchEncryption.KmsKeyArn = aws.String(v.(string))
	}

	return cloudwatchEncryption
}

func expandEncryptionConfiguration(l []interface{}) *awstypes.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	encryptionConfiguration := &awstypes.EncryptionConfiguration{
		CloudWatchEncryption:   expandCloudWatchEncryption(m["cloudwatch_encryption"].([]interface{})),
		JobBookmarksEncryption: expandJobBookmarksEncryption(m["job_bookmarks_encryption"].([]interface{})),
		S3Encryption:           expandS3Encryptions(m["s3_encryption"].([]interface{})),
	}

	return encryptionConfiguration
}

func expandJobBookmarksEncryption(l []interface{}) *awstypes.JobBookmarksEncryption {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	jobBookmarksEncryption := &awstypes.JobBookmarksEncryption{
		JobBookmarksEncryptionMode: awstypes.JobBookmarksEncryptionMode(m["job_bookmarks_encryption_mode"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyARN]; ok && v.(string) != "" {
		jobBookmarksEncryption.KmsKeyArn = aws.String(v.(string))
	}

	return jobBookmarksEncryption
}

func expandS3Encryptions(l []interface{}) []awstypes.S3Encryption {
	s3Encryptions := make([]awstypes.S3Encryption, 0)

	for _, s3Encryption := range l {
		if s3Encryption == nil {
			continue
		}
		s3Encryptions = append(s3Encryptions, expandS3Encryption(s3Encryption.(map[string]interface{})))
	}

	return s3Encryptions
}

func expandS3Encryption(m map[string]interface{}) awstypes.S3Encryption {
	s3Encryption := awstypes.S3Encryption{
		S3EncryptionMode: awstypes.S3EncryptionMode(m["s3_encryption_mode"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyARN]; ok && v.(string) != "" {
		s3Encryption.KmsKeyArn = aws.String(v.(string))
	}

	return s3Encryption
}

func flattenCloudWatchEncryption(cloudwatchEncryption *awstypes.CloudWatchEncryption) []interface{} {
	if cloudwatchEncryption == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_encryption_mode": string(cloudwatchEncryption.CloudWatchEncryptionMode),
		names.AttrKMSKeyARN:          aws.ToString(cloudwatchEncryption.KmsKeyArn),
	}

	return []interface{}{m}
}

func flattenEncryptionConfiguration(encryptionConfiguration *awstypes.EncryptionConfiguration) []interface{} {
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

func flattenJobBookmarksEncryption(jobBookmarksEncryption *awstypes.JobBookmarksEncryption) []interface{} {
	if jobBookmarksEncryption == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"job_bookmarks_encryption_mode": string(jobBookmarksEncryption.JobBookmarksEncryptionMode),
		names.AttrKMSKeyARN:             aws.ToString(jobBookmarksEncryption.KmsKeyArn),
	}

	return []interface{}{m}
}

func flattenS3Encryptions(s3Encryptions []awstypes.S3Encryption) []interface{} {
	l := make([]interface{}, 0)

	for _, s3Encryption := range s3Encryptions {
		l = append(l, flattenS3Encryption(s3Encryption))
	}

	return l
}

func flattenS3Encryption(s3Encryption awstypes.S3Encryption) map[string]interface{} {
	m := map[string]interface{}{
		names.AttrKMSKeyARN:  aws.ToString(s3Encryption.KmsKeyArn),
		"s3_encryption_mode": string(s3Encryption.S3EncryptionMode),
	}

	return m
}
