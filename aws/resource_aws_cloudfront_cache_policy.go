package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCloudFrontCachePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudFrontCachePolicyCreate,
		ReadContext:   resourceAwsCloudFrontCachePolicyRead,
		UpdateContext: resourceAwsCloudFrontCachePolicyUpdate,
		DeleteContext: resourceAwsCloudFrontCachePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cookie_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudfront.CachePolicyCookieBehaviorNone,
				ValidateFunc: validation.StringInSlice(cloudfront.CachePolicyCookieBehavior_Values(), false),
			},
			"cookie_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_ttl": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      86400,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"enable_accept_encoding_gzip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"header_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudfront.CachePolicyHeaderBehaviorNone,
				ValidateFunc: validation.StringInSlice(cloudfront.CachePolicyHeaderBehavior_Values(), false),
			},
			"header_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"max_ttl": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31536000,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"min_ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"query_string_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudfront.CachePolicyQueryStringBehaviorNone,
				ValidateFunc: validation.StringInSlice(cloudfront.CachePolicyQueryStringBehavior_Values(), false),
			},
			"query_string_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsCloudFrontCachePolicyCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn

	input := cloudfront.CreateCachePolicyInput{
		CachePolicyConfig: expandAwsCloudFrontCachePolicyConfig(d),
	}

	output, err := conn.CreateCachePolicyWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("create cache policy: %s", err)
	}

	d.SetId(aws.StringValue(output.CachePolicy.Id))
	d.Set("etag", aws.StringValue(output.ETag))

	return resourceAwsCloudFrontCachePolicyRead(ctx, d, meta)
}

func resourceAwsCloudFrontCachePolicyRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Id()

	cachePolicy, etag, ok, err := getAwsCloudFrontCachePolicy(ctx, conn, id)
	switch {
	case err != nil:
		return diag.Errorf("get cache policy %s: %s", id, err)
	case !ok:
		log.Printf("[WARN] Cache Policy %s not found; removing from state", id)
		d.SetId("")
		return nil
	}

	d.Set("etag", etag)

	if err := flattenAwsCloudFrontCachePolicyConfig(d, cachePolicy.CachePolicyConfig); err != nil {
		return err
	}

	return nil
}

func resourceAwsCloudFrontCachePolicyUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Id()

	_, etag, ok, err := getAwsCloudFrontCachePolicy(ctx, conn, id)
	switch {
	case err != nil:
		return diag.Errorf("get cache policy: %s", err)
	case !ok:
		return diag.Errorf("cache policy %s not found", id)
	}

	input := cloudfront.UpdateCachePolicyInput{
		CachePolicyConfig: expandAwsCloudFrontCachePolicyConfig(d),
		Id:                aws.String(id),
		IfMatch:           aws.String(etag),
	}

	output, err := conn.UpdateCachePolicyWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("update failed: %s", err)
	}

	d.Set("etag", output.ETag)

	return resourceAwsCloudFrontCachePolicyRead(ctx, d, meta)
}

func resourceAwsCloudFrontCachePolicyDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Id()

	_, etag, ok, err := getAwsCloudFrontCachePolicy(ctx, conn, id)
	switch {
	case !ok:
		log.Printf("[WARN] Cache Policy %s does not exist", id)
		return nil
	case err != nil:
		return diag.Errorf("failed to get etag of cache policy %s: %s", id, err)
	}

	input := cloudfront.DeleteCachePolicyInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteCachePolicyWithContext(ctx, &input)
	switch {
	case isAWSErr(err, "NoSuchCachePolicy", ""):
		log.Printf("[WARN] Cache Policy %s does not exist", id)
		return nil
	case err != nil:
		return diag.Errorf("failed to delete: %s", err)
	}

	return nil
}

func getAwsCloudFrontCachePolicy(
	ctx context.Context,
	conn *cloudfront.CloudFront,
	id string,
) (cachePolicy *cloudfront.CachePolicy, etag string, ok bool, err error) {
	input := cloudfront.GetCachePolicyInput{Id: aws.String(id)}
	output, err := conn.GetCachePolicyWithContext(ctx, &input)
	switch {
	case isAWSErr(err, "NoSuchCachePolicy", ""):
		return nil, "", false, nil
	case err != nil:
		return nil, "", false, err
	}

	return output.CachePolicy, aws.StringValue(output.ETag), true, nil
}

