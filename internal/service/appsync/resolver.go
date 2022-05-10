package appsync

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResolver() *schema.Resource {
	return &schema.Resource{
		Create: resourceResolverCreate,
		Read:   resourceResolverRead,
		Update: resourceResolverUpdate,
		Delete: resourceResolverDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"field": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_source": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"pipeline_config"},
			},
			"max_batch_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"request_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"response_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kind": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      appsync.ResolverKindUnit,
				ValidateFunc: validation.StringInSlice(appsync.ResolverKind_Values(), true),
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
							Type:     schema.TypeInt,
							Optional: true,
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResolverCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.CreateResolverInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		TypeName:  aws.String(d.Get("type").(string)),
		FieldName: aws.String(d.Get("field").(string)),
		Kind:      aws.String(d.Get("kind").(string)),
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

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := verify.RetryOnAWSCode(appsync.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.CreateResolver(input)
	})

	if err != nil {
		return fmt.Errorf("error creating AppSync Resolver: %w", err)
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("type").(string) + "-" + d.Get("field").(string))

	return resourceResolverRead(d, meta)
}

func resourceResolverRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, typeName, fieldName, err := DecodeResolverID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.GetResolverInput{
		ApiId:     aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName: aws.String(fieldName),
	}

	resp, err := conn.GetResolver(input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Resolver (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppSync Resolver (%s): %w", d.Id(), err)
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

	if err := d.Set("sync_config", flattenSyncConfig(resolver.SyncConfig)); err != nil {
		return fmt.Errorf("error setting sync_config: %w", err)
	}

	if err := d.Set("pipeline_config", flattenPipelineConfig(resolver.PipelineConfig)); err != nil {
		return fmt.Errorf("Error setting pipeline_config: %w", err)
	}

	if err := d.Set("caching_config", flattenCachingConfig(resolver.CachingConfig)); err != nil {
		return fmt.Errorf("Error setting caching_config: %w", err)
	}

	return nil
}

func resourceResolverUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.UpdateResolverInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		FieldName: aws.String(d.Get("field").(string)),
		TypeName:  aws.String(d.Get("type").(string)),
		Kind:      aws.String(d.Get("kind").(string)),
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

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := verify.RetryOnAWSCode(appsync.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.UpdateResolver(input)
	})

	if err != nil {
		return fmt.Errorf("error updating AppSync Resolver (%s): %s", d.Id(), err)
	}

	return resourceResolverRead(d, meta)
}

func resourceResolverDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, typeName, fieldName, err := DecodeResolverID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.DeleteResolverInput{
		ApiId:     aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName: aws.String(fieldName),
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err = verify.RetryOnAWSCode(appsync.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.DeleteResolver(input)
	})

	if err != nil {
		return fmt.Errorf("error deleting AppSync Resolver (%s): %s", d.Id(), err)
	}

	return nil
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
