package lightsail

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_bucket", name="Bucket")
// @Tags(identifierAttribute="id")
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: verify.RegionDiffSuppress,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	region, err := flex.ExpandResourceRegion(d.Get("region"), meta.(*conns.ProviderMeta).AllowedRegions, meta.(*conns.ProviderMeta).Region)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceRegion, ResBucket, d.Get("name").(string), err)
	}

	conn := meta.(*conns.ProviderMeta).AWSClients[region].LightsailConn()

	in := lightsail.CreateBucketInput{
		BucketName: aws.String(d.Get("name").(string)),
		BundleId:   aws.String(d.Get("bundle_id").(string)),
		Tags:       GetTagsIn(ctx),
	}

	out, err := conn.CreateBucketWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateBucket, ResBucket, d.Get("name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeCreateBucket, ResBucket, d.Get("name").(string))

	if diag != nil {
		return diag
	}

	d.SetId(d.Get("name").(string))

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// To allow importing a resource that is not in the provider default region
	// import using terraform import <bucket-name>,<region-name>
	partCount := flex.ResourceIdPartCount(d.Id())
	var id string
	var region string
	if partCount == 2 {
		idParts := strings.Split(d.Id(), flex.ResourceIdSeparator)
		region = idParts[1]
		id = idParts[0]
	} else {
		if v, ok := d.GetOk("region"); ok {
			region = v.(string)
		} else {
			region = meta.(*conns.ProviderMeta).Region
		}
		id = d.Id()
	}

	conn := meta.(*conns.ProviderMeta).AWSClients[region].LightsailConn()

	out, err := FindBucketById(ctx, conn, id)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResBucket, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucket, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("availability_zone", out.Location.AvailabilityZone)
	d.Set("bundle_id", out.BundleId)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("name", out.Name)
	d.Set("region", out.Location.RegionName)
	d.Set("support_code", out.SupportCode)
	d.Set("url", out.Url)

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	region, err := flex.ExpandResourceRegion(d.Get("region"), meta.(*conns.ProviderMeta).AllowedRegions, meta.(*conns.ProviderMeta).Region)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceRegion, ResBucket, d.Get("name").(string), err)
	}

	conn := meta.(*conns.ProviderMeta).AWSClients[region].LightsailConn()

	if d.HasChange("bundle_id") {
		in := lightsail.UpdateBucketBundleInput{
			BucketName: aws.String(d.Id()),
			BundleId:   aws.String(d.Get("bundle_id").(string)),
		}
		out, err := conn.UpdateBucketBundleWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateBucket, ResBucket, d.Get("name").(string), err)
		}

		diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeUpdateBucket, ResBucket, d.Get("name").(string))

		if diag != nil {
			return diag
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	region, err := flex.ExpandResourceRegion(d.Get("region"), meta.(*conns.ProviderMeta).AllowedRegions, meta.(*conns.ProviderMeta).Region)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceRegion, ResBucket, d.Get("name").(string), err)
	}

	conn := meta.(*conns.ProviderMeta).AWSClients[region].LightsailConn()

	out, err := conn.DeleteBucketWithContext(ctx, &lightsail.DeleteBucketInput{
		BucketName: aws.String(d.Id()),
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionDeleting, ResBucket, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeDeleteBucket, ResBucket, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}
