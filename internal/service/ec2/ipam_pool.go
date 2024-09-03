// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_pool", name="IPAM Pool")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceIPAMPool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMPoolCreate,
		ReadWithoutTimeout:   resourceIPAMPoolRead,
		UpdateWithoutTimeout: resourceIPAMPoolUpdate,
		DeleteWithoutTimeout: resourceIPAMPoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AddressFamily](),
			},
			"allocation_default_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"allocation_max_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"allocation_min_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"allocation_resource_tags": tftags.TagsSchema(),
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_import": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"aws_service": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IpamPoolAwsService](),
			},
			"cascade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_scope_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"None"}, false),
					verify.ValidRegionName,
				),
				Default: "None",
			},
			"pool_depth": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"public_ip_source": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IpamPoolPublicIpSource](),
				// default is byoip when AddressFamily = ipv6
				DiffSuppressFunc: func(k, o, n string, d *schema.ResourceData) bool {
					if o == "byoip" && n == "" {
						return true
					}
					return false
				},
			},
			"publicly_advertisable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"source_ipam_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIPAMPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	addressFamily := awstypes.AddressFamily(d.Get("address_family").(string))
	input := &ec2.CreateIpamPoolInput{
		AddressFamily:     addressFamily,
		ClientToken:       aws.String(id.UniqueId()),
		IpamScopeId:       aws.String(d.Get("ipam_scope_id").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeIpamPool),
	}

	if v, ok := d.GetOk("allocation_default_netmask_length"); ok {
		input.AllocationDefaultNetmaskLength = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("allocation_max_netmask_length"); ok {
		input.AllocationMaxNetmaskLength = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("allocation_min_netmask_length"); ok {
		input.AllocationMinNetmaskLength = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("allocation_resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.AllocationResourceTags = ipamResourceTags(tftags.New(ctx, v.(map[string]interface{})))
	}

	if v, ok := d.GetOk("auto_import"); ok {
		input.AutoImport = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("locale"); ok && v != "None" {
		input.Locale = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aws_service"); ok {
		input.AwsService = awstypes.IpamPoolAwsService(v.(string))
	}

	var publicIpSource awstypes.IpamPoolPublicIpSource
	if v, ok := d.GetOk("public_ip_source"); ok {
		publicIpSource = awstypes.IpamPoolPublicIpSource(v.(string))
		input.PublicIpSource = publicIpSource
	}

	// PubliclyAdvertisable must be set if if the AddressFamily is IPv6 and PublicIpSource is byoip.
	// The request can only contain PubliclyAdvertisable if the AddressFamily is IPv6 and PublicIpSource is byoip.
	if addressFamily == awstypes.AddressFamilyIpv6 && publicIpSource != awstypes.IpamPoolPublicIpSourceAmazon {
		input.PubliclyAdvertisable = aws.Bool(d.Get("publicly_advertisable").(bool))
	}

	if v, ok := d.GetOk("source_ipam_pool_id"); ok {
		input.SourceIpamPoolId = aws.String(v.(string))
	}

	output, err := conn.CreateIpamPool(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Pool: %s", err)
	}

	d.SetId(aws.ToString(output.IpamPool.IpamPoolId))

	if _, err := waitIPAMPoolCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMPoolRead(ctx, d, meta)...)
}

func resourceIPAMPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	pool, err := findIPAMPoolByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool (%s): %s", d.Id(), err)
	}

	d.Set("address_family", pool.AddressFamily)
	d.Set("allocation_resource_tags", keyValueTags(ctx, tagsFromIPAMAllocationTags(pool.AllocationResourceTags)).Map())
	d.Set(names.AttrARN, pool.IpamPoolArn)
	d.Set("auto_import", pool.AutoImport)
	d.Set("aws_service", pool.AwsService)
	d.Set(names.AttrDescription, pool.Description)
	scopeID := strings.Split(aws.ToString(pool.IpamScopeArn), "/")[1]
	d.Set("ipam_scope_id", scopeID)
	d.Set("ipam_scope_type", pool.IpamScopeType)
	d.Set("locale", pool.Locale)
	d.Set("pool_depth", pool.PoolDepth)
	d.Set("publicly_advertisable", pool.PubliclyAdvertisable)
	d.Set("public_ip_source", pool.PublicIpSource)
	d.Set("source_ipam_pool_id", pool.SourceIpamPoolId)
	d.Set(names.AttrState, pool.State)

	setTagsOut(ctx, pool.Tags)

	return diags
}

func resourceIPAMPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyIpamPoolInput{
			IpamPoolId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("allocation_default_netmask_length"); ok {
			input.AllocationDefaultNetmaskLength = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("allocation_max_netmask_length"); ok {
			input.AllocationMaxNetmaskLength = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("allocation_min_netmask_length"); ok {
			input.AllocationMinNetmaskLength = aws.Int32(int32(v.(int)))
		}

		if d.HasChange("allocation_resource_tags") {
			o, n := d.GetChange("allocation_resource_tags")
			oldTags := tftags.New(ctx, o)
			newTags := tftags.New(ctx, n)

			if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
				input.RemoveAllocationResourceTags = ipamResourceTags(removedTags.IgnoreAWS())
			}

			if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
				input.AddAllocationResourceTags = ipamResourceTags(updatedTags.IgnoreAWS())
			}
		}

		if v, ok := d.GetOk("auto_import"); ok {
			input.AutoImport = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.ModifyIpamPool(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM Pool (%s): %s", d.Id(), err)
		}

		if _, err := waitIPAMPoolUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPAMPoolRead(ctx, d, meta)...)
}

func resourceIPAMPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DeleteIpamPoolInput{
		IpamPoolId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("cascade"); ok {
		input.Cascade = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Deleting IPAM Pool: %s", d.Id())
	_, err := conn.DeleteIpamPool(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Pool (%s): %s", d.Id(), err)
	}

	if _, err = waitIPAMPoolDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func ipamResourceTags(tags tftags.KeyValueTags) []awstypes.RequestIpamResourceTag {
	result := make([]awstypes.RequestIpamResourceTag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := awstypes.RequestIpamResourceTag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

func tagsFromIPAMAllocationTags(rts []awstypes.IpamResourceTag) []awstypes.Tag {
	if len(rts) == 0 {
		return nil
	}

	tags := []awstypes.Tag{}
	for _, ts := range rts {
		tags = append(tags, awstypes.Tag{
			Key:   ts.Key,
			Value: ts.Value,
		})
	}

	return tags
}
