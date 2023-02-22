package lightsail

import (
	"context"
	"errors"
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	in := lightsail.CreateDiskInput{
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		SizeInGb:         aws.Int64(int64(d.Get("size_in_gb").(int))),
		DiskName:         aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateDiskWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateDisk, ResDisk, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateDisk, ResDisk, d.Get("name").(string), errors.New("No operations found for Create Disk request"))
	}

	op := out.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateDisk, ResDisk, d.Get("name").(string), errors.New("Error waiting for Create Disk request operation"))
	}

	return resourceDiskRead(ctx, d, meta)
}

func resourceDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResDisk, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResDisk, d.Id(), err)
	}

	return nil
}

func resourceDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResDisk, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResDisk, d.Id(), err)
		}
	}

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

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteDisk, ResDisk, d.Get("name").(string), errors.New("No operations found for Delete Disk request"))
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteDisk, ResDisk, d.Get("name").(string), errors.New("Error waiting for Delete Disk request operation"))
	}

	return nil
}
