// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elastictranscoder

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elastictranscoder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elastictranscoder_pipeline")
func ResourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineCreate,
		ReadWithoutTimeout:   resourcePipelineRead,
		UpdateWithoutTimeout: resourcePipelineUpdate,
		DeleteWithoutTimeout: resourcePipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},

			// ContentConfig also requires ThumbnailConfig
			"content_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.PipelineOutputConfig
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Optional: true,
							// AWS may insert the bucket name here taken from output_bucket
							Computed: true,
						},
						names.AttrStorageClass: {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Standard",
								"ReducedRedundancy",
							}, false),
						},
					},
				},
			},

			"content_config_permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"Read",
									"ReadAcp",
									"WriteAcp",
									"FullControl",
								}, false),
							},
						},
						"grantee": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"grantee_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Canonical",
								"Email",
								"Group",
							}, false),
						},
					},
				},
			},

			"input_bucket": {
				Type:     schema.TypeString,
				Required: true,
			},

			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`),
						"only alphanumeric characters, hyphens, underscores, and periods allowed"),
					validation.StringLenBetween(1, 40),
				),
			},

			"notifications": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"completed": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"error": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"progressing": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"warning": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},

			// The output_bucket must be set, or both of content_config.bucket
			// and thumbnail_config.bucket.
			// This is set as Computed, because the API may or may not return
			// this as set based on the other 2 configurations.
			"output_bucket": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrRole: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},

			"thumbnail_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.PipelineOutputConfig
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Optional: true,
							// AWS may insert the bucket name here taken from output_bucket
							Computed: true,
						},
						names.AttrStorageClass: {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Standard",
								"ReducedRedundancy",
							}, false),
						},
					},
				},
			},

			"thumbnail_config_permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"Read",
									"ReadAcp",
									"WriteAcp",
									"FullControl",
								}, false),
							},
						},
						"grantee": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"grantee_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Canonical",
								"Email",
								"Group",
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticTranscoderClient(ctx)

	req := &elastictranscoder.CreatePipelineInput{
		AwsKmsKeyArn:    aws.String(d.Get("aws_kms_key_arn").(string)),
		ContentConfig:   expandETPiplineOutputConfig(d, "content_config"),
		InputBucket:     aws.String(d.Get("input_bucket").(string)),
		Notifications:   expandETNotifications(d),
		Role:            aws.String(d.Get(names.AttrRole).(string)),
		ThumbnailConfig: expandETPiplineOutputConfig(d, "thumbnail_config"),
	}

	if v, ok := d.GetOk("output_bucket"); ok {
		req.OutputBucket = aws.String(v.(string))
	}

	if name, ok := d.GetOk(names.AttrName); ok {
		req.Name = aws.String(name.(string))
	} else {
		name := id.PrefixedUniqueId("tf-et-")
		d.Set(names.AttrName, name)
		req.Name = aws.String(name)
	}

	if (req.OutputBucket == nil && (req.ContentConfig == nil || req.ContentConfig.Bucket == nil)) ||
		(req.OutputBucket != nil && req.ContentConfig != nil && req.ContentConfig.Bucket != nil) {
		return sdkdiag.AppendErrorf(diags, "you must specify only one of output_bucket or content_config.bucket")
	}

	log.Printf("[DEBUG] Elastic Transcoder Pipeline create opts: %+v", req)
	resp, err := conn.CreatePipeline(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Transcoder Pipeline: %s", err)
	}

	d.SetId(aws.ToString(resp.Pipeline.Id))

	for _, w := range resp.Warnings {
		log.Printf("[WARN] Elastic Transcoder Pipeline %v: %v", *w.Code, *w.Message)
	}

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func expandETNotifications(d *schema.ResourceData) *awstypes.Notifications {
	list, ok := d.GetOk("notifications")
	if !ok {
		return nil
	}

	l := list.([]interface{})
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		log.Printf("[ERR] First element of Notifications list is nil")
		return nil
	}

	rN := l[0].(map[string]interface{})

	return &awstypes.Notifications{
		Completed:   aws.String(rN["completed"].(string)),
		Error:       aws.String(rN["error"].(string)),
		Progressing: aws.String(rN["progressing"].(string)),
		Warning:     aws.String(rN["warning"].(string)),
	}
}

func flattenETNotifications(n *awstypes.Notifications) []map[string]interface{} {
	if n == nil {
		return nil
	}

	allEmpty := func(s ...*string) bool {
		for _, s := range s {
			if aws.ToString(s) != "" {
				return false
			}
		}
		return true
	}

	// the API always returns a Notifications value, even when all fields are nil
	if allEmpty(n.Completed, n.Error, n.Progressing, n.Warning) {
		return nil
	}

	result := map[string]interface{}{
		"completed":   aws.ToString(n.Completed),
		"error":       aws.ToString(n.Error),
		"progressing": aws.ToString(n.Progressing),
		"warning":     aws.ToString(n.Warning),
	}

	return []map[string]interface{}{result}
}

func expandETPiplineOutputConfig(d *schema.ResourceData, key string) *awstypes.PipelineOutputConfig {
	list, ok := d.GetOk(key)
	if !ok {
		return nil
	}

	l := list.([]interface{})
	if len(l) == 0 {
		return nil
	}

	cc := l[0].(map[string]interface{})

	cfg := &awstypes.PipelineOutputConfig{
		Bucket:       aws.String(cc[names.AttrBucket].(string)),
		StorageClass: aws.String(cc[names.AttrStorageClass].(string)),
	}

	switch key {
	case "content_config":
		cfg.Permissions = expandETPermList(d.Get("content_config_permissions").(*schema.Set))
	case "thumbnail_config":
		cfg.Permissions = expandETPermList(d.Get("thumbnail_config_permissions").(*schema.Set))
	}

	return cfg
}

func flattenETPipelineOutputConfig(cfg *awstypes.PipelineOutputConfig) []map[string]interface{} {
	if cfg == nil {
		return nil
	}

	result := map[string]interface{}{
		names.AttrBucket:       aws.ToString(cfg.Bucket),
		names.AttrStorageClass: aws.ToString(cfg.StorageClass),
	}

	return []map[string]interface{}{result}
}

func expandETPermList(permissions *schema.Set) []awstypes.Permission {
	var perms []awstypes.Permission

	for _, p := range permissions.List() {
		if p == nil {
			continue
		}

		m := p.(map[string]interface{})

		perm := awstypes.Permission{
			Access:      flex.ExpandStringValueList(m["access"].([]interface{})),
			Grantee:     aws.String(m["grantee"].(string)),
			GranteeType: aws.String(m["grantee_type"].(string)),
		}

		perms = append(perms, perm)
	}
	return perms
}

func flattenETPermList(perms []awstypes.Permission) []map[string]interface{} {
	var set []map[string]interface{}

	for _, p := range perms {
		result := map[string]interface{}{
			"access":       flex.FlattenStringValueList(p.Access),
			"grantee":      aws.ToString(p.Grantee),
			"grantee_type": aws.ToString(p.GranteeType),
		}

		set = append(set, result)
	}
	return set
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticTranscoderClient(ctx)

	req := &elastictranscoder.UpdatePipelineInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("aws_kms_key_arn") {
		req.AwsKmsKeyArn = aws.String(d.Get("aws_kms_key_arn").(string))
	}

	if d.HasChange("content_config") {
		req.ContentConfig = expandETPiplineOutputConfig(d, "content_config")
	}

	if d.HasChange("input_bucket") {
		req.InputBucket = aws.String(d.Get("input_bucket").(string))
	}

	if d.HasChange(names.AttrName) {
		req.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("notifications") {
		req.Notifications = expandETNotifications(d)
	}

	if d.HasChange(names.AttrRole) {
		req.Role = aws.String(d.Get(names.AttrRole).(string))
	}

	if d.HasChange("thumbnail_config") {
		req.ThumbnailConfig = expandETPiplineOutputConfig(d, "thumbnail_config")
	}

	log.Printf("[DEBUG] Updating Elastic Transcoder Pipeline: %#v", req)
	output, err := conn.UpdatePipeline(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Elastic Transcoder pipeline: %s", err)
	}

	for _, w := range output.Warnings {
		log.Printf("[WARN] Elastic Transcoder Pipeline %v: %v", aws.ToString(w.Code),
			aws.ToString(w.Message))
	}

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticTranscoderClient(ctx)

	resp, err := conn.ReadPipeline(ctx, &elastictranscoder.ReadPipelineInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			log.Printf("[WARN] Elastic Transcoder Pipeline (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Elastic Transcoder Pipeline (%s): %s", d.Id(), err)
	}

	pipeline := resp.Pipeline

	d.Set(names.AttrARN, pipeline.Arn)

	d.Set("aws_kms_key_arn", pipeline.AwsKmsKeyArn)

	if pipeline.ContentConfig != nil {
		err := d.Set("content_config", flattenETPipelineOutputConfig(pipeline.ContentConfig))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting content_config: %s", err)
		}

		if pipeline.ContentConfig.Permissions != nil {
			err := d.Set("content_config_permissions", flattenETPermList(pipeline.ContentConfig.Permissions))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting content_config_permissions: %s", err)
			}
		}
	}

	d.Set("input_bucket", pipeline.InputBucket)
	d.Set(names.AttrName, pipeline.Name)

	notifications := flattenETNotifications(pipeline.Notifications)
	if err := d.Set("notifications", notifications); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting notifications: %s", err)
	}

	d.Set(names.AttrRole, pipeline.Role)

	if pipeline.ThumbnailConfig != nil {
		err := d.Set("thumbnail_config", flattenETPipelineOutputConfig(pipeline.ThumbnailConfig))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting thumbnail_config: %s", err)
		}

		if pipeline.ThumbnailConfig.Permissions != nil {
			err := d.Set("thumbnail_config_permissions", flattenETPermList(pipeline.ThumbnailConfig.Permissions))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting thumbnail_config_permissions: %s", err)
			}
		}
	}

	d.Set("output_bucket", pipeline.OutputBucket)

	return diags
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticTranscoderClient(ctx)

	log.Printf("[DEBUG] Elastic Transcoder Delete Pipeline: %s", d.Id())
	_, err := conn.DeletePipeline(ctx, &elastictranscoder.DeletePipelineInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Transcoder Pipeline: %s", err)
	}
	return diags
}
