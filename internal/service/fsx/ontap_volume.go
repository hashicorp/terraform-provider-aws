// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_ontap_volume", name="ONTAP Volume")
// @Tags(identifierAttribute="arn")
func ResourceOntapVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOntapVolumeCreate,
		ReadWithoutTimeout:   resourceOntapVolumeRead,
		UpdateWithoutTimeout: resourceOntapVolumeUpdate,
		DeleteWithoutTimeout: resourceOntapVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("skip_final_backup", false)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flexcache_endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"junction_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 203),
			},
			"ontap_volume_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.InputOntapVolumeType_Values(), false),
			},
			"security_style": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(fsx.StorageVirtualMachineRootVolumeSecurityStyle_Values(), false),
			},
			"size_in_megabytes": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},
			"skip_final_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"storage_efficiency_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"storage_virtual_machine_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(21, 21),
			},
			"tiering_policy": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cooling_period": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 183),
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(fsx.TieringPolicyName_Values(), false),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.VolumeTypeOntap,
				ValidateFunc: validation.StringInSlice(fsx.VolumeType_Values(), false),
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOntapVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	name := d.Get("name").(string)
	input := &fsx.CreateVolumeInput{
		Name: aws.String(name),
		OntapConfiguration: &fsx.CreateOntapVolumeConfiguration{
			SizeInMegabytes:         aws.Int64(int64(d.Get("size_in_megabytes").(int))),
			StorageVirtualMachineId: aws.String(d.Get("storage_virtual_machine_id").(string)),
		},
		Tags:       getTagsIn(ctx),
		VolumeType: aws.String(d.Get("volume_type").(string)),
	}

	if v, ok := d.GetOk("junction_path"); ok {
		input.OntapConfiguration.JunctionPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ontap_volume_type"); ok {
		input.OntapConfiguration.OntapVolumeType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_style"); ok {
		input.OntapConfiguration.SecurityStyle = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("storage_efficiency_enabled"); ok {
		input.OntapConfiguration.StorageEfficiencyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("tiering_policy"); ok {
		input.OntapConfiguration.TieringPolicy = expandOntapVolumeTieringPolicy(v.([]interface{}))
	}

	result, err := conn.CreateVolumeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx ONTAP Volume (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(result.Volume.VolumeId))

	if _, err := waitVolumeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx ONTAP Volume (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOntapVolumeRead(ctx, d, meta)...)
}

func resourceOntapVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	volume, err := FindVolumeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx ONTAP Volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx ONTAP Volume (%s): %s", d.Id(), err)
	}

	ontapConfig := volume.OntapConfiguration
	if ontapConfig == nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx ONTAP Volume (%s): empty ONTAP configuration", d.Id())
	}

	d.Set("arn", volume.ResourceARN)
	d.Set("name", volume.Name)
	d.Set("file_system_id", volume.FileSystemId)
	d.Set("junction_path", ontapConfig.JunctionPath)
	d.Set("ontap_volume_type", ontapConfig.OntapVolumeType)
	d.Set("security_style", ontapConfig.SecurityStyle)
	d.Set("size_in_megabytes", ontapConfig.SizeInMegabytes)
	d.Set("storage_efficiency_enabled", ontapConfig.StorageEfficiencyEnabled)
	d.Set("storage_virtual_machine_id", ontapConfig.StorageVirtualMachineId)
	if err := d.Set("tiering_policy", flattenOntapVolumeTieringPolicy(ontapConfig.TieringPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tiering_policy: %s", err)
	}
	d.Set("uuid", ontapConfig.UUID)
	d.Set("volume_type", volume.VolumeType)

	return diags
}

func resourceOntapVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateVolumeInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			OntapConfiguration: &fsx.UpdateOntapVolumeConfiguration{},
			VolumeId:           aws.String(d.Id()),
		}

		if d.HasChange("junction_path") {
			input.OntapConfiguration.JunctionPath = aws.String(d.Get("junction_path").(string))
		}

		if d.HasChange("security_style") {
			input.OntapConfiguration.SecurityStyle = aws.String(d.Get("security_style").(string))
		}

		if d.HasChange("size_in_megabytes") {
			input.OntapConfiguration.SizeInMegabytes = aws.Int64(int64(d.Get("size_in_megabytes").(int)))
		}

		if d.HasChange("storage_efficiency_enabled") {
			input.OntapConfiguration.StorageEfficiencyEnabled = aws.Bool(d.Get("storage_efficiency_enabled").(bool))
		}

		if d.HasChange("tiering_policy") {
			input.OntapConfiguration.TieringPolicy = expandOntapVolumeTieringPolicy(d.Get("tiering_policy").([]interface{}))
		}

		_, err := conn.UpdateVolumeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx ONTAP Volume (%s): %s", d.Id(), err)
		}

		if _, err := waitVolumeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx ONTAP Volume (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOntapVolumeRead(ctx, d, meta)...)
}

func resourceOntapVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[DEBUG] Deleting FSx ONTAP Volume: %s", d.Id())
	_, err := conn.DeleteVolumeWithContext(ctx, &fsx.DeleteVolumeInput{
		OntapConfiguration: &fsx.DeleteVolumeOntapConfiguration{
			SkipFinalBackup: aws.Bool(d.Get("skip_final_backup").(bool)),
		},
		VolumeId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx ONTAP Volume (%s): %s", d.Id(), err)
	}

	if _, err := waitVolumeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx ONTAP Volume (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandOntapVolumeTieringPolicy(cfg []interface{}) *fsx.TieringPolicy {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.TieringPolicy{}

	//Cooling period only accepts a minimum of 2 but int will return 0 not nil if unset
	//Therefore we only set it if it is 2 or more
	if v, ok := conf["cooling_period"].(int); ok && v >= 2 {
		out.CoolingPeriod = aws.Int64(int64(v))
	}

	if v, ok := conf["name"].(string); ok {
		out.Name = aws.String(v)
	}

	return &out
}

func flattenOntapVolumeTieringPolicy(rs *fsx.TieringPolicy) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	minCoolingPeriod := 2

	m := make(map[string]interface{})
	if aws.Int64Value(rs.CoolingPeriod) >= int64(minCoolingPeriod) {
		m["cooling_period"] = aws.Int64Value(rs.CoolingPeriod)
	}

	if rs.Name != nil {
		m["name"] = aws.StringValue(rs.Name)
	}

	return []interface{}{m}
}
