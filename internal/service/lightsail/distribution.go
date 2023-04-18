package lightsail

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_distribution", name="Distribution")
// @Tags(identifierAttribute="id")
func ResourceDistribution() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDistributionCreate,
		ReadWithoutTimeout:   resourceDistributionRead,
		UpdateWithoutTimeout: resourceDistributionUpdate,
		DeleteWithoutTimeout: resourceDistributionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"alternative_domain_names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The alternate domain names of the distribution.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) of the distribution.",
			},
			"bundle_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The bundle ID to use for the distribution.",
			},
			"cache_behavior": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "An array of objects that describe the per-path cache behavior of the distribution.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"behavior": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The cache behavior for the specified path.",
							ValidateFunc: validation.StringInSlice(lightsail.BehaviorEnum_Values(), false),
						},
						"path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The path to a directory or file to cached, or not cache. Use an asterisk symbol to specify wildcard directories (path/to/assets/*), and file types (*.html, *jpg, *js). Directories and file paths are case-sensitive.",
						},
					},
				},
			},
			"cache_behavior_settings": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "An object that describes the cache behavior settings of the distribution.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_http_methods": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "The HTTP methods that are processed and forwarded to the distribution's origin.",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`.*\S.*`), "Value must match regex: .*\\S.*"),
						},
						"cached_http_methods": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "The HTTP method responses that are cached by your distribution.",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`.*\S.*`), "Value must match regex: .*\\S.*"),
						},
						"default_ttl": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The default amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the content has been updated.",
						},
						"forwarded_cookies": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "An object that describes the cookies that are forwarded to the origin. Your content is cached based on the cookies that are forwarded.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookies_allow_list": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "The specific cookies to forward to your distribution's origin.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"option": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "Specifies which cookies to forward to the distribution's origin for a cache behavior: all, none, or allow-list to forward only the cookies specified in the cookiesAllowList parameter.",
										ValidateFunc: validation.StringInSlice(lightsail.ForwardValues_Values(), false),
									},
								},
							},
						},
						"forwarded_headers": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "An object that describes the headers that are forwarded to the origin. Your content is cached based on the headers that are forwarded.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"headers_allow_list": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "The specific headers to forward to your distribution's origin.",
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(lightsail.HeaderEnum_Values(), false),
										},
									},
									"option": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "The headers that you want your distribution to forward to your origin and base caching on.",
										ValidateFunc: validation.StringInSlice([]string{"default", lightsail.ForwardValuesAllowList, lightsail.ForwardValuesAll}, false),
									},
								},
							},
						},
						"forwarded_query_strings": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "An object that describes the query strings that are forwarded to the origin. Your content is cached based on the query strings that are forwarded.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"option": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Indicates whether the distribution forwards and caches based on query strings.",
									},
									"query_strings_allowed_list": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "The specific query strings that the distribution forwards to the origin.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"maximum_ttl": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The maximum amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the object has been updated.",
						},
						"minimum_ttl": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The minimum amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the object has been updated.",
						},
					},
				},
			},
			"certificate_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the SSL/TLS certificate attached to the distribution, if any.",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\w[\w\-]*\w`), "Certificate name must match regex: \\w[\\w\\-]*\\w"),
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the distribution was created.",
			},
			"default_cache_behavior": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "An object that describes the default cache behavior of the distribution.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"behavior": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The cache behavior of the distribution.",
							ValidateFunc: validation.StringInSlice(lightsail.BehaviorEnum_Values(), false),
						},
					},
				},
			},
			"domain_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain name of the distribution.",
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The IP address type of the distribution.",
				ValidateFunc: validation.StringInSlice(lightsail.IpAddressType_Values(), false),
				Default:      "dualstack",
			},
			"location": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An object that describes the location of the distribution, such as the AWS Region and Availability Zone.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The Availability Zone.",
							ValidateFunc: validation.StringInSlice(lightsail.BehaviorEnum_Values(), false),
						},
						"region_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The AWS Region name.",
						},
					},
				},
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the distribution is enabled.",
				Default:     true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the distribution.",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\w[\w\-]*\w`), "name must match regex: \\w[\\w\\-]*\\w"),
			},
			"origin": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "An object that describes the origin resource of the distribution, such as a Lightsail instance, bucket, or load balancer.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`\w[\w\-]*\w`), "Name must match regex: \\w[\\w\\-]*\\w"),
							Description:  "The name of the origin resource.",
						},
						"protocol_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(lightsail.OriginProtocolPolicyEnum_Values(), false),
							Description:  "The protocol that your Amazon Lightsail distribution uses when establishing a connection with your origin to pull content.",
						},
						"region_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidRegionName,
							Description:  "The AWS Region name of the origin resource.",
						},
						"resource_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource type of the origin resource (e.g., Instance).",
						},
					},
				},
			},
			"origin_public_dns": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public DNS of the origin.",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Lightsail resource type (e.g., Distribution).",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the distribution.",
			},
			"support_code": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The support code. Include this code in your email to support when you have questions about your Lightsail distribution. This code enables our support team to look up your Lightsail information more easily.",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameDistribution = "Distribution"
)

func resourceDistributionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := &lightsail.CreateDistributionInput{
		BundleId:             aws.String(d.Get("bundle_id").(string)),
		DefaultCacheBehavior: expandCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DistributionName:     aws.String(d.Get("name").(string)),
		Origin:               expandInputOrigin(d.Get("origin").([]interface{})[0].(map[string]interface{})),
		Tags:                 GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("cache_behavior_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.CacheBehaviorSettings = expandCacheSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cache_behavior"); ok && v.(*schema.Set).Len() > 0 {
		in.CacheBehaviors = expandCacheBehaviorsPerPath(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		in.IpAddressType = aws.String(v.(string))
	}

	out, err := conn.CreateDistributionWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionCreating, ResNameDistribution, d.Get("name").(string), err)
	}

	if out == nil || out.Distribution == nil {
		return create.DiagError(names.Lightsail, create.ErrActionCreating, ResNameDistribution, d.Get("name").(string), errors.New("empty output"))
	}

	id := aws.StringValue(out.Distribution.Name)

	diag := expandOperation(ctx, conn, out.Operation, lightsail.OperationTypeCreateDistribution, ResNameDistribution, id)

	if diag != nil {
		return diag
	}

	d.SetId(id)

	isEnabled := d.Get("is_enabled").(bool)

	if !isEnabled {
		updateIn := &lightsail.UpdateDistributionInput{
			DistributionName: aws.String(id),
			IsEnabled:        aws.Bool(isEnabled),
		}
		updateOut, err := conn.UpdateDistributionWithContext(ctx, updateIn)

		if err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResNameDistribution, d.Id(), err)
		}

		diagUpdate := expandOperation(ctx, conn, updateOut.Operation, lightsail.OperationTypeUpdateDistribution, ResNameDistribution, d.Id())

		if diagUpdate != nil {
			return diagUpdate
		}
	}

	return resourceDistributionRead(ctx, d, meta)
}

func resourceDistributionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindDistributionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lightsail Distribution (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResNameDistribution, d.Id(), err)
	}

	d.Set("alternative_domain_names", out.AlternativeDomainNames)
	d.Set("arn", out.Arn)
	d.Set("bundle_id", out.BundleId)
	if err := d.Set("cache_behavior", flattenCacheBehaviorsPerPath(out.CacheBehaviors)); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
	}

	if out.CacheBehaviorSettings != nil {
		if err := d.Set("cache_behavior_settings", []interface{}{flattenCacheSettings(out.CacheBehaviorSettings)}); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
		}
	} else {
		d.Set("cache_behavior_settings", nil)
	}

	d.Set("certificate_name", out.CertificateName)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))

	if err := d.Set("default_cache_behavior", []interface{}{flattenCacheBehavior(out.DefaultCacheBehavior)}); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
	}
	d.Set("domain_name", out.DomainName)
	d.Set("is_enabled", out.IsEnabled)
	d.Set("ip_address_type", out.IpAddressType)
	d.Set("location", []interface{}{flattenResourceLocation(out.Location)})
	if err := d.Set("origin", []interface{}{flattenOrigin(out.Origin)}); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
	}
	d.Set("name", out.Name)
	d.Set("origin_public_dns", out.OriginPublicDNS)
	d.Set("resource_type", out.ResourceType)
	d.Set("status", out.Status)
	d.Set("support_code", out.SupportCode)

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceDistributionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	update := false
	bundleUpdate := false

	in := &lightsail.UpdateDistributionInput{
		DistributionName: aws.String(d.Id()),
	}

	bundleIn := &lightsail.UpdateDistributionBundleInput{
		DistributionName: aws.String(d.Id()),
	}

	if d.HasChanges("cache_behavior_settings") {
		in.CacheBehaviorSettings = expandCacheSettings(d.Get("cache_behavior_settings").([]interface{})[0].(map[string]interface{}))
		update = true
	}

	if d.HasChanges("cache_behavior") {
		in.CacheBehaviors = expandCacheBehaviorsPerPath(d.Get("cache_behavior").(*schema.Set).List())
		update = true
	}

	if d.HasChanges("default_cache_behavior") {
		in.DefaultCacheBehavior = expandCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{}))
		update = true
	}

	if d.HasChanges("is_enabled") {
		in.IsEnabled = aws.Bool(d.Get("is_enabled").(bool))
		update = true
	}

	if d.HasChanges("origin") {
		in.Origin = expandInputOrigin(d.Get("origin").([]interface{})[0].(map[string]interface{}))
		update = true
	}

	if d.HasChanges("bundle_id") {
		bundleIn.BundleId = aws.String(d.Get("bundle_id").(string))
		bundleUpdate = true
	}

	if d.HasChange("ip_address_type") {
		out, err := conn.SetIpAddressTypeWithContext(ctx, &lightsail.SetIpAddressTypeInput{
			ResourceName:  aws.String(d.Id()),
			ResourceType:  aws.String("Distribution"),
			IpAddressType: aws.String(d.Get("ip_address_type").(string)),
		})

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeSetIpAddressType, ResNameDistribution, d.Id(), err)
		}

		diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeSetIpAddressType, ResNameDistribution, d.Id())

		if diag != nil {
			return diag
		}
	}

	if update {
		log.Printf("[DEBUG] Updating Lightsail Distribution (%s): %#v", d.Id(), in)
		out, err := conn.UpdateDistributionWithContext(ctx, in)
		if err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResNameDistribution, d.Id(), err)
		}

		diag := expandOperation(ctx, conn, out.Operation, lightsail.OperationTypeUpdateDistribution, ResNameDistribution, d.Id())

		if diag != nil {
			return diag
		}
	}

	if bundleUpdate {
		log.Printf("[DEBUG] Updating Lightsail Distribution Bundle (%s): %#v", d.Id(), in)
		out, err := conn.UpdateDistributionBundleWithContext(ctx, bundleIn)
		if err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResNameDistribution, d.Id(), err)
		}

		diag := expandOperation(ctx, conn, out.Operation, lightsail.OperationTypeUpdateDistributionBundle, ResNameDistribution, d.Id())

		if diag != nil {
			return diag
		}
	}

	return resourceDistributionRead(ctx, d, meta)
}

func resourceDistributionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	log.Printf("[INFO] Deleting Lightsail Distribution %s", d.Id())

	out, err := conn.DeleteDistributionWithContext(ctx, &lightsail.DeleteDistributionInput{
		DistributionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) || tfawserr.ErrMessageContains(err, lightsail.ErrCodeInvalidInputException, "Requested resource not found") {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionDeleting, ResNameDistribution, d.Id(), err)
	}

	diag := expandOperation(ctx, conn, out.Operation, lightsail.OperationTypeDeleteDistribution, ResNameDistribution, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}

