// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_instance", name="Instance")
// @Tags(identifierAttribute="id", resourceType="Instance")
func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstanceUpdate,
		DeleteWithoutTimeout: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"add_on": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(flattenAddOnTypeValues(types.AddOnType("").Values()), false),
						},
						"snapshot_time": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`), "must be in HH:00 format, and in Coordinated Universal Time (UTC)."),
						},
						names.AttrStatus: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]`), "must begin with an alphanumeric character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+[^_.-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"blueprint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// Optional attributes
			"key_pair_name": {
				// Not compatible with aws_key_pair (yet)
				// We'll need a new aws_lightsail_key_pair resource
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "LightsailDefaultKeyPair" && new == "" {
						return true
					}
					return false
				},
			},

			// cannot be retrieved from the API
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// additional info returned from the API
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "dualstack",
			},
			"ipv6_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"is_static_ip": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUsername: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange(names.AttrAvailabilityZone, func(ctx context.Context, old, new, meta any) error {
				// The availability_zone must be in the same region as the provider region
				if !strings.HasPrefix(new.(string), meta.(*conns.AWSClient).Region) {
					return fmt.Errorf("availability_zone must be within the same region as provider region: %s", meta.(*conns.AWSClient).Region)
				}
				return nil
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	iName := d.Get(names.AttrName).(string)

	in := lightsail.CreateInstancesInput{
		AvailabilityZone: aws.String(d.Get(names.AttrAvailabilityZone).(string)),
		BlueprintId:      aws.String(d.Get("blueprint_id").(string)),
		BundleId:         aws.String(d.Get("bundle_id").(string)),
		InstanceNames:    []string{iName},
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("key_pair_name"); ok {
		in.KeyPairName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_data"); ok {
		in.UserData = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIPAddressType); ok {
		in.IpAddressType = types.IpAddressType(v.(string))
	}

	out, err := conn.CreateInstances(ctx, &in)
	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeCreateInstance), ResInstance, iName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeCreateInstance, ResInstance, iName)

	if diag != nil {
		return diag
	}

	d.SetId(iName)

	// Cannot enable add ons with creation request
	if expandAddOnEnabled(d.Get("add_on").([]interface{})) {
		in := lightsail.EnableAddOnInput{
			ResourceName: aws.String(iName),
			AddOnRequest: expandAddOnRequest(d.Get("add_on").([]interface{})),
		}

		out, err := conn.EnableAddOn(ctx, &in)

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeEnableAddOn), ResInstance, iName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeEnableAddOn, ResInstance, iName)

		if diag != nil {
			return diag
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindInstanceById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResInstance, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResInstance, d.Id(), err)
	}

	d.Set("add_on", flattenAddOns(out.AddOns))
	d.Set(names.AttrAvailabilityZone, out.Location.AvailabilityZone)
	d.Set("blueprint_id", out.BlueprintId)
	d.Set("bundle_id", out.BundleId)
	d.Set("key_pair_name", out.SshKeyName)
	d.Set(names.AttrName, out.Name)

	// additional attributes
	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrUsername, out.Username)
	d.Set(names.AttrCreatedAt, out.CreatedAt.Format(time.RFC3339))
	d.Set("cpu_count", out.Hardware.CpuCount)
	d.Set("ram_size", out.Hardware.RamSizeInGb)

	d.Set("ipv6_addresses", out.Ipv6Addresses)
	d.Set(names.AttrIPAddressType, out.IpAddressType)
	d.Set("is_static_ip", out.IsStaticIp)
	d.Set("private_ip_address", out.PrivateIpAddress)
	d.Set("public_ip_address", out.PublicIpAddress)

	setTagsOut(ctx, out.Tags)

	return diags
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	out, err := conn.DeleteInstance(ctx, &lightsail.DeleteInstanceInput{
		InstanceName:      aws.String(d.Id()),
		ForceDeleteAddOns: aws.Bool(true),
	})

	if err != nil && errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionDeleting, ResInstance, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDeleteInstance, ResInstance, d.Id())

	if diag != nil {
		return diag
	}

	return diags
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	if d.HasChange(names.AttrIPAddressType) {
		out, err := conn.SetIpAddressType(ctx, &lightsail.SetIpAddressTypeInput{
			ResourceName:  aws.String(d.Id()),
			ResourceType:  types.ResourceTypeInstance,
			IpAddressType: types.IpAddressType(d.Get(names.AttrIPAddressType).(string)),
		})

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeSetIpAddressType), ResInstance, d.Id(), err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeSetIpAddressType, ResInstance, d.Id())

		if diag != nil {
			return diag
		}
	}

	if d.HasChange("add_on") {
		o, n := d.GetChange("add_on")

		if err := updateAddOn(ctx, conn, d.Id(), o, n); err != nil {
			return err
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func expandAddOnRequest(addOnListRaw []interface{}) *types.AddOnRequest {
	if len(addOnListRaw) == 0 {
		return &types.AddOnRequest{}
	}

	addOnRequest := &types.AddOnRequest{}

	for _, addOnRaw := range addOnListRaw {
		addOnMap := addOnRaw.(map[string]interface{})
		addOnRequest.AddOnType = types.AddOnType(addOnMap[names.AttrType].(string))
		addOnRequest.AutoSnapshotAddOnRequest = &types.AutoSnapshotAddOnRequest{
			SnapshotTimeOfDay: aws.String(addOnMap["snapshot_time"].(string)),
		}
	}

	return addOnRequest
}

func expandAddOnEnabled(addOnListRaw []interface{}) bool {
	if len(addOnListRaw) == 0 {
		return false
	}

	var enabled bool
	for _, addOnRaw := range addOnListRaw {
		addOnMap := addOnRaw.(map[string]interface{})
		enabled = addOnMap[names.AttrStatus].(string) == "Enabled"
	}

	return enabled
}

func flattenAddOns(addOns []types.AddOn) []interface{} {
	var rawAddOns []interface{}

	for _, addOn := range addOns {
		rawAddOn := map[string]interface{}{
			names.AttrType:   aws.ToString(addOn.Name),
			"snapshot_time":  aws.ToString(addOn.SnapshotTimeOfDay),
			names.AttrStatus: aws.ToString(addOn.Status),
		}
		rawAddOns = append(rawAddOns, rawAddOn)
	}

	return rawAddOns
}

func updateAddOn(ctx context.Context, conn *lightsail.Client, name string, oldAddOnsRaw interface{}, newAddOnsRaw interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	oldAddOns := expandAddOnRequest(oldAddOnsRaw.([]interface{}))
	newAddOns := expandAddOnRequest(newAddOnsRaw.([]interface{}))
	oldAddOnStatus := expandAddOnEnabled(oldAddOnsRaw.([]interface{}))
	newAddonStatus := expandAddOnEnabled(newAddOnsRaw.([]interface{}))

	if (oldAddOnStatus && newAddonStatus) || !newAddonStatus {
		in := lightsail.DisableAddOnInput{
			ResourceName: aws.String(name),
			AddOnType:    oldAddOns.AddOnType,
		}

		out, err := conn.DisableAddOn(ctx, &in)

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeDisableAddOn), ResInstance, name, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDisableAddOn, ResInstance, name)

		if diag != nil {
			return diag
		}
	}

	if newAddonStatus {
		in := lightsail.EnableAddOnInput{
			ResourceName: aws.String(name),
			AddOnRequest: newAddOns,
		}

		out, err := conn.EnableAddOn(ctx, &in)

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeEnableAddOn), ResInstance, name, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeEnableAddOn, ResInstance, name)

		if diag != nil {
			return diag
		}
	}

	return diags
}

func flattenAddOnTypeValues(t []types.AddOnType) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func FindInstanceById(ctx context.Context, conn *lightsail.Client, id string) (*types.Instance, error) {
	in := &lightsail.GetInstanceInput{InstanceName: aws.String(id)}
	out, err := conn.GetInstance(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Instance == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Instance, nil
}
