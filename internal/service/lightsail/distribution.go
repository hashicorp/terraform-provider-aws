// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_distribution", name="Distribution")
// @Tags(identifierAttribute="id", resourceType="Distribution")
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
			names.AttrARN: {
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
							ValidateFunc: validation.StringInSlice(flattenBehaviorEnumValues(types.BehaviorEnum("").Values()), false),
						},
						names.AttrPath: {
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
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`.*\S.*`), "Value must match regex: .*\\S.*"),
						},
						"cached_http_methods": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "The HTTP method responses that are cached by your distribution.",
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`.*\S.*`), "Value must match regex: .*\\S.*"),
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
										ValidateFunc: validation.StringInSlice(flattenForwardValuesValues(types.ForwardValues("").Values()), false),
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
											ValidateFunc: validation.StringInSlice(flattenHeaderEnumValues(types.HeaderEnum("").Values()), false),
										},
									},
									"option": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "The headers that you want your distribution to forward to your origin and base caching on.",
										ValidateFunc: validation.StringInSlice(enum.Slice("default", types.ForwardValuesAllowList, types.ForwardValuesAll), false),
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
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`\w[\w\-]*\w`), "Certificate name must match regex: \\w[\\w\\-]*\\w"),
			},
			names.AttrCreatedAt: {
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
							ValidateFunc: validation.StringInSlice(flattenBehaviorEnumValues(types.BehaviorEnum("").Values()), false),
						},
					},
				},
			},
			names.AttrDomainName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain name of the distribution.",
			},
			names.AttrIPAddressType: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The IP address type of the distribution.",
				ValidateFunc: validation.StringInSlice(flattenIPAddressTypeValues(types.IpAddressType("").Values()), false),
				Default:      "dualstack",
			},
			names.AttrLocation: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An object that describes the location of the distribution, such as the AWS Region and Availability Zone.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZone: {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The Availability Zone.",
							ValidateFunc: validation.StringInSlice(flattenBehaviorEnumValues(types.BehaviorEnum("").Values()), false),
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
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the distribution.",
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`\w[\w\-]*\w`), "name must match regex: \\w[\\w\\-]*\\w"),
			},
			"origin": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "An object that describes the origin resource of the distribution, such as a Lightsail instance, bucket, or load balancer.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`\w[\w\-]*\w`), "Name must match regex: \\w[\\w\\-]*\\w"),
							Description:  "The name of the origin resource.",
						},
						"protocol_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenOriginProtocolPolicyEnumValues(types.OriginProtocolPolicyEnum("").Values()), false),
							Description:  "The protocol that your Amazon Lightsail distribution uses when establishing a connection with your origin to pull content.",
						},
						"region_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidRegionName,
							Description:  "The AWS Region name of the origin resource.",
						},
						names.AttrResourceType: {
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
			names.AttrResourceType: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Lightsail resource type (e.g., Distribution).",
			},
			names.AttrStatus: {
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	in := &lightsail.CreateDistributionInput{
		BundleId:             aws.String(d.Get("bundle_id").(string)),
		DefaultCacheBehavior: expandCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DistributionName:     aws.String(d.Get(names.AttrName).(string)),
		Origin:               expandInputOrigin(d.Get("origin").([]interface{})[0].(map[string]interface{})),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cache_behavior_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.CacheBehaviorSettings = expandCacheSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cache_behavior"); ok && v.(*schema.Set).Len() > 0 {
		in.CacheBehaviors = expandCacheBehaviorsPerPath(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrIPAddressType); ok {
		in.IpAddressType = types.IpAddressType(v.(string))
	}

	if v, ok := d.GetOk("certificate_name"); ok {
		in.CertificateName = aws.String(v.(string))
	}

	out, err := conn.CreateDistribution(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionCreating, ResNameDistribution, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.Distribution == nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionCreating, ResNameDistribution, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	id := aws.ToString(out.Distribution.Name)

	diag := expandOperation(ctx, conn, *out.Operation, types.OperationTypeCreateDistribution, ResNameDistribution, id)

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
		updateOut, err := conn.UpdateDistribution(ctx, updateIn)

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, create.ErrActionUpdating, ResNameDistribution, d.Id(), err)
		}

		diagUpdate := expandOperation(ctx, conn, *updateOut.Operation, types.OperationTypeUpdateDistribution, ResNameDistribution, d.Id())

		if diagUpdate != nil {
			return diagUpdate
		}
	}

	return append(diags, resourceDistributionRead(ctx, d, meta)...)
}

func resourceDistributionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindDistributionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lightsail Distribution (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResNameDistribution, d.Id(), err)
	}

	d.Set("alternative_domain_names", out.AlternativeDomainNames)
	d.Set(names.AttrARN, out.Arn)
	d.Set("bundle_id", out.BundleId)
	if err := d.Set("cache_behavior", flattenCacheBehaviorsPerPath(out.CacheBehaviors)); err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
	}

	if out.CacheBehaviorSettings != nil {
		if err := d.Set("cache_behavior_settings", []interface{}{flattenCacheSettings(out.CacheBehaviorSettings)}); err != nil {
			return create.AppendDiagError(diags, names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
		}
	} else {
		d.Set("cache_behavior_settings", nil)
	}

	d.Set("certificate_name", out.CertificateName)
	d.Set(names.AttrCreatedAt, out.CreatedAt.Format(time.RFC3339))

	if err := d.Set("default_cache_behavior", []interface{}{flattenCacheBehavior(out.DefaultCacheBehavior)}); err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
	}
	d.Set(names.AttrDomainName, out.DomainName)
	d.Set("is_enabled", out.IsEnabled)
	d.Set(names.AttrIPAddressType, out.IpAddressType)
	d.Set(names.AttrLocation, []interface{}{flattenResourceLocation(out.Location)})
	if err := d.Set("origin", []interface{}{flattenOrigin(out.Origin)}); err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionSetting, ResNameDistribution, d.Id(), err)
	}
	d.Set(names.AttrName, out.Name)
	d.Set("origin_public_dns", out.OriginPublicDNS)
	d.Set(names.AttrResourceType, out.ResourceType)
	d.Set(names.AttrStatus, out.Status)
	d.Set("support_code", out.SupportCode)

	setTagsOut(ctx, out.Tags)

	return diags
}

func resourceDistributionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

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

	if d.HasChanges("certificate_name") {
		in.CertificateName = aws.String(d.Get("certificate_name").(string))
		update = true
	}

	if d.HasChanges("bundle_id") {
		bundleIn.BundleId = aws.String(d.Get("bundle_id").(string))
		bundleUpdate = true
	}

	if d.HasChange(names.AttrIPAddressType) {
		out, err := conn.SetIpAddressType(ctx, &lightsail.SetIpAddressTypeInput{
			ResourceName:  aws.String(d.Id()),
			ResourceType:  types.ResourceTypeDistribution,
			IpAddressType: types.IpAddressType(d.Get(names.AttrIPAddressType).(string)),
		})

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeSetIpAddressType), ResNameDistribution, d.Id(), err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeSetIpAddressType, ResNameDistribution, d.Id())

		if diag != nil {
			return diag
		}
	}

	if update {
		log.Printf("[DEBUG] Updating Lightsail Distribution (%s): %#v", d.Id(), in)
		out, err := conn.UpdateDistribution(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, create.ErrActionUpdating, ResNameDistribution, d.Id(), err)
		}

		diag := expandOperation(ctx, conn, *out.Operation, types.OperationTypeUpdateDistribution, ResNameDistribution, d.Id())

		if diag != nil {
			return diag
		}
	}

	if bundleUpdate {
		log.Printf("[DEBUG] Updating Lightsail Distribution Bundle (%s): %#v", d.Id(), in)
		out, err := conn.UpdateDistributionBundle(ctx, bundleIn)
		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, create.ErrActionUpdating, ResNameDistribution, d.Id(), err)
		}

		diag := expandOperation(ctx, conn, *out.Operation, types.OperationTypeUpdateDistributionBundle, ResNameDistribution, d.Id())

		if diag != nil {
			return diag
		}
	}

	return append(diags, resourceDistributionRead(ctx, d, meta)...)
}

func resourceDistributionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	log.Printf("[INFO] Deleting Lightsail Distribution %s", d.Id())

	out, err := conn.DeleteDistribution(ctx, &lightsail.DeleteDistributionInput{
		DistributionName: aws.String(d.Id()),
	})

	if IsANotFoundError(err) || errs.IsA[*types.InvalidInputException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionDeleting, ResNameDistribution, d.Id(), err)
	}

	diag := expandOperation(ctx, conn, *out.Operation, types.OperationTypeDeleteDistribution, ResNameDistribution, d.Id())

	if diag != nil {
		return diag
	}

	return diags
}

func FindDistributionByID(ctx context.Context, conn *lightsail.Client, id string) (*types.LightsailDistribution, error) {
	in := &lightsail.GetDistributionsInput{
		DistributionName: aws.String(id),
	}
	out, err := conn.GetDistributions(ctx, in)
	if IsANotFoundError(err) || errs.IsA[*types.InvalidInputException](err) {
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

	return &out.Distributions[0], nil
}

func flattenCookieObject(apiObject *types.CookieObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CookiesAllowList; len(v) > 0 {
		m["cookies_allow_list"] = v
	}

	if v := apiObject.Option; v != "" {
		m["option"] = v
	}

	return m
}

func flattenHeaderObject(apiObject *types.HeaderObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.HeadersAllowList; len(v) > 0 {
		m["headers_allow_list"] = v
	}

	if v := apiObject.Option; v != "" {
		m["option"] = v
	}

	return m
}

func flattenQueryStringObject(apiObject *types.QueryStringObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.QueryStringsAllowList; len(v) > 0 {
		m["query_strings_allowed_list"] = v
	}

	if v := apiObject.Option; v != nil {
		m["option"] = aws.ToBool(v)
	}

	return m
}

func flattenCacheSettings(apiObject *types.CacheSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AllowedHTTPMethods; v != nil {
		m["allowed_http_methods"] = aws.ToString(v)
	}

	if v := apiObject.CachedHTTPMethods; v != nil {
		m["cached_http_methods"] = aws.ToString(v)
	}

	if v := apiObject.DefaultTTL; v != nil {
		m["default_ttl"] = int(aws.ToInt64(v))
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
		m["maximum_ttl"] = int(aws.ToInt64(v))
	}

	if v := apiObject.MinimumTTL; v != nil {
		m["minimum_ttl"] = int(aws.ToInt64(v))
	}

	return m
}

func flattenCacheBehaviorPerPath(apiObject types.CacheBehaviorPerPath) map[string]interface{} {
	if apiObject == (types.CacheBehaviorPerPath{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Behavior; v != "" {
		m["behavior"] = v
	}

	if v := apiObject.Path; v != nil {
		m[names.AttrPath] = aws.ToString(v)
	}

	return m
}

func flattenCacheBehaviorsPerPath(apiObjects []types.CacheBehaviorPerPath) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (types.CacheBehaviorPerPath{}) {
			continue
		}

		l = append(l, flattenCacheBehaviorPerPath(apiObject))
	}

	return l
}

func flattenCacheBehavior(apiObject *types.CacheBehavior) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Behavior; v != "" {
		m["behavior"] = v
	}

	return m
}

func flattenOrigin(apiObject *types.Origin) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		m[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.ProtocolPolicy; v != "" {
		m["protocol_policy"] = v
	}

	if v := apiObject.RegionName; v != "" {
		m["region_name"] = v
	}

	if v := apiObject.ResourceType; v != "" {
		m[names.AttrResourceType] = v
	}

	return m
}

