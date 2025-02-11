// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_instance_profile", name="Instance Profile")
// @Tags(identifierAttribute="arn")
func resourceInstanceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceProfileCreate,
		ReadWithoutTimeout:   resourceInstanceProfileRead,
		UpdateWithoutTimeout: resourceInstanceProfileUpdate,
		DeleteWithoutTimeout: resourceInstanceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"exclude_app_packages_from_cleanup": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"package_cleanup": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"reboot_after_use": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInstanceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &devicefarm.CreateInstanceProfileInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("exclude_app_packages_from_cleanup"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludeAppPackagesFromCleanup = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("package_cleanup"); ok {
		input.PackageCleanup = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("reboot_after_use"); ok {
		input.RebootAfterUse = aws.Bool(v.(bool))
	}

	output, err := conn.CreateInstanceProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Instance Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.InstanceProfile.Arn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DeviceFarm Instance Profile (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
}

func resourceInstanceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	instanceProf, err := findInstanceProfileByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Instance Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Instance Profile (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(instanceProf.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, instanceProf.Name)
	d.Set(names.AttrDescription, instanceProf.Description)
	d.Set("exclude_app_packages_from_cleanup", flex.FlattenStringValueSet(instanceProf.ExcludeAppPackagesFromCleanup))
	d.Set("package_cleanup", instanceProf.PackageCleanup)
	d.Set("reboot_after_use", instanceProf.RebootAfterUse)

	return diags
}

func resourceInstanceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &devicefarm.UpdateInstanceProfileInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("exclude_app_packages_from_cleanup") {
			input.ExcludeAppPackagesFromCleanup = flex.ExpandStringValueSet(d.Get("exclude_app_packages_from_cleanup").(*schema.Set))
		}

		if d.HasChange("package_cleanup") {
			input.PackageCleanup = aws.Bool(d.Get("package_cleanup").(bool))
		}

		if d.HasChange("reboot_after_use") {
			input.RebootAfterUse = aws.Bool(d.Get("reboot_after_use").(bool))
		}

		_, err := conn.UpdateInstanceProfile(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Instance Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
}

func resourceInstanceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	log.Printf("[DEBUG] Deleting DeviceFarm Instance Profile: %s", d.Id())
	_, err := conn.DeleteInstanceProfile(ctx, &devicefarm.DeleteInstanceProfileInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Instance Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findInstanceProfileByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.InstanceProfile, error) {
	input := &devicefarm.GetInstanceProfileInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetInstanceProfile(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InstanceProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InstanceProfile, nil
}
