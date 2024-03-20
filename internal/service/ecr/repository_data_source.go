// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_ecr_repository")
func DataSourceRepository() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRepositoryRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key": {
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
			"name": {
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{name},
	}

	if v, ok := d.GetOk("registry_id"); ok {
		input.RegistryId = aws.String(v.(string))
	}

	repository, err := FindRepository(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository (%s): %s", name, err)
	}

	d.SetId(aws.ToString(repository.RepositoryName))
	arn := aws.ToString(repository.RepositoryArn)
	d.Set("arn", arn)
	if err := d.Set("encryption_configuration", flattenRepositoryEncryptionConfiguration(repository.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting image_scanning_configuration: %s", err)
	}
	d.Set("image_tag_mutability", repository.ImageTagMutability)
	d.Set("name", repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)

	tags, err := listTags(ctx, conn, arn)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		log.Printf("[WARN] failed listing tags for ECR Repository (%s): %s", d.Id(), err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ECR Repository (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	imageDetails, err := FindImageDetails(ctx, conn, &ecr.DescribeImagesInput{
		RepositoryName: repository.RepositoryName,
		RegistryId:     repository.RegistryId,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading images for ECR Repository (%s): %s", d.Id(), err)
	}

	if len(imageDetails) >= 1 {
		slices.SortFunc(imageDetails, func(a, b *awstypes.ImageDetail) int {
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
