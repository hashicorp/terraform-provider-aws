package elbv2

import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	loadBalancerTagPropagationTimeout = 2 * time.Minute
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerCreate,
		ReadWithoutTimeout:   resourceLoadBalancerRead,
		UpdateWithoutTimeout: resourceLoadBalancerUpdate,
		DeleteWithoutTimeout: resourceLoadBalancerDelete,
		// Subnets are ForceNew for Network Load Balancers
		CustomizeDiff: customdiff.Sequence(
			customizeDiffNLBSubnets,
			verify.SetTagsDiff,
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validName,
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validNamePrefix,
			},

			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"load_balancer_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      elbv2.LoadBalancerTypeEnumApplication,
				ValidateFunc: validation.StringInSlice(elbv2.LoadBalancerTypeEnum_Values(), false),
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
				Set:      schema.HashString,
			},

			"subnets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
			},

			"subnet_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ipv6_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsIPv6Address,
						},
						"outpost_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"allocation_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"private_ipv4_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
					if m["allocation_id"] != "" {
						buf.WriteString(fmt.Sprintf("%s-", m["allocation_id"].(string)))
					}
					if m["private_ipv4_address"] != "" {
						buf.WriteString(fmt.Sprintf("%s-", m["private_ipv4_address"].(string)))
					}
					return create.StringHashcode(buf.String())
				},
			},

			"access_logs": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("access_logs.0.enabled").(bool)
							},
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("access_logs.0.enabled").(bool)
							},
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"idle_timeout": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          60,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumNetwork),
			},

			"drop_invalid_header_fields": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBType("network"),
			},

			"preserve_host_header": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},

			"enable_cross_zone_load_balancing": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumApplication),
			},

			"enable_http2": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          true,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumNetwork),
			},

			"enable_waf_fail_open": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumNetwork),
			},

			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					elbv2.IpAddressTypeIpv4,
					elbv2.IpAddressTypeDualstack,
				}, false),
			},

			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"desync_mitigation_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "defensive",
				ValidateFunc: validation.StringInSlice([]string{
					"monitor",
					"defensive",
					"strictest",
				}, false),
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func suppressIfLBType(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return d.Get("load_balancer_type").(string) == t
	}
}

