// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"golang.org/x/exp/slices"
)

// @SDKDataSource("aws_ecr_image")
func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			"image_digest": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				AtLeastOneOf:  []string{"image_digest", "image_tag", "most_recent"},
				ConflictsWith: []string{"most_recent"},
			},
			"image_pushed_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_tag": {
				Type:          schema.TypeString,
				Optional:      true,
				AtLeastOneOf:  []string{"image_digest", "image_tag", "most_recent"},
				ConflictsWith: []string{"most_recent"},
			},
			"image_tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"most_recent": {
				Type:          schema.TypeBool,
				Optional:      true,
				AtLeastOneOf:  []string{"image_digest", "image_tag", "most_recent"},
				ConflictsWith: []string{"image_digest", "image_tag"},
			},
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
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
	}

	if v, ok := d.GetOk("image_digest"); ok {
		input.ImageIds = []*ecr.ImageIdentifier{
			{
				ImageDigest: aws.String(v.(string)),
			},
		}
	}

	if v, ok := d.GetOk("image_tag"); ok {
		if len(input.ImageIds) == 0 {
			input.ImageIds = []*ecr.ImageIdentifier{
				{
					ImageTag: aws.String(v.(string)),
				},
			}
		} else {
			input.ImageIds[0].ImageTag = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("registry_id"); ok {
		input.RegistryId = aws.String(v.(string))
	}

	imageDetails, err := FindImageDetails(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Images: %s", err)
	}

	if len(imageDetails) == 0 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(imageDetails) > 1 {
		if !d.Get("most_recent").(bool) {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more specific search criteria, or set `most_recent` attribute to true.")
		}

		slices.SortFunc(imageDetails, func(a, b *ecr.ImageDetail) bool {
			return aws.TimeValue(a.ImagePushedAt).After(aws.TimeValue(b.ImagePushedAt))
		})
	}

	imageDetail := imageDetails[0]
	d.SetId(aws.StringValue(imageDetail.ImageDigest))
	d.Set("image_digest", imageDetail.ImageDigest)
	d.Set("image_pushed_at", imageDetail.ImagePushedAt.Unix())
	d.Set("image_size_in_bytes", imageDetail.ImageSizeInBytes)
	d.Set("image_tags", aws.StringValueSlice(imageDetail.ImageTags))
	d.Set("registry_id", imageDetail.RegistryId)
	d.Set("repository_name", imageDetail.RepositoryName)

	return diags
}

func FindImageDetails(ctx context.Context, conn *ecr.ECR, input *ecr.DescribeImagesInput) ([]*ecr.ImageDetail, error) {
	var output []*ecr.ImageDetail

	err := conn.DescribeImagesPagesWithContext(ctx, input, func(page *ecr.DescribeImagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ImageDetails {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
