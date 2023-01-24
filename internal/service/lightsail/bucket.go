package lightsail

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,
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
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	in := lightsail.CreateBucketInput{
		BucketName: aws.String(d.Get("name").(string)),
		BundleId:   aws.String(d.Get("bundle_id").(string)),
	}

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateBucketWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateBucket, ResBucket, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateBucket, ResBucket, d.Get("name").(string), errors.New("No operations found for Create Bucket request"))
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateBucket, ResBucket, d.Get("name").(string), errors.New("Error waiting for Create Bucket request operation"))
	}

	d.SetId(d.Get("name").(string))

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindBucketById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CE, create.ErrActionReading, ResBucket, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.CE, create.ErrActionReading, ResBucket, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("availability_zone", out.Location.AvailabilityZone)
	d.Set("bundle_id", out.BundleId)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("name", out.Name)
	d.Set("region", out.Location.RegionName)
	d.Set("support_code", out.SupportCode)
	d.Set("url", out.Url)

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucket, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucket, d.Id(), err)
	}

	return nil
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResBucket, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResBucket, d.Id(), err)
		}
	}

	if d.HasChange("bundle_id") {
		in := lightsail.UpdateBucketBundleInput{
			BucketName: aws.String(d.Id()),
			BundleId:   aws.String(d.Get("bundle_id").(string)),
		}
		out, err := conn.UpdateBucketBundleWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateBucket, ResBucket, d.Get("name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateBucket, ResBucket, d.Get("name").(string), errors.New("No operations found for request"))
		}

		op := out.Operations[0]

		err = waitOperation(ctx, conn, op.Id)
		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateBucket, ResBucket, d.Get("name").(string), errors.New("Error waiting for request operation"))
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	out, err := conn.DeleteBucketWithContext(ctx, &lightsail.DeleteBucketInput{
		BucketName: aws.String(d.Id()),
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.CE, create.ErrActionDeleting, ResBucket, d.Id(), err)
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteCertificate, ResBucket, d.Id(), err)
	}

	return nil
}
