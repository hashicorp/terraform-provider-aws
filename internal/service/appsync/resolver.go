// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_appsync_resolver")
func ResourceResolver() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResolverCreate,
		ReadWithoutTimeout:   resourceResolverRead,
		UpdateWithoutTimeout: resourceResolverUpdate,
		DeleteWithoutTimeout: resourceResolverDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"caching_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"caching_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ttl": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 3600),
						},
					},
				},
			},
			"code": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"runtime"},
				ValidateFunc: validation.StringLenBetween(1, 32768),
			},
			"data_source": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"pipeline_config"},
			},
			"field": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kind": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      appsync.ResolverKindUnit,
				ValidateFunc: validation.StringInSlice(appsync.ResolverKind_Values(), true),
			},
			"max_batch_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"pipeline_config": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"data_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"functions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"request_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"response_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"runtime": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				RequiredWith: []string{"code"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appsync.RuntimeName_Values(), false),
						},
						"runtime_version": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"sync_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conflict_detection": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appsync.ConflictDetectionType_Values(), false),
						},
						"conflict_handler": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appsync.ConflictHandlerType_Values(), false),
						},
						"lambda_conflict_handler_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lambda_conflict_handler_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResolverCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	input := &appsync.CreateResolverInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		TypeName:  aws.String(d.Get("type").(string)),
		FieldName: aws.String(d.Get("field").(string)),
		Kind:      aws.String(d.Get("kind").(string)),
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("max_batch_size"); ok {
		input.MaxBatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_config"); ok && len(v.([]interface{})) > 0 {
		input.PipelineConfig = expandPipelineConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("request_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("caching_config"); ok {
		input.CachingConfig = expandResolverCachingConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.CreateResolverWithContext(ctx, input)
	}, appsync.ErrCodeConcurrentModificationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppSync Resolver: %s", err)
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("type").(string) + "-" + d.Get("field").(string))

	return append(diags, resourceResolverRead(ctx, d, meta)...)
}

func resourceResolverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, typeName, fieldName, err := DecodeResolverID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppSync Resolver (%s): %s", d.Id(), err)
	}

	input := &appsync.GetResolverInput{
		ApiId:     aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName: aws.String(fieldName),
	}

	resp, err := conn.GetResolverWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Resolver (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppSync Resolver (%s): %s", d.Id(), err)
	}

	resolver := resp.Resolver
	d.Set("api_id", apiID)
	d.Set("arn", resolver.ResolverArn)
	d.Set("type", resolver.TypeName)
	d.Set("field", resolver.FieldName)
	d.Set("data_source", resolver.DataSourceName)
	d.Set("request_template", resolver.RequestMappingTemplate)
	d.Set("response_template", resolver.ResponseMappingTemplate)
	d.Set("kind", resolver.Kind)
	d.Set("max_batch_size", resolver.MaxBatchSize)
	d.Set("code", resolver.Code)

	if err := d.Set("sync_config", flattenSyncConfig(resolver.SyncConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sync_config: %s", err)
	}

	if err := d.Set("pipeline_config", flattenPipelineConfig(resolver.PipelineConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting pipeline_config: %s", err)
	}

	if err := d.Set("caching_config", flattenCachingConfig(resolver.CachingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting caching_config: %s", err)
	}

	if err := d.Set("runtime", flattenRuntime(resolver.Runtime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime: %s", err)
	}

	return diags
}

func resourceResolverUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	input := &appsync.UpdateResolverInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		FieldName: aws.String(d.Get("field").(string)),
		TypeName:  aws.String(d.Get("type").(string)),
		Kind:      aws.String(d.Get("kind").(string)),
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_config"); ok {
		config := v.([]interface{})[0].(map[string]interface{})
		input.PipelineConfig = &appsync.PipelineConfig{
			Functions: flex.ExpandStringList(config["functions"].([]interface{})),
		}
	}

	if v, ok := d.GetOk("request_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("caching_config"); ok {
		input.CachingConfig = expandResolverCachingConfig(v.([]interface{}))
	}

	if v, ok := d.GetOkExists("max_batch_size"); ok {
		input.MaxBatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.UpdateResolverWithContext(ctx, input)
	}, appsync.ErrCodeConcurrentModificationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppSync Resolver (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResolverRead(ctx, d, meta)...)
}

func resourceResolverDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, typeName, fieldName, err := DecodeResolverID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync Resolver (%s): %s", d.Id(), err)
	}

	input := &appsync.DeleteResolverInput{
		ApiId:     aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName: aws.String(fieldName),
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.DeleteResolverWithContext(ctx, input)
	}, appsync.ErrCodeConcurrentModificationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync Resolver (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeResolverID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "-", 3)
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format ApiID-TypeName-FieldName, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func expandResolverCachingConfig(l []interface{}) *appsync.CachingConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cachingConfig := &appsync.CachingConfig{
		CachingKeys: flex.ExpandStringSet(m["caching_keys"].(*schema.Set)),
	}

	if v, ok := m["ttl"].(int); ok && v != 0 {
		cachingConfig.Ttl = aws.Int64(int64(v))
	}

	return cachingConfig
}

func expandPipelineConfig(l []interface{}) *appsync.PipelineConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &appsync.PipelineConfig{}

	if v, ok := m["functions"].([]interface{}); ok && len(v) > 0 {
		config.Functions = flex.ExpandStringList(v)
	}

	return config
}

func flattenPipelineConfig(c *appsync.PipelineConfig) []interface{} {
	if c == nil {
		return nil
	}

	if len(c.Functions) == 0 {
		return nil
	}

	m := map[string]interface{}{
		"functions": flex.FlattenStringList(c.Functions),
	}

	return []interface{}{m}
}

func flattenCachingConfig(c *appsync.CachingConfig) []interface{} {
	if c == nil {
		return nil
	}

	if len(c.CachingKeys) == 0 && aws.Int64Value(c.Ttl) == 0 {
		return nil
	}

	m := map[string]interface{}{
		"caching_keys": flex.FlattenStringSet(c.CachingKeys),
		"ttl":          int(aws.Int64Value(c.Ttl)),
	}

	return []interface{}{m}
}