func suppressIfLBTypeNot(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return d.Get("load_balancer_type").(string) != t
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.PrefixedUniqueId("tf-lb-")
	}
	d.Set("name", name)

	elbOpts := &elbv2.CreateLoadBalancerInput{
		Name: aws.String(name),
		Type: aws.String(d.Get("load_balancer_type").(string)),
	}

	if len(tags) > 0 {
		elbOpts.Tags = Tags(tags.IgnoreAWS())
	}

	if _, ok := d.GetOk("internal"); ok {
		elbOpts.Scheme = aws.String("internal")
	}

	if v, ok := d.GetOk("security_groups"); ok {
		elbOpts.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnets"); ok {
		elbOpts.Subnets = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_mapping"); ok {
		rawMappings := v.(*schema.Set).List()
		elbOpts.SubnetMappings = make([]*elbv2.SubnetMapping, len(rawMappings))
		for i, mapping := range rawMappings {
			subnetMap := mapping.(map[string]interface{})

			elbOpts.SubnetMappings[i] = &elbv2.SubnetMapping{
				SubnetId: aws.String(subnetMap["subnet_id"].(string)),
			}

			if subnetMap["allocation_id"].(string) != "" {
				elbOpts.SubnetMappings[i].AllocationId = aws.String(subnetMap["allocation_id"].(string))
			}

			if subnetMap["private_ipv4_address"].(string) != "" {
				elbOpts.SubnetMappings[i].PrivateIPv4Address = aws.String(subnetMap["private_ipv4_address"].(string))
			}

			if subnetMap["ipv6_address"].(string) != "" {
				elbOpts.SubnetMappings[i].IPv6Address = aws.String(subnetMap["ipv6_address"].(string))
			}
		}
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		elbOpts.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		elbOpts.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	log.Printf("[DEBUG] ALB create configuration: %#v", elbOpts)

	resp, err := conn.CreateLoadBalancerWithContext(ctx, elbOpts)

	// Some partitions may not support tag-on-create
	if elbOpts.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] ELBv2 Load Balancer (%s) create failed (%s) with tags. Trying create without tags.", name, err)
		elbOpts.Tags = nil
		resp, err = conn.CreateLoadBalancerWithContext(ctx, elbOpts)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating %s Load Balancer: %s", d.Get("load_balancer_type").(string), err)
	}

	if len(resp.LoadBalancers) != 1 {
		return sdkdiag.AppendErrorf(diags, "no load balancers returned following creation of %s", d.Get("name").(string))
	}

	lb := resp.LoadBalancers[0]
	d.SetId(aws.StringValue(lb.LoadBalancerArn))
	log.Printf("[INFO] LB ID: %s", d.Id())

	_, err = waitLoadBalancerActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Load Balancer (%s) to be active: %s", d.Get("name").(string), err)
	}

	// Post-create tagging supported in some partitions
	if elbOpts.Tags == nil && len(tags) > 0 {
		err := UpdateTags(ctx, conn, d.Id(), nil, tags)

		// if default tags only, log and continue (i.e., should error if explicitly setting tags and they can't be)
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] error adding tags after create for ELBv2 Load Balancer (%s): %s", d.Id(), err)
			return append(diags, resourceLoadBalancerUpdate(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ELBv2 Load Balancer (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLoadBalancerUpdate(ctx, d, meta)...)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn()

	lb, err := FindLoadBalancerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Load Balancer %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Load Balancer (%s): %s", d.Id(), err)
	}

	if err := flattenResource(ctx, d, meta, lb); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn()

	attributes := make([]*elbv2.LoadBalancerAttribute, 0)

	if d.HasChange("access_logs") {
		logs := d.Get("access_logs").([]interface{})

		if len(logs) == 1 && logs[0] != nil {
			log := logs[0].(map[string]interface{})

			enabled := log["enabled"].(bool)

			attributes = append(attributes,
				&elbv2.LoadBalancerAttribute{
					Key:   aws.String("access_logs.s3.enabled"),
					Value: aws.String(strconv.FormatBool(enabled)),
				})
			if enabled {
				attributes = append(attributes,
					&elbv2.LoadBalancerAttribute{
						Key:   aws.String("access_logs.s3.bucket"),
						Value: aws.String(log["bucket"].(string)),
					},
					&elbv2.LoadBalancerAttribute{
						Key:   aws.String("access_logs.s3.prefix"),
						Value: aws.String(log["prefix"].(string)),
					})
			}
		} else {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("access_logs.s3.enabled"),
				Value: aws.String("false"),
			})
		}
	}

	switch d.Get("load_balancer_type").(string) {
	case elbv2.LoadBalancerTypeEnumApplication:
		if d.HasChange("idle_timeout") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("idle_timeout.timeout_seconds"),
				Value: aws.String(fmt.Sprintf("%d", d.Get("idle_timeout").(int))),
			})
		}

		if d.HasChange("enable_http2") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("routing.http2.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("enable_http2").(bool))),
			})
		}

		// The "waf.fail_open.enabled" attribute is not available in all AWS regions
		// e.g. us-gov-east-1; thus, we can instead only modify the attribute as a result of d.HasChange()
		// to avoid "ValidationError: Load balancer attribute key 'waf.fail_open.enabled' is not recognized"
		// when modifying the attribute right after resource creation.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/22037
		if d.HasChange("enable_waf_fail_open") {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("waf.fail_open.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("enable_waf_fail_open").(bool))),
			})
		}

		if d.HasChange("drop_invalid_header_fields") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("routing.http.drop_invalid_header_fields.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("drop_invalid_header_fields").(bool))),
			})
		}

		if d.HasChange("preserve_host_header") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("routing.http.preserve_host_header.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("preserve_host_header").(bool))),
			})
		}

		if d.HasChange("desync_mitigation_mode") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("routing.http.desync_mitigation_mode"),
				Value: aws.String(d.Get("desync_mitigation_mode").(string)),
			})
		}

	case elbv2.LoadBalancerTypeEnumGateway, elbv2.LoadBalancerTypeEnumNetwork:
		if d.HasChange("enable_cross_zone_load_balancing") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("load_balancing.cross_zone.enabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("enable_cross_zone_load_balancing").(bool))),
			})
		}
	}

	if d.HasChange("enable_deletion_protection") || d.IsNewResource() {
		attributes = append(attributes, &elbv2.LoadBalancerAttribute{
			Key:   aws.String("deletion_protection.enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("enable_deletion_protection").(bool))),
		})
	}

	if len(attributes) != 0 {
		input := &elbv2.ModifyLoadBalancerAttributesInput{
			LoadBalancerArn: aws.String(d.Id()),
			Attributes:      attributes,
		}

		log.Printf("[DEBUG] ALB Modify Load Balancer Attributes Request: %#v", input)

		// Not all attributes are supported in all partitions (e.g., ISO)
		var err error
		for {
			_, err = conn.ModifyLoadBalancerAttributesWithContext(ctx, input)
			if err == nil {
				break
			}

			re := regexp.MustCompile(`attribute key ('|")?([^'" ]+)('|")? is not recognized`)
			if sm := re.FindStringSubmatch(err.Error()); len(sm) > 1 {
				log.Printf("[WARN] failed to modify Load Balancer (%s), unsupported attribute (%s): %s", d.Id(), sm[2], err)
				input.Attributes = removeAttribute(input.Attributes, sm[2])
				continue
			}

			break
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "failure configuring LB attributes: %s", err)
		}
	}

	if d.HasChange("security_groups") {
		sgs := flex.ExpandStringSet(d.Get("security_groups").(*schema.Set))

		params := &elbv2.SetSecurityGroupsInput{
			LoadBalancerArn: aws.String(d.Id()),
			SecurityGroups:  sgs,
		}
		_, err := conn.SetSecurityGroupsWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "failure Setting LB Security Groups: %s", err)
		}
	}

	// subnets are assigned at Create; the 'change' here is an empty map for old
	// and current subnets for new, so this change is redundant when the
	// resource is just created, so we don't attempt if it is a newly created
	// resource.
	if d.HasChange("subnets") && !d.IsNewResource() {
		subnets := flex.ExpandStringSet(d.Get("subnets").(*schema.Set))

		params := &elbv2.SetSubnetsInput{
			LoadBalancerArn: aws.String(d.Id()),
			Subnets:         subnets,
		}

		_, err := conn.SetSubnetsWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "failure Setting LB Subnets: %s", err)
		}
	}

	if d.HasChange("ip_address_type") {
		params := &elbv2.SetIpAddressTypeInput{
			LoadBalancerArn: aws.String(d.Id()),
			IpAddressType:   aws.String(d.Get("ip_address_type").(string)),
		}

		_, err := conn.SetIpAddressTypeWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "failure Setting LB IP Address Type: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := resource.RetryContext(ctx, loadBalancerTagPropagationTimeout, func() *resource.RetryError {
			err := UpdateTags(ctx, conn, d.Id(), o, n)

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
				log.Printf("[DEBUG] Retrying tagging of LB (%s) after error: %s", d.Id(), err)
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			err = UpdateTags(ctx, conn, d.Id(), o, n)
		}

		// ISO partitions may not support tagging, giving error
		if verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] Unable to update tags for ELBv2 Load Balancer %s: %s", d.Id(), err)

			_, err := waitLoadBalancerActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Load Balancer (%s) to be active: %s", d.Get("name").(string), err)
			}

			return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating LB (%s) tags: %s", d.Id(), err)
		}
	}

	_, err := waitLoadBalancerActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Load Balancer (%s) to be active: %s", d.Get("name").(string), err)
	}

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn()

	log.Printf("[INFO] Deleting LB: %s", d.Id())

	// Destroy the load balancer
	deleteElbOpts := elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: aws.String(d.Id()),
	}
	if _, err := conn.DeleteLoadBalancerWithContext(ctx, &deleteElbOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting LB: %s", err)
	}

	ec2conn := meta.(*conns.AWSClient).EC2Conn()

	err := cleanupALBNetworkInterfaces(ctx, ec2conn, d.Id())
	if err != nil {
		log.Printf("[WARN] Failed to cleanup ENIs for ALB %q: %#v", d.Id(), err)
	}

	err = waitForNLBNetworkInterfacesToDetach(ctx, ec2conn, d.Id())
	if err != nil {
		log.Printf("[WARN] Failed to wait for ENIs to disappear for NLB %q: %#v", d.Id(), err)
	}

	return diags
}

func FindLoadBalancerByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.LoadBalancer, error) {
	input := &elbv2.DescribeLoadBalancersInput{
		LoadBalancerArns: aws.StringSlice([]string{arn}),
	}

	output, err := FindLoadBalancer(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.LoadBalancerArn) != arn {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindLoadBalancers(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeLoadBalancersInput) ([]*elbv2.LoadBalancer, error) {
	var output []*elbv2.LoadBalancer

	err := conn.DescribeLoadBalancersPagesWithContext(ctx, input, func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LoadBalancers {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLoadBalancer(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeLoadBalancersInput) (*elbv2.LoadBalancer, error) {
	output, err := FindLoadBalancers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].State == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func statusLoadBalancerState(ctx context.Context, conn *elbv2.ELBV2, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLoadBalancerByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State.Code), nil
	}
}

func waitLoadBalancerActive(ctx context.Context, conn *elbv2.ELBV2, arn string, timeout time.Duration) (*elbv2.LoadBalancer, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending:    []string{elbv2.LoadBalancerStateEnumProvisioning, elbv2.LoadBalancerStateEnumFailed},
		Target:     []string{elbv2.LoadBalancerStateEnumActive},
		Refresh:    statusLoadBalancerState(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elbv2.LoadBalancer); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.State.Reason)))

		return output, err
	}

	return nil, err
}