func FindDistributionByID(ctx context.Context, conn *lightsail.Lightsail, id string) (*lightsail.LightsailDistribution, error) {
	in := &lightsail.GetDistributionsInput{
		DistributionName: aws.String(id),
	}
	out, err := conn.GetDistributionsWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) || tfawserr.ErrMessageContains(err, lightsail.ErrCodeInvalidInputException, "Requested resource not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Distributions) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Distributions[0], nil
}

func flattenCookieObject(apiObject *lightsail.CookieObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CookiesAllowList; len(v) > 0 {
		m["cookies_allow_list"] = v
	}

	if v := apiObject.Option; v != nil {
		m["option"] = aws.StringValue(v)
	}

	return m
}

func flattenHeaderObject(apiObject *lightsail.HeaderObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.HeadersAllowList; len(v) > 0 {
		m["headers_allow_list"] = v
	}

	if v := apiObject.Option; v != nil {
		m["option"] = aws.StringValue(v)
	}

	return m
}

func flattenQueryStringObject(apiObject *lightsail.QueryStringObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.QueryStringsAllowList; len(v) > 0 {
		m["query_strings_allowed_list"] = v
	}

	if v := apiObject.Option; v != nil {
		m["option"] = aws.BoolValue(v)
	}

	return m
}

func flattenCacheSettings(apiObject *lightsail.CacheSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AllowedHTTPMethods; v != nil {
		m["allowed_http_methods"] = aws.StringValue(v)
	}

	if v := apiObject.CachedHTTPMethods; v != nil {
		m["cached_http_methods"] = aws.StringValue(v)
	}

	if v := apiObject.DefaultTTL; v != nil {
		m["default_ttl"] = int(aws.Int64Value(v))
	}

	if v := apiObject.ForwardedCookies; v != nil {
		m["forwarded_cookies"] = []interface{}{flattenCookieObject(v)}
	}

	if v := apiObject.ForwardedHeaders; v != nil {
		m["forwarded_headers"] = []interface{}{flattenHeaderObject(v)}
	}

	if v := apiObject.ForwardedQueryStrings; v != nil {
		m["forwarded_query_strings"] = []interface{}{flattenQueryStringObject(v)}
	}

	if v := apiObject.MaximumTTL; v != nil {
		m["maximum_ttl"] = int(aws.Int64Value(v))
	}

	if v := apiObject.MinimumTTL; v != nil {
		m["minimum_ttl"] = int(aws.Int64Value(v))
	}

	return m
}

func flattenCacheBehaviorPerPath(apiObject *lightsail.CacheBehaviorPerPath) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Behavior; v != nil {
		m["behavior"] = aws.StringValue(v)
	}

	if v := apiObject.Path; v != nil {
		m["path"] = aws.StringValue(v)
	}

	return m
}

func flattenCacheBehaviorsPerPath(apiObjects []*lightsail.CacheBehaviorPerPath) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenCacheBehaviorPerPath(apiObject))
	}

	return l
}

func flattenCacheBehavior(apiObject *lightsail.CacheBehavior) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Behavior; v != nil {
		m["behavior"] = aws.StringValue(v)
	}

	return m
}

func flattenOrigin(apiObject *lightsail.Origin) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		m["name"] = aws.StringValue(v)
	}

	if v := apiObject.ProtocolPolicy; v != nil {
		m["protocol_policy"] = aws.StringValue(v)
	}

	if v := apiObject.RegionName; v != nil {
		m["region_name"] = aws.StringValue(v)
	}

	if v := apiObject.ResourceType; v != nil {
		m["resource_type"] = aws.StringValue(v)
	}

	return m
}