func expandAwsCloudFrontCachePolicyConfig(d *schema.ResourceData) *cloudfront.CachePolicyConfig {
	output := &cloudfront.CachePolicyConfig{
		Comment:    nil,
		DefaultTTL: aws.Int64(int64(d.Get("default_ttl").(int))),
		MaxTTL:     aws.Int64(int64(d.Get("max_ttl").(int))),
		MinTTL:     aws.Int64(int64(d.Get("min_ttl").(int))),
		Name:       aws.String(d.Get("name").(string)),
		ParametersInCacheKeyAndForwardedToOrigin: &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{
			CookiesConfig: &cloudfront.CachePolicyCookiesConfig{
				CookieBehavior: aws.String(d.Get("cookie_behavior").(string)),
				Cookies:        nil,
			},
			EnableAcceptEncodingGzip: aws.Bool(d.Get("enable_accept_encoding_gzip").(bool)),
			HeadersConfig: &cloudfront.CachePolicyHeadersConfig{
				HeaderBehavior: aws.String(d.Get("header_behavior").(string)),
				Headers:        nil,
			},
			QueryStringsConfig: &cloudfront.CachePolicyQueryStringsConfig{
				QueryStringBehavior: aws.String(d.Get("query_string_behavior").(string)),
				QueryStrings:        nil,
			},
		},
	}

	if v, ok := d.GetOk("comment"); ok && v.(string) != "" {
		output.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cookie_names"); ok {
		s := v.(*schema.Set)
		output.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies = &cloudfront.CookieNames{
			Items:    expandStringList(s.List()),
			Quantity: aws.Int64(int64(s.Len())),
		}
	}

	if v, ok := d.GetOk("header_names"); ok {
		s := v.(*schema.Set)
		output.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers = &cloudfront.Headers{
			Items:    expandStringList(s.List()),
			Quantity: aws.Int64(int64(s.Len())),
		}
	}

	if v, ok := d.GetOk("query_string_names"); ok {
		s := v.(*schema.Set)
		output.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings = &cloudfront.QueryStringNames{
			Items:    expandStringList(s.List()),
			Quantity: aws.Int64(int64(s.Len())),
		}
	}

	return output
}

func flattenAwsCloudFrontCachePolicyConfig(d *schema.ResourceData, config *cloudfront.CachePolicyConfig) diag.Diagnostics {
	d.Set("comment", config.Comment)
	d.Set("default_ttl", config.DefaultTTL)
	d.Set("max_ttl", config.MaxTTL)
	d.Set("min_ttl", config.MinTTL)
	d.Set("name", config.Name)

	if forwarded := config.ParametersInCacheKeyAndForwardedToOrigin; forwarded != nil {
		if cookies := forwarded.CookiesConfig; cookies != nil {
			d.Set("cookie_behavior", cookies.CookieBehavior)
			if list := cookies.Cookies; list != nil {
				value := schema.NewSet(schema.HashString, flattenStringList(list.Items))
				if err := d.Set("cookie_names", value); err != nil {
					return diag.FromErr(err)
				}
			}
		}

		d.Set("enable_accept_encoding_gzip", forwarded.EnableAcceptEncodingGzip)

		if headers := forwarded.HeadersConfig; headers != nil {
			d.Set("header_behavior", headers.HeaderBehavior)
			if list := headers.Headers; list != nil {
				value := schema.NewSet(schema.HashString, flattenStringList(list.Items))
				if err := d.Set("header_names", value); err != nil {
					return diag.FromErr(err)
				}
			}
		}

		if qs := forwarded.QueryStringsConfig; qs != nil {
			d.Set("query_string_behavior", qs.QueryStringBehavior)
			if list := qs.QueryStrings; list != nil {
				value := schema.NewSet(schema.HashString, flattenStringList(list.Items))
				if err := d.Set("query_string_names", value); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return nil
}
