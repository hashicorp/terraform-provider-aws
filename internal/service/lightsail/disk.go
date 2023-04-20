package lightsail

import (
	"context"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
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

// @SDKResource("aws_lightsail_disk", name="Disk")
// @Tags(identifierAttribute="id")
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).LightsailConn()

	id := d.Get("name").(string)
	in := lightsail.CreateDiskInput{
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		SizeInGb:         aws.Int64(int64(d.Get("size_in_gb").(int))),
		DiskName:         aws.String(id),
		Tags:             GetTagsIn(ctx),
	}

	out, err := conn.CreateDiskWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateDisk, ResDisk, id, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeCreateDisk, ResDisk, id)

	if diag != nil {
		return diag
	}

	d.SetId(id)

	return resourceDiskRead(ctx, d, meta)
}

func resourceDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindDiskById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResDisk, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResDisk, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("availability_zone", out.Location.AvailabilityZone)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("name", out.Name)
	d.Set("size_in_gb", out.SizeInGb)
	d.Set("support_code", out.SupportCode)

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceDiskRead(ctx, d, meta)
}

func resourceDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := conn.DeleteDiskWithContext(ctx, &lightsail.DeleteDiskInput{
		DiskName: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteDisk, ResDisk, d.Get("name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeDeleteDisk, ResDisk, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}