func expandInputOrigin(tfMap map[string]interface{}) *lightsail.InputOrigin {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.InputOrigin{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		a.Name = aws.String(v)
	}

	if v, ok := tfMap["protocol_policy"].(string); ok && v != "" {
		a.ProtocolPolicy = aws.String(v)
	}

	if v, ok := tfMap["region_name"].(string); ok && v != "" {
		a.RegionName = aws.String(v)
	}

	return a
}

func expandCacheBehaviorPerPath(tfMap map[string]interface{}) *lightsail.CacheBehaviorPerPath {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.CacheBehaviorPerPath{}

	if v, ok := tfMap["behavior"].(string); ok && v != "" {
		a.Behavior = aws.String(v)
	}

	if v, ok := tfMap["path"].(string); ok && v != "" {
		a.Path = aws.String(v)
	}

	return a
}

func expandCacheBehaviorsPerPath(tfList []interface{}) []*lightsail.CacheBehaviorPerPath {
	if len(tfList) == 0 {
		return nil
	}

	var s []*lightsail.CacheBehaviorPerPath

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandCacheBehaviorPerPath(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}

func expandAllowList(tfList []interface{}) []*string {
	if len(tfList) == 0 {
		return nil
	}

	var s []*string

	for _, r := range tfList {
		m, ok := r.(string)

		if !ok {
			continue
		}

		s = append(s, aws.String(m))
	}

	return s
}

func expandCookieObject(tfMap map[string]interface{}) *lightsail.CookieObject {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.CookieObject{}

	if v, ok := tfMap["cookies_allow_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		a.CookiesAllowList = expandAllowList(v.(*schema.Set).List())
	}

	if v, ok := tfMap["option"].(string); ok && v != "" {
		a.Option = aws.String(v)
	}

	return a
}

func expandHeaderObject(tfMap map[string]interface{}) *lightsail.HeaderObject {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.HeaderObject{}

	if v, ok := tfMap["headers_allow_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		a.HeadersAllowList = expandAllowList(v.(*schema.Set).List())
	}

	if v, ok := tfMap["option"].(string); ok && v != "" {
		a.Option = aws.String(v)
	}

	return a
}

func expandQueryStringObject(tfMap map[string]interface{}) *lightsail.QueryStringObject {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.QueryStringObject{}

	if v, ok := tfMap["query_strings_allowed_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		a.QueryStringsAllowList = expandAllowList(v.(*schema.Set).List())
	}

	if v, ok := tfMap["option"].(bool); ok {
		a.Option = aws.Bool(v)
	}

	return a
}

func expandCacheSettings(tfMap map[string]interface{}) *lightsail.CacheSettings {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.CacheSettings{}

	if v, ok := tfMap["allowed_http_methods"].(string); ok && v != "" {
		a.AllowedHTTPMethods = aws.String(v)
	}

	if v, ok := tfMap["cached_http_methods"].(string); ok && v != "" {
		a.CachedHTTPMethods = aws.String(v)
	}

	if v, ok := tfMap["default_ttl"].(int); ok && v != 0 {
		a.DefaultTTL = aws.Int64(int64(v))
	}

	if v, ok := tfMap["forwarded_cookies"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.ForwardedCookies = expandCookieObject(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["forwarded_headers"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.ForwardedHeaders = expandHeaderObject(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["forwarded_query_strings"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.ForwardedQueryStrings = expandQueryStringObject(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_ttl"].(int); ok && v != 0 {
		a.MaximumTTL = aws.Int64(int64(v))
	}

	if v, ok := tfMap["minimum_ttl"].(int); ok && v != 0 {
		a.MinimumTTL = aws.Int64(int64(v))
	}

	return a
}

func expandCacheBehavior(tfMap map[string]interface{}) *lightsail.CacheBehavior {
	if tfMap == nil {
		return nil
	}

	a := &lightsail.CacheBehavior{}

	if v, ok := tfMap["behavior"].(string); ok && v != "" {
		a.Behavior = aws.String(v)
	}

	return a
}
