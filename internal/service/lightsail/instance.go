package lightsail

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lightsail.AddOnType_Values(), false),
						},
						"snapshot_time": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`), "must be in HH:00 format, and in Coordinated Universal Time (UTC)."),
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"availability_zone": {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
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
			"ip_address_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "dualstack",
			},
			"ipv6_address": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use `ipv6_addresses` attribute instead",
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
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	iName := d.Get("name").(string)

	in := lightsail.CreateInstancesInput{
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		BlueprintId:      aws.String(d.Get("blueprint_id").(string)),
		BundleId:         aws.String(d.Get("bundle_id").(string)),
		InstanceNames:    aws.StringSlice([]string{iName}),
	}

	if v, ok := d.GetOk("key_pair_name"); ok {
		in.KeyPairName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_data"); ok {
		in.UserData = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		in.IpAddressType = aws.String(v.(string))
	}

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateInstancesWithContext(ctx, &in)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateInstance, ResInstance, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateInstance, ResInstance, d.Get("name").(string), errors.New("No operations found for request"))
	}

	op := out.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateInstance, ResInstance, d.Get("name").(string), errors.New("Error waiting for request operation"))
	}

	// Cannot enable add ons with creation request
	if expandAddOnEnabled(d.Get("add_on").([]interface{})) {
		in := lightsail.EnableAddOnInput{
			ResourceName: aws.String(d.Get("name").(string)),
			AddOnRequest: expandAddOnRequest(d.Get("add_on").([]interface{})),
		}

		out, err := conn.EnableAddOnWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeEnableAddOn, ResInstance, d.Get("name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeEnableAddOn, ResInstance, d.Get("name").(string), errors.New("No operations found for request"))
		}

		op := out.Operations[0]

		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeEnableAddOn, ResInstance, d.Get("name").(string), errors.New("Error waiting for request operation"))
		}
	}

	return resourceInstanceRead(ctx, d, meta)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindInstanceById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResInstance, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, d.Id(), err)
	}

	d.Set("add_on", flattenAddOns(out.AddOns))
	d.Set("availability_zone", out.Location.AvailabilityZone)
	d.Set("blueprint_id", out.BlueprintId)
	d.Set("bundle_id", out.BundleId)
	d.Set("key_pair_name", out.SshKeyName)
	d.Set("name", out.Name)

	// additional attributes
	d.Set("arn", out.Arn)
	d.Set("username", out.Username)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("cpu_count", out.Hardware.CpuCount)
	d.Set("ram_size", out.Hardware.RamSizeInGb)

	// Deprecated: AWS Go SDK v1.36.25 removed Ipv6Address field
	if len(out.Ipv6Addresses) > 0 {
		d.Set("ipv6_address", out.Ipv6Addresses[0])
	}

	d.Set("ipv6_addresses", aws.StringValueSlice(out.Ipv6Addresses))
	d.Set("ip_address_type", out.IpAddressType)
	d.Set("is_static_ip", out.IsStaticIp)
	d.Set("private_ip_address", out.PrivateIpAddress)
	d.Set("public_ip_address", out.PublicIpAddress)

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, d.Id(), err)
	}

	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	out, err := conn.DeleteInstanceWithContext(ctx, &lightsail.DeleteInstanceInput{
		InstanceName:      aws.String(d.Id()),
		ForceDeleteAddOns: aws.Bool(true),
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionDeleting, ResInstance, d.Id(), err)
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteInstance, ResInstance, d.Id(), err)
	}

	return nil
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChange("ip_address_type") {
		out, err := conn.SetIpAddressTypeWithContext(ctx, &lightsail.SetIpAddressTypeInput{
			ResourceName:  aws.String(d.Id()),
			ResourceType:  aws.String("Instance"),
			IpAddressType: aws.String(d.Get("ip_address_type").(string)),
		})

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateInstance, ResInstance, d.Get("name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateInstance, ResInstance, d.Get("name").(string), errors.New("No operations found for request"))
		}

		op := out.Operations[0]

		err = waitOperation(ctx, conn, op.Id)
		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateInstance, ResInstance, d.Get("name").(string), errors.New("Error waiting for request operation"))
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResInstance, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResInstance, d.Id(), err)
		}
	}

	if d.HasChange("add_on") {
		o, n := d.GetChange("add_on")

		if err := updateAddOnWithContext(ctx, conn, d.Id(), o, n); err != nil {
			return err
		}
	}

	return resourceInstanceRead(ctx, d, meta)
}

func expandAddOnRequest(addOnListRaw []interface{}) *lightsail.AddOnRequest {
	if len(addOnListRaw) == 0 {
		return &lightsail.AddOnRequest{}
	}

	addOnRequest := &lightsail.AddOnRequest{}

	for _, addOnRaw := range addOnListRaw {
		addOnMap := addOnRaw.(map[string]interface{})
		addOnRequest.AddOnType = aws.String(addOnMap["type"].(string))
		addOnRequest.AutoSnapshotAddOnRequest = &lightsail.AutoSnapshotAddOnRequest{
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
		enabled = addOnMap["status"].(string) == "Enabled"
	}

	return enabled
}

func flattenAddOns(addOns []*lightsail.AddOn) []interface{} {
	var rawAddOns []interface{}

	for _, addOn := range addOns {
		rawAddOn := map[string]interface{}{
			"type":          aws.StringValue(addOn.Name),
			"snapshot_time": aws.StringValue(addOn.SnapshotTimeOfDay),
			"status":        aws.StringValue(addOn.Status),
		}
		rawAddOns = append(rawAddOns, rawAddOn)
	}

	return rawAddOns
}

func updateAddOnWithContext(ctx context.Context, conn *lightsail.Lightsail, name string, oldAddOnsRaw interface{}, newAddOnsRaw interface{}) diag.Diagnostics {
	oldAddOns := expandAddOnRequest(oldAddOnsRaw.([]interface{}))
	newAddOns := expandAddOnRequest(newAddOnsRaw.([]interface{}))
	oldAddOnStatus := expandAddOnEnabled(oldAddOnsRaw.([]interface{}))
	newAddonStatus := expandAddOnEnabled(newAddOnsRaw.([]interface{}))

	if (oldAddOnStatus && newAddonStatus) || !newAddonStatus {
		in := lightsail.DisableAddOnInput{
			ResourceName: aws.String(name),
			AddOnType:    oldAddOns.AddOnType,
		}

		out, err := conn.DisableAddOnWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeDisableAddOn, ResInstance, name, err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeDisableAddOn, ResInstance, name, errors.New("No operations found for request"))
		}

		op := out.Operations[0]

		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeDisableAddOn, ResInstance, name, errors.New("Error waiting for request operation"))
		}
	}

	if newAddonStatus {
		in := lightsail.EnableAddOnInput{
			ResourceName: aws.String(name),
			AddOnRequest: newAddOns,
		}

		out, err := conn.EnableAddOnWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeEnableAddOn, ResInstance, name, err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeEnableAddOn, ResInstance, name, errors.New("No operations found for request"))
		}

		op := out.Operations[0]

		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeEnableAddOn, ResInstance, name, errors.New("Error waiting for request operation"))
		}
	}

	return nil
}
