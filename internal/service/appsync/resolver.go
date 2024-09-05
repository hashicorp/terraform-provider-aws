// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_resolver", name="Resolver)
func resourceResolver() *schema.Resource {
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
			names.AttrARN: {
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
			names.AttrField: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kind": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ResolverKindUnit,
				ValidateDiagFunc: enum.Validate[awstypes.ResolverKind](),
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
						names.AttrName: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RuntimeName](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConflictDetectionType](),
						},
						"conflict_handler": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConflictHandlerType](),
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
			names.AttrType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResolverCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, typeName, fieldName := d.Get("api_id").(string), d.Get(names.AttrType).(string), d.Get(names.AttrField).(string)
	id := resolverCreateResourceID(apiID, typeName, fieldName)
	input := &appsync.CreateResolverInput{
		ApiId:     aws.String(apiID),
		FieldName: aws.String(fieldName),
		Kind:      awstypes.ResolverKind(d.Get("kind").(string)),
		TypeName:  aws.String(typeName),
	}

	if v, ok := d.GetOk("caching_config"); ok {
		input.CachingConfig = expandResolverCachingConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_batch_size"); ok {
		input.MaxBatchSize = int32(v.(int))
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

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	_, err := retryResolverOp(ctx, apiID, func() (interface{}, error) {
		return conn.CreateResolver(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppSync Resolver (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceResolverRead(ctx, d, meta)...)
}

func resourceResolverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, typeName, fieldName, err := resolverParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resolver, err := findResolverByThreePartKey(ctx, conn, apiID, typeName, fieldName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync Resolver (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Resolver (%s): %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set(names.AttrARN, resolver.ResolverArn)
	if err := d.Set("caching_config", flattenCachingConfig(resolver.CachingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting caching_config: %s", err)
	}
	d.Set("code", resolver.Code)
	d.Set("data_source", resolver.DataSourceName)
	d.Set(names.AttrField, resolver.FieldName)
	d.Set("kind", resolver.Kind)
	d.Set("max_batch_size", resolver.MaxBatchSize)
	if err := d.Set("pipeline_config", flattenPipelineConfig(resolver.PipelineConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting pipeline_config: %s", err)
	}
	d.Set("request_template", resolver.RequestMappingTemplate)
	d.Set("response_template", resolver.ResponseMappingTemplate)
	if err := d.Set("runtime", flattenRuntime(resolver.Runtime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime: %s", err)
	}
	if err := d.Set("sync_config", flattenSyncConfig(resolver.SyncConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sync_config: %s", err)
	}
	d.Set(names.AttrType, resolver.TypeName)

	return diags
}

func resourceResolverUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, typeName, fieldName, err := resolverParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateResolverInput{
		ApiId:     aws.String(apiID),
		FieldName: aws.String(fieldName),
		Kind:      awstypes.ResolverKind(d.Get("kind").(string)),
		TypeName:  aws.String(typeName),
	}

	if v, ok := d.GetOk("caching_config"); ok {
		input.CachingConfig = expandResolverCachingConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_batch_size"); ok {
		input.MaxBatchSize = int32(v.(int))
	}

	if v, ok := d.GetOk("pipeline_config"); ok {
		config := v.([]interface{})[0].(map[string]interface{})
		input.PipelineConfig = &awstypes.PipelineConfig{
			Functions: flex.ExpandStringValueList(config["functions"].([]interface{})),
		}
	}

	if v, ok := d.GetOk("request_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	_, err = retryResolverOp(ctx, apiID, func() (interface{}, error) {
		return conn.UpdateResolver(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppSync Resolver (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResolverRead(ctx, d, meta)...)
}

func resourceResolverDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, typeName, fieldName, err := resolverParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Appsync Resolver: %s", d.Id())
	_, err = retryResolverOp(ctx, apiID, func() (interface{}, error) {
		return conn.DeleteResolver(ctx, &appsync.DeleteResolverInput{
			ApiId:     aws.String(apiID),
			FieldName: aws.String(fieldName),
			TypeName:  aws.String(typeName),
		})
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync Resolver (%s): %s", d.Id(), err)
	}

	return diags
}

const resolverResourceIDSeparator = "-"

func resolverCreateResourceID(apiID, typeName, fieldName string) string {
	parts := []string{apiID, typeName, fieldName}
	id := strings.Join(parts, resolverResourceIDSeparator)

	return id
}

func resolverParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, resolverResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected API-ID%[2]sTYPE-NAME%[2]sFIELD-NAME", id, resolverResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func retryResolverOp(ctx context.Context, apiID string, f func() (interface{}, error)) (interface{}, error) { //nolint:unparam
	mutexKey := "appsync-schema-" + apiID
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	const (
		timeout = 2 * time.Minute
	)
	return tfresource.RetryWhenIsA[*awstypes.ConcurrentModificationException](ctx, timeout, f)
}

func findResolverByThreePartKey(ctx context.Context, conn *appsync.Client, apiID, typeName, fieldName string) (*awstypes.Resolver, error) {
	input := &appsync.GetResolverInput{
		ApiId:     aws.String(apiID),
		FieldName: aws.String(fieldName),
		TypeName:  aws.String(typeName),
	}

	output, err := conn.GetResolver(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Resolver == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Resolver, nil
}

func expandResolverCachingConfig(tfList []interface{}) *awstypes.CachingConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.CachingConfig{
		CachingKeys: flex.ExpandStringValueSet(tfMap["caching_keys"].(*schema.Set)),
	}

	if v, ok := tfMap["ttl"].(int); ok && v != 0 {
		apiObject.Ttl = int64(v)
	}

	return apiObject
}

func expandPipelineConfig(tfList []interface{}) *awstypes.PipelineConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.PipelineConfig{}

	if v, ok := tfMap["functions"].([]interface{}); ok && len(v) > 0 {
		apiObject.Functions = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func flattenPipelineConfig(apiObject *awstypes.PipelineConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	if len(apiObject.Functions) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{
		"functions": apiObject.Functions,
	}

	return []interface{}{tfMap}
}

func flattenCachingConfig(apiObject *awstypes.CachingConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	if len(apiObject.CachingKeys) == 0 && apiObject.Ttl == 0 {
		return nil
	}

	tfMap := map[string]interface{}{
		"caching_keys": apiObject.CachingKeys,
		"ttl":          apiObject.Ttl,
	}

	return []interface{}{tfMap}
}
