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

func resourceAwsCloudFrontOriginRequestPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudFrontOriginRequestPolicyCreate,
		ReadContext:   resourceAwsCloudFrontOriginRequestPolicyRead,
		UpdateContext: resourceAwsCloudFrontOriginRequestPolicyUpdate,
		DeleteContext: resourceAwsCloudFrontOriginRequestPolicyDelete,
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
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.OriginRequestPolicyCookieBehavior_Values(), false),
			},
			"cookie_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"header_behavior": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.OriginRequestPolicyHeaderBehavior_Values(), false),
			},
			"header_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"query_string_behavior": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.OriginRequestPolicyQueryStringBehavior_Values(), false),
			},
			"query_string_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsCloudFrontOriginRequestPolicyCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn

	input := cloudfront.CreateOriginRequestPolicyInput{
		OriginRequestPolicyConfig: expandAwsCloudFrontOriginRequestPolicyConfig(d),
	}

	output, err := conn.CreateOriginRequestPolicyWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("create origin request policy: %s", err)
	}

	d.SetId(aws.StringValue(output.OriginRequestPolicy.Id))
	d.Set("etag", aws.StringValue(output.ETag))

	return resourceAwsCloudFrontOriginRequestPolicyRead(ctx, d, meta)
}

func resourceAwsCloudFrontOriginRequestPolicyRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Id()

	policy, etag, ok, err := getAwsCloudFrontOriginRequestPolicy(ctx, conn, id)
	switch {
	case err != nil:
		return diag.Errorf("get origin request policy %s: %s", id, err)
	case !ok:
		log.Printf("[WARN] Origin Request Policy %s not found; removing from state", id)
		d.SetId("")
		return nil
	}

	d.Set("etag", etag)

	if err := flattenAwsCloudFrontOriginRequestPolicyConfig(d, policy.OriginRequestPolicyConfig); err != nil {
		return err
	}

	return nil
}

func resourceAwsCloudFrontOriginRequestPolicyUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Id()

	_, etag, ok, err := getAwsCloudFrontOriginRequestPolicy(ctx, conn, id)
	switch {
	case err != nil:
		return diag.Errorf("get origin request policy: %s", err)
	case !ok:
		return diag.Errorf("origin request policy %s not found", id)
	}

	input := cloudfront.UpdateOriginRequestPolicyInput{
		OriginRequestPolicyConfig: expandAwsCloudFrontOriginRequestPolicyConfig(d),
		Id:                        aws.String(id),
		IfMatch:                   aws.String(etag),
	}

	output, err := conn.UpdateOriginRequestPolicyWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("update failed: %s", err)
	}

	d.Set("etag", output.ETag)

	return resourceAwsCloudFrontOriginRequestPolicyRead(ctx, d, meta)
}

func resourceAwsCloudFrontOriginRequestPolicyDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Id()

	_, etag, ok, err := getAwsCloudFrontOriginRequestPolicy(ctx, conn, id)
	switch {
	case !ok:
		log.Printf("[WARN] Origin Request Policy %s does not exist", id)
		return nil
	case err != nil:
		return diag.Errorf("failed to get etag of origin request policy %s: %s", id, err)
	}

	input := cloudfront.DeleteOriginRequestPolicyInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteOriginRequestPolicyWithContext(ctx, &input)
	switch {
	case isAWSErr(err, "NoSuchOriginRequestPolicy", ""):
		log.Printf("[WARN] Origin Request Policy %s does not exist", id)
		return nil
	case err != nil:
		return diag.Errorf("failed to delete: %s", err)
	}

	return nil
}

func getAwsCloudFrontOriginRequestPolicy(
	ctx context.Context,
	conn *cloudfront.CloudFront,
	id string,
) (policy *cloudfront.OriginRequestPolicy, etag string, ok bool, err error) {
	input := cloudfront.GetOriginRequestPolicyInput{Id: aws.String(id)}
	output, err := conn.GetOriginRequestPolicyWithContext(ctx, &input)
	switch {
	case isAWSErr(err, "NoSuchOriginRequestPolicy", ""):
		return nil, "", false, nil
	case err != nil:
		return nil, "", false, err
	}

	return output.OriginRequestPolicy, aws.StringValue(output.ETag), true, nil
}

func expandAwsCloudFrontOriginRequestPolicyConfig(d *schema.ResourceData) *cloudfront.OriginRequestPolicyConfig {
	output := &cloudfront.OriginRequestPolicyConfig{
		Comment: nil,
		CookiesConfig: &cloudfront.OriginRequestPolicyCookiesConfig{
			CookieBehavior: aws.String(d.Get("cookie_behavior").(string)),
			Cookies:        nil,
		},
		HeadersConfig: &cloudfront.OriginRequestPolicyHeadersConfig{
			HeaderBehavior: aws.String(d.Get("header_behavior").(string)),
			Headers:        nil,
		},
		Name: aws.String(d.Get("name").(string)),
		QueryStringsConfig: &cloudfront.OriginRequestPolicyQueryStringsConfig{
			QueryStringBehavior: aws.String(d.Get("query_string_behavior").(string)),
			QueryStrings:        nil,
		},
	}

	if v, ok := d.GetOk("comment"); ok && v != "" {
		output.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cookie_names"); ok {
		s := v.(*schema.Set)
		output.CookiesConfig.Cookies = &cloudfront.CookieNames{
			Items:    expandStringList(s.List()),
			Quantity: aws.Int64(int64(s.Len())),
		}
	}

	if v, ok := d.GetOk("header_names"); ok {
		s := v.(*schema.Set)
		output.HeadersConfig.Headers = &cloudfront.Headers{
			Items:    expandStringList(s.List()),
			Quantity: aws.Int64(int64(s.Len())),
		}
	}

	if v, ok := d.GetOk("query_string_names"); ok {
		s := v.(*schema.Set)
		output.QueryStringsConfig.QueryStrings = &cloudfront.QueryStringNames{
			Items:    expandStringList(s.List()),
			Quantity: aws.Int64(int64(s.Len())),
		}
	}

	return output
}

func flattenAwsCloudFrontOriginRequestPolicyConfig(
	d *schema.ResourceData,
	config *cloudfront.OriginRequestPolicyConfig,
) diag.Diagnostics {
	d.Set("comment", config.Comment)
	d.Set("name", config.Name)

	d.Set("cookie_behavior", config.CookiesConfig.CookieBehavior)
	if list := config.CookiesConfig.Cookies; list != nil {
		value := schema.NewSet(schema.HashString, flattenStringList(list.Items))
		if err := d.Set("cookie_names", value); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Set("header_behavior", config.HeadersConfig.HeaderBehavior)
	if list := config.HeadersConfig.Headers; list != nil {
		value := schema.NewSet(schema.HashString, flattenStringList(list.Items))
		if err := d.Set("header_names", value); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Set("query_string_behavior", config.QueryStringsConfig.QueryStringBehavior)
	if list := config.QueryStringsConfig.QueryStrings; list != nil {
		value := schema.NewSet(schema.HashString, flattenStringList(list.Items))
		if err := d.Set("query_string_names", value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
