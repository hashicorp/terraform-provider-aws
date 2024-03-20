// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
			"image_uri": {
				Type:     schema.TypeString,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
	}

	if v, ok := d.GetOk("image_digest"); ok {
		input.ImageIds = []awstypes.ImageIdentifier{
			{
				ImageDigest: aws.String(v.(string)),
			},
		}
	}

	if v, ok := d.GetOk("image_tag"); ok {
		if len(input.ImageIds) == 0 {
			input.ImageIds = []awstypes.ImageIdentifier{
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

		slices.SortFunc(imageDetails, func(a, b *awstypes.ImageDetail) int {
			if aws.ToTime(a.ImagePushedAt).After(aws.ToTime(b.ImagePushedAt)) {
				return -1
			}
			if aws.ToTime(a.ImagePushedAt).Before(aws.ToTime(b.ImagePushedAt)) {
				return 1
			}
			return 0
		})
	}

	imageDetail := imageDetails[0]

	repositoryName := aws.ToString(imageDetail.RepositoryName)
	repositoryInput := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repositoryName},
		RegistryId:      imageDetail.RegistryId,
	}

	repository, err := FindRepository(ctx, conn, repositoryInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Images: %s", err)
	}

	d.SetId(aws.ToString(imageDetail.ImageDigest))
	d.Set("image_digest", imageDetail.ImageDigest)
	d.Set("image_pushed_at", imageDetail.ImagePushedAt.Unix())
	d.Set("image_size_in_bytes", imageDetail.ImageSizeInBytes)
	d.Set("image_tags", imageDetail.ImageTags)
	d.Set("image_uri", fmt.Sprintf("%s@%s", aws.ToString(repository.RepositoryUri), aws.ToString(imageDetail.ImageDigest)))
	d.Set("registry_id", imageDetail.RegistryId)
	d.Set("repository_name", imageDetail.RepositoryName)

	return diags
}

func FindImageDetails(ctx context.Context, conn *ecr.Client, input *ecr.DescribeImagesInput) ([]*awstypes.ImageDetail, error) {
	var output []*awstypes.ImageDetail

	err := describeImagesPages(ctx, conn, input, func(page *ecr.DescribeImagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ImageDetails {
			imageDetail := v
			output = append(output, &imageDetail)
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
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
