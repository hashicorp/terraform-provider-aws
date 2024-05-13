// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_package")
func ResourcePackage() *schema.Resource {
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(opensearchservice.PackageType_Values(), false),
			},
		},
	}
}

func resourcePackageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	name := d.Get("package_name").(string)
	input := &opensearchservice.CreatePackageInput{
		PackageDescription: aws.String(d.Get("package_description").(string)),
		PackageName:        aws.String(name),
		PackageType:        aws.String(d.Get("package_type").(string)),
	}

	if v, ok := d.GetOk("package_source"); ok {
		input.PackageSource = expandPackageSource(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreatePackageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Package (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.PackageDetails.PackageID))

	return append(diags, resourcePackageRead(ctx, d, meta)...)
}

func resourcePackageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	pkg, err := FindPackageByID(ctx, conn, d.Id())

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

func resourcePackageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	input := &opensearchservice.UpdatePackageInput{
		PackageID:          aws.String(d.Id()),
		PackageDescription: aws.String(d.Get("package_description").(string)),
		PackageSource:      expandPackageSource(d.Get("package_source").([]interface{})[0].(map[string]interface{})),
	}

	_, err := conn.UpdatePackageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpenSearch Package (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePackageRead(ctx, d, meta)...)
}

func resourcePackageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Package: %s", d.Id())
	_, err := conn.DeletePackageWithContext(ctx, &opensearchservice.DeletePackageInput{
		PackageID: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Package (%s): %s", d.Id(), err)
	}

	return diags
}

func FindPackageByID(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) (*opensearchservice.PackageDetails, error) {
	input := &opensearchservice.DescribePackagesInput{
		Filters: []*opensearchservice.DescribePackagesFilter{
			{
				Name:  aws.String("PackageID"),
				Value: aws.StringSlice([]string{id}),
			},
		},
	}

	return findPackage(ctx, conn, input)
}

func findPackage(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribePackagesInput) (*opensearchservice.PackageDetails, error) {
	output, err := findPackages(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findPackages(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribePackagesInput) ([]*opensearchservice.PackageDetails, error) {
	var output []*opensearchservice.PackageDetails

	err := conn.DescribePackagesPagesWithContext(ctx, input, func(page *opensearchservice.DescribePackagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PackageDetailsList {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
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

func expandPackageSource(v interface{}) *opensearchservice.PackageSource {
	if v == nil {
		return nil
	}

	return &opensearchservice.PackageSource{
		S3BucketName: aws.String(v.(map[string]interface{})[names.AttrS3BucketName].(string)),
		S3Key:        aws.String(v.(map[string]interface{})["s3_key"].(string)),
	}
}