func removeAttribute(attributes []*elbv2.LoadBalancerAttribute, key string) []*elbv2.LoadBalancerAttribute {
	for i, a := range attributes {
		if aws.StringValue(a.Key) == key {
			return append(attributes[:i], attributes[i+1:]...)
		}
	}

	log.Printf("[WARN] Unable to remove attribute %s from Load Balancer attributes: not found", key)
	return attributes
}

// ALB automatically creates ENI(s) on creation
// but the cleanup is asynchronous and may take time
// which then blocks IGW, SG or VPC on deletion
// So we make the cleanup "synchronous" here
func cleanupALBNetworkInterfaces(ctx context.Context, conn *ec2.EC2, lbArn string) error {
	name, err := getLBNameFromARN(lbArn)

	if err != nil {
		return err
	}

	networkInterfaces, err := tfec2.FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(ctx, conn, "amazon-elb", "ELB "+name)

	if err != nil {
		return err
	}

	var errs *multierror.Error

	for _, networkInterface := range networkInterfaces {
		if networkInterface.Attachment == nil {
			continue
		}

		attachmentID := aws.StringValue(networkInterface.Attachment.AttachmentId)
		networkInterfaceID := aws.StringValue(networkInterface.NetworkInterfaceId)

		err = tfec2.DetachNetworkInterface(ctx, conn, networkInterfaceID, attachmentID, tfec2.NetworkInterfaceDetachedTimeout)

		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}

		err = tfec2.DeleteNetworkInterface(ctx, conn, networkInterfaceID)

		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}
	}

	return errs.ErrorOrNil()
}

