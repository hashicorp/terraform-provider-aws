package ecr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image_digest": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"image_tag": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"image_pushed_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	params := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
	}

	regId, ok := d.GetOk("registry_id")
	if ok {
		params.RegistryId = aws.String(regId.(string))
	}

	imgId := ecr.ImageIdentifier{}
	digest, ok := d.GetOk("image_digest")
	if ok {
		imgId.ImageDigest = aws.String(digest.(string))
	}
	tag, ok := d.GetOk("image_tag")
	if ok {
		imgId.ImageTag = aws.String(tag.(string))
	}

	if imgId.ImageDigest == nil && imgId.ImageTag == nil {
		return sdkdiag.AppendErrorf(diags, "At least one of either image_digest or image_tag must be defined")
	}

	params.ImageIds = []*ecr.ImageIdentifier{&imgId}

	var imageDetails []*ecr.ImageDetail
	log.Printf("[DEBUG] Reading ECR Images: %s", params)
	err := conn.DescribeImagesPagesWithContext(ctx, params, func(page *ecr.DescribeImagesOutput, lastPage bool) bool {
		imageDetails = append(imageDetails, page.ImageDetails...)
		return true
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing ECR images: %s", err)
	}

	if len(imageDetails) == 0 {
		return sdkdiag.AppendErrorf(diags, "No matching image found")
	}
	if len(imageDetails) > 1 {
		return sdkdiag.AppendErrorf(diags, "More than one image found for tag/digest combination")
	}

	image := imageDetails[0]

	d.SetId(aws.StringValue(image.ImageDigest))
	d.Set("registry_id", image.RegistryId)
	d.Set("image_digest", image.ImageDigest)
	d.Set("image_pushed_at", image.ImagePushedAt.Unix())
	d.Set("image_size_in_bytes", image.ImageSizeInBytes)
	if err := d.Set("image_tags", aws.StringValueSlice(image.ImageTags)); err != nil {
		return sdkdiag.AppendErrorf(diags, "to set image_tags: %s", err)
	}

	return diags
}
