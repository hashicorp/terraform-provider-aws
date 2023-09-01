// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_disk_attachment")
func ResourceDiskAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDiskAttachmentCreate,
		ReadWithoutTimeout:   resourceDiskAttachmentRead,
		DeleteWithoutTimeout: resourceDiskAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disk_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"disk_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDiskAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	in := lightsail.AttachDiskInput{
		DiskName:     aws.String(d.Get("disk_name").(string)),
		DiskPath:     aws.String(d.Get("disk_path").(string)),
		InstanceName: aws.String(d.Get("instance_name").(string)),
	}

	out, err := conn.AttachDisk(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, string(types.OperationTypeAttachDisk), ResDiskAttachment, d.Get("disk_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeAttachDisk, ResDiskAttachment, d.Get("disk_name").(string))

	if diag != nil {
		return diag
	}

	// Generate an ID
	vars := []string{
		d.Get("disk_name").(string),
		d.Get("instance_name").(string),
	}

	d.SetId(strings.Join(vars, ","))

	return resourceDiskAttachmentRead(ctx, d, meta)
}

func resourceDiskAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindDiskAttachmentById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResDiskAttachment, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResDiskAttachment, d.Id(), err)
	}

	d.Set("disk_name", out.Name)
	d.Set("disk_path", out.Path)
	d.Set("instance_name", out.AttachedTo)

	return nil
}

func resourceDiskAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	id_parts := strings.SplitN(d.Id(), ",", -1)
	dName := id_parts[0]
	iName := id_parts[1]

	// A Disk can only be detached from a stopped instance
	iStateOut, err := waitInstanceState(ctx, conn, &iName)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, iName, errors.New("Error waiting for Instance to enter running or stopped state"))
	}

	if aws.ToString(iStateOut.State.Name) == "running" {
		stopOut, err := conn.StopInstance(ctx, &lightsail.StopInstanceInput{
			InstanceName: aws.String(iName),
		})

		if err != nil {
			return create.DiagError(names.Lightsail, string(types.OperationTypeStopInstance), ResInstance, iName, err)
		}

		diag := expandOperations(ctx, conn, stopOut.Operations, types.OperationTypeStopInstance, ResInstance, iName)

		if diag != nil {
			return diag
		}
	}

	out, err := conn.DetachDisk(ctx, &lightsail.DetachDiskInput{
		DiskName: aws.String(dName),
	})

	if err != nil {
		return create.DiagError(names.Lightsail, string(types.OperationTypeDetachDisk), ResDiskAttachment, d.Get("disk_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDetachDisk, ResDiskAttachment, d.Get("disk_name").(string))

	if diag != nil {
		return diag
	}

	iStateOut, err = waitInstanceState(ctx, conn, &iName)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, iName, errors.New("Error waiting for Instance to enter running or stopped state"))
	}

	if aws.ToString(iStateOut.State.Name) != "running" {
		startOut, err := conn.StartInstance(ctx, &lightsail.StartInstanceInput{
			InstanceName: aws.String(iName),
		})

		if err != nil {
			return create.DiagError(names.Lightsail, string(types.OperationTypeStartInstance), ResInstance, iName, err)
		}

		diag := expandOperations(ctx, conn, startOut.Operations, types.OperationTypeStartInstance, ResInstance, iName)

		if diag != nil {
			return diag
		}
	}

	return nil
}

func FindDiskAttachmentById(ctx context.Context, conn *lightsail.Client, id string) (*types.Disk, error) {
	id_parts := strings.SplitN(id, ",", -1)

	if len(id_parts) != 2 {
		return nil, errors.New("invalid Disk Attachment id")
	}

	dName := id_parts[0]
	iName := id_parts[1]

	in := &lightsail.GetDiskInput{
		DiskName: aws.String(dName),
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

	disk := out.Disk

	if disk == nil || !aws.ToBool(disk.IsAttached) || aws.ToString(disk.Name) != dName || aws.ToString(disk.AttachedTo) != iName {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Disk, nil
}