func waitForNLBNetworkInterfacesToDetach(ctx context.Context, conn *ec2.EC2, lbArn string) error {
	const (
		loadBalancerNetworkInterfaceDetachTimeout = 5 * time.Minute
	)
	name, err := getLBNameFromARN(lbArn)

	if err != nil {
		return err
	}

	errAttached := errors.New("attached")

	_, err = tfresource.RetryWhen(ctx, loadBalancerNetworkInterfaceDetachTimeout,
		func() (interface{}, error) {
			networkInterfaces, err := tfec2.FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(ctx, conn, "amazon-aws", "ELB "+name)

			if err != nil {
				return nil, err
			}

			if len(networkInterfaces) > 0 {
				return networkInterfaces, errAttached
			}

			return networkInterfaces, nil
		},
		func(err error) (bool, error) {
			if errors.Is(err, errAttached) {
				return true, err
			}

			return false, err
		},
	)

	return err
}

func getLBNameFromARN(arn string) (string, error) {
	re := regexp.MustCompile("([^/]+/[^/]+/[^/]+)$")
	matches := re.FindStringSubmatch(arn)
	if len(matches) != 2 {
		return "", fmt.Errorf("unexpected ARN format: %q", arn)
	}

	// e.g. app/example-alb/b26e625cdde161e6
	return matches[1], nil
}

// flattenSubnetsFromAvailabilityZones creates a slice of strings containing the subnet IDs
// for the ALB based on the AvailabilityZones structure returned by the API.
func flattenSubnetsFromAvailabilityZones(availabilityZones []*elbv2.AvailabilityZone) []string {
	var result []string
	for _, az := range availabilityZones {
		result = append(result, aws.StringValue(az.SubnetId))
	}
	return result
}

func flattenSubnetMappingsFromAvailabilityZones(availabilityZones []*elbv2.AvailabilityZone) []map[string]interface{} {
	l := make([]map[string]interface{}, 0)
	for _, availabilityZone := range availabilityZones {
		m := make(map[string]interface{})
		m["subnet_id"] = aws.StringValue(availabilityZone.SubnetId)
		m["outpost_id"] = aws.StringValue(availabilityZone.OutpostId)

		for _, loadBalancerAddress := range availabilityZone.LoadBalancerAddresses {
			m["allocation_id"] = aws.StringValue(loadBalancerAddress.AllocationId)
			m["private_ipv4_address"] = aws.StringValue(loadBalancerAddress.PrivateIPv4Address)
			m["ipv6_address"] = aws.StringValue(loadBalancerAddress.IPv6Address)
		}

		l = append(l, m)
	}
	return l
}

func SuffixFromARN(arn *string) string {
	if arn == nil {
		return ""
	}

	if arnComponents := regexp.MustCompile(`arn:.*:loadbalancer/(.*)`).FindAllStringSubmatch(*arn, -1); len(arnComponents) == 1 {
		if len(arnComponents[0]) == 2 {
			return arnComponents[0][1]
		}
	}

	return ""
}

