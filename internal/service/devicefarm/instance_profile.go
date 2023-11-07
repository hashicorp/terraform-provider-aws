// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_instance_profile", name="Instance Profile")
// @Tags(identifierAttribute="arn")
func ResourceInstanceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceProfileCreate,
		ReadWithoutTimeout:   resourceInstanceProfileRead,
		UpdateWithoutTimeout: resourceInstanceProfileUpdate,
		DeleteWithoutTimeout: resourceInstanceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"exclude_app_packages_from_cleanup": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
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
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	name := d.Get("name").(string)
	input := &devicefarm.CreateInstanceProfileInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("exclude_app_packages_from_cleanup"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludeAppPackagesFromCleanup = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("package_cleanup"); ok {
		input.PackageCleanup = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("reboot_after_use"); ok {
		input.RebootAfterUse = aws.Bool(v.(bool))
	}

	output, err := conn.CreateInstanceProfileWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Instance Profile (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.InstanceProfile.Arn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DeviceFarm Instance Profile (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
}

func resourceInstanceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	instaceProf, err := FindInstanceProfileByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Instance Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Instance Profile (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(instaceProf.Arn)
	d.Set("arn", arn)
	d.Set("name", instaceProf.Name)
	d.Set("description", instaceProf.Description)
	d.Set("exclude_app_packages_from_cleanup", flex.FlattenStringSet(instaceProf.ExcludeAppPackagesFromCleanup))
	d.Set("package_cleanup", instaceProf.PackageCleanup)
	d.Set("reboot_after_use", instaceProf.RebootAfterUse)

	return diags
}

func resourceInstanceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateInstanceProfileInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("exclude_app_packages_from_cleanup") {
			input.ExcludeAppPackagesFromCleanup = flex.ExpandStringSet(d.Get("exclude_app_packages_from_cleanup").(*schema.Set))
		}

		if d.HasChange("package_cleanup") {
			input.PackageCleanup = aws.Bool(d.Get("package_cleanup").(bool))
		}

		if d.HasChange("reboot_after_use") {
			input.RebootAfterUse = aws.Bool(d.Get("reboot_after_use").(bool))
		}

		_, err := conn.UpdateInstanceProfileWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Instance Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceInstanceProfileRead(ctx, d, meta)...)
}

func resourceInstanceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	log.Printf("[DEBUG] Deleting DeviceFarm Instance Profile: %s", d.Id())
	_, err := conn.DeleteInstanceProfileWithContext(ctx, &devicefarm.DeleteInstanceProfileInput{
		Arn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Instance Profile (%s): %s", d.Id(), err)
	}

	return diags
}
