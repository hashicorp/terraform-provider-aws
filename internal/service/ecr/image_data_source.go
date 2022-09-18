package ecr

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImageRead,
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
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func dataSourceImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	params := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
	}

	regId, ok := d.GetOk("registry_id")
	if ok {
		params.RegistryId = aws.String(regId.(string))
	}

	findMostRecent := false

	if _, ok := d.GetOk("most_recent"); ok {
		findMostRecent = true
	}

	imgId := ecr.ImageIdentifier{}
	if v, ok := d.GetOk("image_digest"); ok {
		if findMostRecent {
			return fmt.Errorf("cannot set image_digest when most_recent is set to true")
		}
		imgId.ImageDigest = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_tag"); ok {
		if findMostRecent {
			return fmt.Errorf("cannot set image_tag when most_recent is set to true")
		}
		imgId.ImageTag = aws.String(v.(string))
	}

	if !findMostRecent {
		params.ImageIds = []*ecr.ImageIdentifier{&imgId}
	}

	log.Printf("[DEBUG] Reading ECR Images: %s", params)
	imageDetails, err := FindImageDetails(conn, params)
	if err != nil {
		return fmt.Errorf("Error describing ECR images: %w", err)
	}

	if len(imageDetails) == 0 {
		return fmt.Errorf("No matching image found")
	}

	if len(imageDetails) > 1 && !findMostRecent {
		return fmt.Errorf("More than one image found for tag/digest combination")
	}

	sort.Sort(sort.Reverse(byImagePushedAt(imageDetails)))

	image := imageDetails[0]

	d.SetId(aws.StringValue(image.ImageDigest))
	d.Set("registry_id", image.RegistryId)
	d.Set("image_digest", image.ImageDigest)
	d.Set("image_pushed_at", image.ImagePushedAt.Unix())
	d.Set("image_size_in_bytes", image.ImageSizeInBytes)
	if err := d.Set("image_tags", aws.StringValueSlice(image.ImageTags)); err != nil {
		return fmt.Errorf("failed to set image_tags: %w", err)
	}

	return nil
}