// flattenResource takes a *elbv2.LoadBalancer and populates all respective resource fields.
func flattenResource(ctx context.Context, d *schema.ResourceData, meta interface{}, lb *elbv2.LoadBalancer) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	d.Set("arn", lb.LoadBalancerArn)
	d.Set("arn_suffix", SuffixFromARN(lb.LoadBalancerArn))
	d.Set("name", lb.LoadBalancerName)
	d.Set("internal", lb.Scheme != nil && aws.StringValue(lb.Scheme) == "internal")
	d.Set("security_groups", flex.FlattenStringList(lb.SecurityGroups))
	d.Set("vpc_id", lb.VpcId)
	d.Set("zone_id", lb.CanonicalHostedZoneId)
	d.Set("dns_name", lb.DNSName)
	d.Set("ip_address_type", lb.IpAddressType)
	d.Set("load_balancer_type", lb.Type)
	d.Set("customer_owned_ipv4_pool", lb.CustomerOwnedIpv4Pool)

	if err := d.Set("subnets", flattenSubnetsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return fmt.Errorf("setting subnets: %w", err)
	}

	if err := d.Set("subnet_mapping", flattenSubnetMappingsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return fmt.Errorf("setting subnet_mapping: %w", err)
	}

	attributesResp, err := conn.DescribeLoadBalancerAttributesWithContext(ctx, &elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("retrieving LB Attributes: %w", err)
	}

	accessLogMap := map[string]interface{}{
		"bucket":  "",
		"enabled": false,
		"prefix":  "",
	}

	for _, attr := range attributesResp.Attributes {
		switch aws.StringValue(attr.Key) {
		case "access_logs.s3.enabled":
			accessLogMap["enabled"] = aws.StringValue(attr.Value) == "true"
		case "access_logs.s3.bucket":
			accessLogMap["bucket"] = aws.StringValue(attr.Value)
		case "access_logs.s3.prefix":
			accessLogMap["prefix"] = aws.StringValue(attr.Value)
		case "idle_timeout.timeout_seconds":
			timeout, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("parsing ALB timeout: %w", err)
			}
			log.Printf("[DEBUG] Setting ALB Timeout Seconds: %d", timeout)
			d.Set("idle_timeout", timeout)
		case "routing.http.drop_invalid_header_fields.enabled":
			dropInvalidHeaderFieldsEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting LB Invalid Header Fields Enabled: %t", dropInvalidHeaderFieldsEnabled)
			d.Set("drop_invalid_header_fields", dropInvalidHeaderFieldsEnabled)
		case "routing.http.preserve_host_header.enabled":
			preserveHostHeaderEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting LB Preserve Host Header Enabled: %t", preserveHostHeaderEnabled)
			d.Set("preserve_host_header", preserveHostHeaderEnabled)
		case "deletion_protection.enabled":
			protectionEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting LB Deletion Protection Enabled: %t", protectionEnabled)
			d.Set("enable_deletion_protection", protectionEnabled)
		case "routing.http2.enabled":
			http2Enabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting ALB HTTP/2 Enabled: %t", http2Enabled)
			d.Set("enable_http2", http2Enabled)
		case "waf.fail_open.enabled":
			wafFailOpenEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting ALB WAF fail open Enabled: %t", wafFailOpenEnabled)
			d.Set("enable_waf_fail_open", wafFailOpenEnabled)
		case "load_balancing.cross_zone.enabled":
			crossZoneLbEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting NLB Cross Zone Load Balancing Enabled: %t", crossZoneLbEnabled)
			d.Set("enable_cross_zone_load_balancing", crossZoneLbEnabled)
		case "routing.http.desync_mitigation_mode":
			desyncMitigationMode := aws.StringValue(attr.Value)
			log.Printf("[DEBUG] Setting ALB Desync Mitigation Mode: %s", desyncMitigationMode)
			d.Set("desync_mitigation_mode", desyncMitigationMode)
		}
	}

	if err := d.Set("access_logs", []interface{}{accessLogMap}); err != nil {
		return fmt.Errorf("setting access_logs: %w", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] Unable to list tags for ELBv2 Load Balancer %s: %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing tags for (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

// Load balancers of type 'network' cannot have their subnets updated at
// this time. If the type is 'network' and subnets have changed, mark the
// diff as a ForceNew operation
func customizeDiffNLBSubnets(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// The current criteria for determining if the operation should be ForceNew:
	// - lb of type "network"
	// - existing resource (id is not "")
	// - there are actual changes to be made in the subnets
	//
	// Any other combination should be treated as normal. At this time, subnet
	// handling is the only known difference between Network Load Balancers and
	// Application Load Balancers, so the logic below is simple individual checks.
	// If other differences arise we'll want to refactor to check other
	// conditions in combinations, but for now all we handle is subnets
	if lbType := diff.Get("load_balancer_type").(string); lbType != elbv2.LoadBalancerTypeEnumNetwork {
		return nil
	}

	if diff.Id() == "" {
		return nil
	}

	o, n := diff.GetChange("subnets")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	remove := os.Difference(ns).List()
	add := ns.Difference(os).List()
	if len(remove) > 0 || len(add) > 0 {
		if err := diff.SetNew("subnets", n); err != nil {
			return err
		}

		if err := diff.ForceNew("subnets"); err != nil {
			return err
		}
	}
	return nil
}
