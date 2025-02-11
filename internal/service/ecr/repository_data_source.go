// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecr_repository", name="Repository")
// @Tags(identifierAttribute="arn")
func dataSourceRepository() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRepositoryRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrKMSKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_on_push": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"image_tag_mutability": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"most_recent_image_tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{name},
	}

	if v, ok := d.GetOk("registry_id"); ok {
		input.RegistryId = aws.String(v.(string))
	}

	repository, err := findRepository(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository (%s): %s", name, err)
	}

	d.SetId(aws.ToString(repository.RepositoryName))
	arn := aws.ToString(repository.RepositoryArn)
	d.Set(names.AttrARN, arn)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenRepositoryEncryptionConfiguration(repository.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting image_scanning_configuration: %s", err)
	}
	d.Set("image_tag_mutability", repository.ImageTagMutability)
	d.Set(names.AttrName, repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)

	imageDetails, err := findImageDetails(ctx, conn, &ecr.DescribeImagesInput{
		RepositoryName: repository.RepositoryName,
		RegistryId:     repository.RegistryId,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading images for ECR Repository (%s): %s", d.Id(), err)
	}

	if len(imageDetails) >= 1 {
		slices.SortFunc(imageDetails, func(a, b types.ImageDetail) int {
			if aws.ToTime(a.ImagePushedAt).After(aws.ToTime(b.ImagePushedAt)) {
				return -1
			}
			if aws.ToTime(a.ImagePushedAt).Before(aws.ToTime(b.ImagePushedAt)) {
				return 1
			}
			return 0
		})

		d.Set("most_recent_image_tags", imageDetails[0].ImageTags)
	} else {
		d.Set("most_recent_image_tags", nil)
	}

	return diags
}
