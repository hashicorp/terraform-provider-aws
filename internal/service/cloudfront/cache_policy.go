package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceCachePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceCachePolicyCreate,
		Read:   resourceCachePolicyRead,
		Update: resourceCachePolicyUpdate,
		Delete: resourceCachePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
			"etag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  31536000,
			},
			"min_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parameters_in_cache_key_and_forwarded_to_origin": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookies_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookie_behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"none", "whitelist", "allExcept", "all"}, false),
									},
									"cookies": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"enable_accept_encoding_brotli": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_accept_encoding_gzip": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"headers_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header_behavior": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"none", "whitelist"}, false),
									},
									"headers": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"query_strings_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query_string_behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"none", "whitelist", "allExcept", "all"}, false),
									},
									"query_strings": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
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

func resourceCachePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	name := d.Get("name").(string)
	apiObject := &cloudfront.CachePolicyConfig{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("comment"); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_ttl"); ok {
		apiObject.DefaultTTL = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("max_ttl"); ok {
		apiObject.MaxTTL = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("min_ttl"); ok {
		apiObject.MinTTL = aws.Int64(int64(v.(int)))
	}

	// if v, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
	// 	apiObject.ParametersInCacheKeyAndForwardedToOrigin = expandParametersInCacheKeyAndForwardedToOrigin(v.([]interface{})[0].(map[string]interface{}))
	// }

	if v, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ParametersInCacheKeyAndForwardedToOrigin = expandCloudFrontCachePolicyParametersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.CreateCachePolicyInput{
		CachePolicyConfig: apiObject,
	}

	log.Printf("[DEBUG] Creating CloudFront Cache Policy: (%s)", input)
	output, err := conn.CreateCachePolicy(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFront Cache Policy (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.CachePolicy.Id))

	return resourceCachePolicyRead(d, meta)
}

func resourceCachePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	output, err := FindCachePolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Cache Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFront Cache Policy (%s): %w", d.Id(), err)
	}

	apiObject := output.CachePolicy.CachePolicyConfig
	d.Set("comment", apiObject.Comment)
	d.Set("default_ttl", apiObject.DefaultTTL)
	d.Set("etag", output.ETag)
	d.Set("max_ttl", apiObject.MaxTTL)
	d.Set("min_ttl", apiObject.MinTTL)
	d.Set("name", apiObject.Name)
	// if apiObject.ParametersInCacheKeyAndForwardedToOrigin != nil {
	// 	if err := d.Set("parameters_in_cache_key_and_forwarded_to_origin", []interface{}{flattenParametersInCacheKeyAndForwardedToOrigin(apiObject.ParametersInCacheKeyAndForwardedToOrigin)}); err != nil {
	// 		return fmt.Errorf("error setting parameters_in_cache_key_and_forwarded_to_origin: %w", err)
	// 	}
	// } else {
	// 	d.Set("parameters_in_cache_key_and_forwarded_to_origin", nil)
	// }
	d.Set("parameters_in_cache_key_and_forwarded_to_origin", setParametersConfig(apiObject.ParametersInCacheKeyAndForwardedToOrigin))

	return nil
}

func resourceCachePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateCachePolicy.html:
	// "When you update a cache policy configuration, all the fields are updated with the values provided in the request. You cannot update some fields independent of others."
	//
	apiObject := &cloudfront.CachePolicyConfig{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_ttl"); ok {
		apiObject.DefaultTTL = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("max_ttl"); ok {
		apiObject.MaxTTL = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("min_ttl"); ok {
		apiObject.MinTTL = aws.Int64(int64(v.(int)))
	}

	// if v, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
	// 	input.ParametersInCacheKeyAndForwardedToOrigin = expandParametersInCacheKeyAndForwardedToOrigin(v.([]interface{})[0].(map[string]interface{}))
	// }

	if v, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ParametersInCacheKeyAndForwardedToOrigin = expandCloudFrontCachePolicyParametersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.UpdateCachePolicyInput{
		CachePolicyConfig: apiObject,
		Id:                aws.String(d.Id()),
		IfMatch:           aws.String(d.Get("etag").(string)),
	}

	log.Printf("[DEBUG] Updating CloudFront Cache Policy: (%s)", input)
	_, err := conn.UpdateCachePolicy(input)

	if err != nil {
		return fmt.Errorf("error updating CloudFront Cache Policy (%s): %w", d.Id(), err)
	}

	return resourceCachePolicyRead(d, meta)
}

func resourceCachePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	log.Printf("[DEBUG] Deleting CloudFront Cache Policy: (%s)", d.Id())
	_, err := conn.DeleteCachePolicy(&cloudfront.DeleteCachePolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchCachePolicy) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFront Cache Policy (%s): %w", d.Id(), err)
	}

	return nil
}
