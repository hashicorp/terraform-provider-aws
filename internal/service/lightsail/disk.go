// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_disk", name="Disk")
// @Tags(identifierAttribute="id", resourceType="Disk")
func ResourceDisk() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDiskCreate,
		ReadWithoutTimeout:   resourceDiskRead,
		UpdateWithoutTimeout: resourceDiskUpdate,
		DeleteWithoutTimeout: resourceDiskDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+[^_.-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"size_in_gb": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	id := d.Get(names.AttrName).(string)
	in := lightsail.CreateDiskInput{
		AvailabilityZone: aws.String(d.Get(names.AttrAvailabilityZone).(string)),
		SizeInGb:         aws.Int32(int32(d.Get("size_in_gb").(int))),
		DiskName:         aws.String(id),
		Tags:             getTagsIn(ctx),
	}

	out, err := conn.CreateDisk(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeCreateDisk), ResDisk, id, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeCreateDisk, ResDisk, id)

	if diag != nil {
		return diag
	}

	d.SetId(id)

	return append(diags, resourceDiskRead(ctx, d, meta)...)
}

func resourceDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindDiskById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResDisk, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResDisk, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrAvailabilityZone, out.Location.AvailabilityZone)
	d.Set(names.AttrCreatedAt, out.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set("size_in_gb", out.SizeInGb)
	d.Set("support_code", out.SupportCode)

	setTagsOut(ctx, out.Tags)

	return diags
}

func resourceDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceDiskRead(ctx, d, meta)
}

func resourceDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := conn.DeleteDisk(ctx, &lightsail.DeleteDiskInput{
		DiskName: aws.String(d.Id()),
	})

	if IsANotFoundError(err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeDeleteDisk), ResDisk, d.Get(names.AttrName).(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDeleteDisk, ResDisk, d.Id())

	if diag != nil {
		return diag
	}

	return diags
}

func FindDiskById(ctx context.Context, conn *lightsail.Client, id string) (*types.Disk, error) {
	in := &lightsail.GetDiskInput{
		DiskName: aws.String(id),
	}

	out, err := conn.GetDisk(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Disk == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Disk, nil
}
