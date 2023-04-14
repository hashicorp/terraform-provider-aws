package lightsail

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	BucketResourceAccessIdPartsCount = 2
)

// @SDKResource("aws_lightsail_bucket_resource_access")
func ResourceBucketResourceAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketResourceAccessCreate,
		ReadWithoutTimeout:   resourceBucketResourceAccessRead,
		DeleteWithoutTimeout: resourceBucketResourceAccessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,52}[a-z0-9]$`), "Invalid Bucket name. Must match regex: ^[a-z0-9][a-z0-9-]{1,52}[a-z0-9]$"),
			},
			"resource_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\w[\w\-]*\w`), "Invalid resource name. Must match regex: \\w[\\w\\-]*\\w"),
			},
		},
	}
}

func resourceBucketResourceAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.SetResourceAccessForBucketInput{
		BucketName:   aws.String(d.Get("bucket_name").(string)),
		ResourceName: aws.String(d.Get("resource_name").(string)),
		Access:       aws.String(lightsail.ResourceBucketAccessAllow),
	}

	out, err := conn.SetResourceAccessForBucketWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeSetResourceAccessForBucket, ResBucketResourceAccess, d.Get("bucket_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeSetResourceAccessForBucket, ResBucketResourceAccess, d.Get("bucket_name").(string))

	if diag != nil {
		return diag
	}

	idParts := []string{d.Get("bucket_name").(string), d.Get("resource_name").(string)}
	id, err := flex.FlattenResourceId(idParts, BucketResourceAccessIdPartsCount)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionFlatteningResourceId, ResBucketResourceAccess, d.Get("bucket_name").(string), err)
	}

	d.SetId(id)

	return resourceBucketResourceAccessRead(ctx, d, meta)
}

func resourceBucketResourceAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindBucketResourceAccessById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResBucketResourceAccess, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucketResourceAccess, d.Id(), err)
	}

	parts, err := flex.ExpandResourceId(d.Id(), BucketResourceAccessIdPartsCount)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResBucketResourceAccess, d.Id(), err)
	}

	d.Set("bucket_name", parts[0])
	d.Set("resource_name", out.Name)

	return nil
}

func resourceBucketResourceAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	parts, err := flex.ExpandResourceId(d.Id(), BucketResourceAccessIdPartsCount)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResBucketResourceAccess, d.Id(), err)
	}

	out, err := conn.SetResourceAccessForBucketWithContext(ctx, &lightsail.SetResourceAccessForBucketInput{
		BucketName:   aws.String(parts[0]),
		ResourceName: aws.String(parts[1]),
		Access:       aws.String(lightsail.ResourceBucketAccessDeny),
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeSetResourceAccessForBucket, ResBucketResourceAccess, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeSetResourceAccessForBucket, ResBucketResourceAccess, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}
