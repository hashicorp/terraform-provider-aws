// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_opensearch_package_association")
func ResourcePackageAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourcePackageAssociationCreate,
		ReadWithoutTimeout:   ResourcePackageAssociationRead,
		DeleteWithoutTimeout: ResourcePackageAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
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

func ResourcePackageAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	input := &opensearchservice.AssociatePackageInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
		PackageID:  aws.String(d.Get("package_id").(string)),
	}

	output, err := conn.AssociatePackageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating OpenSearch package: %s", err)
	}

	d.SetId(fmt.Sprintf("package-association:%s-%s", aws.StringValue(output.DomainPackageDetails.DomainName), aws.StringValue(output.DomainPackageDetails.PackageID)))

	return diags
}

func ResourcePackageAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	var domainName = d.Get("domain_name").(string)
	var packageID = d.Get("package_id").(string)

	output, err := getAssociatedDomainPackage(ctx, domainName, packageID, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing OpenSearch Package for domain (%s): %s", domainName, err)
	}

	d.Set("package_id", output.PackageID)
	d.Set("package_type", output.ReferencePath)
	d.Set("available_package_version", output.PackageVersion)
	return diags
}

func ResourcePackageAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Package: %s", d.Id())
	_, err := conn.DissociatePackageWithContext(ctx, &opensearchservice.DissociatePackageInput{
		PackageID: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "dissociating OpenSearch Package (%s): %s", d.Id(), err)
	}

	return diags
}

func getAssociatedDomainPackage(ctx context.Context, domainID, packageID string, conn *opensearchservice.OpenSearchService) (*opensearchservice.DomainPackageDetails, error) {
	input := &opensearchservice.ListPackagesForDomainInput{
		DomainName: aws.String(domainID),
	}
	for {
		resp, err := conn.ListPackagesForDomainWithContext(ctx, input)
		if err != nil {
			return nil, err
		}
		if len(resp.DomainPackageDetailsList) == 0 {
			return nil, fmt.Errorf("no packages found for domain %s", domainID)
		}
		for _, domainPackageDetails := range resp.DomainPackageDetailsList {
			if aws.StringValue(domainPackageDetails.PackageID) == packageID {
				return domainPackageDetails, nil
			}
		}
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return nil, nil
}