func expandInputOrigin(tfMap map[string]interface{}) *types.InputOrigin {
	if tfMap == nil {
		return nil
	}

	a := &types.InputOrigin{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		a.Name = aws.String(v)
	}

	if v, ok := tfMap["protocol_policy"].(string); ok && v != "" {
		a.ProtocolPolicy = types.OriginProtocolPolicyEnum(v)
	}

	if v, ok := tfMap["region_name"].(string); ok && v != "" {
		a.RegionName = types.RegionName(v)
	}

	return a
}

func expandCacheBehaviorPerPath(tfMap map[string]interface{}) types.CacheBehaviorPerPath {
	if tfMap == nil {
		return types.CacheBehaviorPerPath{}
	}

	a := types.CacheBehaviorPerPath{}

	if v, ok := tfMap["behavior"].(string); ok && v != "" {
		a.Behavior = types.BehaviorEnum(v)
	}

	if v, ok := tfMap[names.AttrPath].(string); ok && v != "" {
		a.Path = aws.String(v)
	}

	return a
}

func expandCacheBehaviorsPerPath(tfList []interface{}) []types.CacheBehaviorPerPath {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.CacheBehaviorPerPath

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandCacheBehaviorPerPath(m)

		if a == (types.CacheBehaviorPerPath{}) {
			continue
		}

		s = append(s, a)
	}

	return s
}

func expandAllowList(tfList []interface{}) []string {
	if len(tfList) == 0 {
		return nil
	}

	var s []string

	for _, r := range tfList {
		m, ok := r.(string)

		if !ok {
			continue
		}

		s = append(s, m)
	}

	return s
}

func expandHeaderEnumList(tfList []interface{}) []types.HeaderEnum {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.HeaderEnum

	for _, r := range tfList {
		m, ok := r.(string)

		if !ok {
			continue
		}

		s = append(s, types.HeaderEnum(m))
	}

	return s
}
func expandCookieObject(tfMap map[string]interface{}) *types.CookieObject {
	if tfMap == nil {
		return nil
	}

	a := &types.CookieObject{}

	if v, ok := tfMap["cookies_allow_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		a.CookiesAllowList = expandAllowList(v.(*schema.Set).List())
	}

	if v, ok := tfMap["option"].(string); ok && v != "" {
		a.Option = types.ForwardValues(v)
	}

	return a
}

func expandHeaderObject(tfMap map[string]interface{}) *types.HeaderObject {
	if tfMap == nil {
		return nil
	}

	a := &types.HeaderObject{}

	if v, ok := tfMap["headers_allow_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		a.HeadersAllowList = expandHeaderEnumList(v.(*schema.Set).List())
	}

	if v, ok := tfMap["option"].(string); ok && v != "" {
		a.Option = types.ForwardValues(v)
	}

	return a
}

func expandQueryStringObject(tfMap map[string]interface{}) *types.QueryStringObject {
	if tfMap == nil {
		return nil
	}

	a := &types.QueryStringObject{}

	if v, ok := tfMap["query_strings_allowed_list"]; ok && len(v.(*schema.Set).List()) > 0 {
		a.QueryStringsAllowList = expandAllowList(v.(*schema.Set).List())
	}

	if v, ok := tfMap["option"].(bool); ok {
		a.Option = aws.Bool(v)
	}

	return a
}

func expandCacheSettings(tfMap map[string]interface{}) *types.CacheSettings {
	if tfMap == nil {
		return nil
	}

	a := &types.CacheSettings{}

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

func expandCacheBehavior(tfMap map[string]interface{}) *types.CacheBehavior {
	if tfMap == nil {
		return nil
	}

	a := &types.CacheBehavior{}

	if v, ok := tfMap["behavior"].(string); ok && v != "" {
		a.Behavior = types.BehaviorEnum(v)
	}

	return a
}

func flattenForwardValuesValues(t []types.ForwardValues) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenHeaderEnumValues(t []types.HeaderEnum) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenIPAddressTypeValues(t []types.IpAddressType) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenBehaviorEnumValues(t []types.BehaviorEnum) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenOriginProtocolPolicyEnumValues(t []types.OriginProtocolPolicyEnum) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
