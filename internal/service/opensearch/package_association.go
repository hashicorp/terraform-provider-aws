// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_package_association", name="Package Association")
func resourcePackageAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePackageAssociationCreate,
		ReadWithoutTimeout:   resourcePackageAssociationRead,
		DeleteWithoutTimeout: resourcePackageAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"package_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"reference_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePackageAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	packageID := d.Get("package_id").(string)
	id := fmt.Sprintf("%s-%s", domainName, packageID)
	input := &opensearch.AssociatePackageInput{
		DomainName: aws.String(domainName),
		PackageID:  aws.String(packageID),
	}

	_, err := conn.AssociatePackage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Package Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitPackageAssociationCreated(ctx, conn, domainName, packageID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Package Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePackageAssociationRead(ctx, d, meta)...)
}

func resourcePackageAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	packageID := d.Get("package_id").(string)
	pkgAssociation, err := findPackageAssociationByTwoPartKey(ctx, conn, domainName, packageID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Package Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Package Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDomainName, pkgAssociation.DomainName)
	d.Set("package_id", pkgAssociation.PackageID)
	d.Set("reference_path", pkgAssociation.ReferencePath)

	return diags
}

func resourcePackageAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Package Association: %s", d.Id())
	domainName := d.Get(names.AttrDomainName).(string)
	packageID := d.Get("package_id").(string)
	_, err := conn.DissociatePackage(ctx, &opensearch.DissociatePackageInput{
		DomainName: aws.String(domainName),
		PackageID:  aws.String(packageID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Package is not associated to this domain") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Package Association (%s): %s", d.Id(), err)
	}

	if _, err := waitPackageAssociationDeleted(ctx, conn, domainName, packageID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Package Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findPackageAssociationByTwoPartKey(ctx context.Context, conn *opensearch.Client, domainName, packageID string) (*awstypes.DomainPackageDetails, error) {
	input := &opensearch.ListPackagesForDomainInput{
		DomainName: aws.String(domainName),
	}
	filter := func(v awstypes.DomainPackageDetails) bool {
		return aws.ToString(v.PackageID) == packageID
	}

	return findPackageAssociation(ctx, conn, input, filter)
}

func findPackageAssociation(ctx context.Context, conn *opensearch.Client, input *opensearch.ListPackagesForDomainInput, filter tfslices.Predicate[awstypes.DomainPackageDetails]) (*awstypes.DomainPackageDetails, error) {
	output, err := findPackageAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPackageAssociations(ctx context.Context, conn *opensearch.Client, input *opensearch.ListPackagesForDomainInput, filter tfslices.Predicate[awstypes.DomainPackageDetails]) ([]awstypes.DomainPackageDetails, error) {
	var output []awstypes.DomainPackageDetails

	pages := opensearch.NewListPackagesForDomainPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DomainPackageDetailsList {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusPackageAssociation(ctx context.Context, conn *opensearch.Client, domainName, packageID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPackageAssociationByTwoPartKey(ctx, conn, domainName, packageID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DomainPackageStatus), nil
	}
}

func waitPackageAssociationCreated(ctx context.Context, conn *opensearch.Client, domainName, packageID string, timeout time.Duration) (*awstypes.DomainPackageDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainPackageStatusAssociating),
		Target:  enum.Slice(awstypes.DomainPackageStatusActive),
		Refresh: statusPackageAssociation(ctx, conn, domainName, packageID),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainPackageDetails); ok {
		if status, details := output.DomainPackageStatus, output.ErrorDetails; status == awstypes.DomainPackageStatusAssociationFailed && details != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(details.ErrorType), aws.ToString(details.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitPackageAssociationDeleted(ctx context.Context, conn *opensearch.Client, domainName, packageID string, timeout time.Duration) (*awstypes.DomainPackageDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainPackageStatusDissociating),
		Target:  []string{},
		Refresh: statusPackageAssociation(ctx, conn, domainName, packageID),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainPackageDetails); ok {
		if status, details := output.DomainPackageStatus, output.ErrorDetails; status == awstypes.DomainPackageStatusDissociationFailed && details != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(details.ErrorType), aws.ToString(details.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}
