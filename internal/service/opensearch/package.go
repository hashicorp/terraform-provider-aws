// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_package", name="Package")
func resourcePackage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePackageCreate,
		ReadWithoutTimeout:   resourcePackageRead,
		UpdateWithoutTimeout: resourcePackageUpdate,
		DeleteWithoutTimeout: resourcePackageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"available_package_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"package_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"package_source": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrS3BucketName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"s3_key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"package_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PackageType](),
			},
		},
	}
}

func resourcePackageCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	name := d.Get("package_name").(string)
	input := &opensearch.CreatePackageInput{
		PackageDescription: aws.String(d.Get("package_description").(string)),
		PackageName:        aws.String(name),
		PackageType:        awstypes.PackageType(d.Get("package_type").(string)),
	}

	if v, ok := d.GetOk("package_source"); ok {
		input.PackageSource = expandPackageSource(v.([]any)[0].(map[string]any))
	}

	output, err := conn.CreatePackage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Package (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PackageDetails.PackageID))

	return append(diags, resourcePackageRead(ctx, d, meta)...)
}

func resourcePackageRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	pkg, err := findPackageByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Package (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Package (%s): %s", d.Id(), err)
	}

	d.Set("available_package_version", pkg.AvailablePackageVersion)
	d.Set("package_description", pkg.PackageDescription)
	d.Set("package_id", pkg.PackageID)
	d.Set("package_name", pkg.PackageName)
	d.Set("package_type", pkg.PackageType)

	return diags
}

func resourcePackageUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	input := &opensearch.UpdatePackageInput{
		PackageID:          aws.String(d.Id()),
		PackageDescription: aws.String(d.Get("package_description").(string)),
		PackageSource:      expandPackageSource(d.Get("package_source").([]any)[0].(map[string]any)),
	}

	_, err := conn.UpdatePackage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpenSearch Package (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePackageRead(ctx, d, meta)...)
}

func resourcePackageDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Package: %s", d.Id())
	_, err := conn.DeletePackage(ctx, &opensearch.DeletePackageInput{
		PackageID: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Package not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Package (%s): %s", d.Id(), err)
	}

	return diags
}

func findPackageByID(ctx context.Context, conn *opensearch.Client, id string) (*awstypes.PackageDetails, error) {
	input := &opensearch.DescribePackagesInput{
		Filters: []awstypes.DescribePackagesFilter{
			{
				Name:  awstypes.DescribePackagesFilterNamePackageID,
				Value: []string{id},
			},
		},
	}

	return findPackage(ctx, conn, input)
}

func findPackage(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribePackagesInput) (*awstypes.PackageDetails, error) {
	output, err := findPackages(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPackages(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribePackagesInput) ([]awstypes.PackageDetails, error) {
	var output []awstypes.PackageDetails

	pages := opensearch.NewDescribePackagesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Package not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PackageDetailsList...)
	}

	return output, nil
}

func expandPackageSource(v any) *awstypes.PackageSource {
	if v == nil {
		return nil
	}

	return &awstypes.PackageSource{
		S3BucketName: aws.String(v.(map[string]any)[names.AttrS3BucketName].(string)),
		S3Key:        aws.String(v.(map[string]any)["s3_key"].(string)),
	}
}
