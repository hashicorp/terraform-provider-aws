package lightsail

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.AttachDiskInput{
		DiskName:     aws.String(d.Get("disk_name").(string)),
		DiskPath:     aws.String(d.Get("disk_path").(string)),
		InstanceName: aws.String(d.Get("instance_name").(string)),
	}

	out, err := conn.AttachDiskWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeAttachDisk, ResDiskAttachment, d.Get("disk_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeAttachDisk, ResDiskAttachment, d.Get("disk_name").(string))

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
	conn := meta.(*conns.AWSClient).LightsailConn()

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
	conn := meta.(*conns.AWSClient).LightsailConn()

	id_parts := strings.SplitN(d.Id(), ",", -1)
	dName := id_parts[0]
	iName := id_parts[1]

	// A Disk can only be detached from a stopped instance
	iStateOut, err := waitInstanceStateWithContext(ctx, conn, &iName)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, iName, errors.New("Error waiting for Instance to enter running or stopped state"))
	}

	if aws.StringValue(iStateOut.State.Name) == "running" {
		stopOut, err := conn.StopInstanceWithContext(ctx, &lightsail.StopInstanceInput{
			InstanceName: aws.String(iName),
		})

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeStopInstance, ResInstance, iName, err)
		}

		diag := expandOperations(ctx, conn, stopOut.Operations, lightsail.OperationTypeStopInstance, ResInstance, iName)

		if diag != nil {
			return diag
		}
	}

	out, err := conn.DetachDiskWithContext(ctx, &lightsail.DetachDiskInput{
		DiskName: aws.String(dName),
	})

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDetachDisk, ResDiskAttachment, d.Get("disk_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeDetachDisk, ResDiskAttachment, d.Get("disk_name").(string))

	if diag != nil {
		return diag
	}

	iStateOut, err = waitInstanceStateWithContext(ctx, conn, &iName)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResInstance, iName, errors.New("Error waiting for Instance to enter running or stopped state"))
	}

	if aws.StringValue(iStateOut.State.Name) != "running" {
		startOut, err := conn.StartInstanceWithContext(ctx, &lightsail.StartInstanceInput{
			InstanceName: aws.String(iName),
		})

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeStartInstance, ResInstance, iName, err)
		}

		diag := expandOperations(ctx, conn, startOut.Operations, lightsail.OperationTypeStartInstance, ResInstance, iName)

		if diag != nil {
			return diag
		}
	}

	return nil
}
