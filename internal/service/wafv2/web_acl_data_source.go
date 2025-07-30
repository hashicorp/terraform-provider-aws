// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_wafv2_web_acl", name="Web ACL")
func dataSourceWebACL() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWebACLRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{names.AttrName, names.AttrResourceARN},
				},
				names.AttrResourceARN: {
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{names.AttrName, names.AttrResourceARN},
					ValidateFunc: verify.ValidARN,
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
			}
		},
	}
}

func dataSourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	resourceArn := d.Get(names.AttrResourceARN).(string)
	scope := awstypes.Scope(d.Get(names.AttrScope).(string))

	var webACL *awstypes.WebACL
	var err error

	if resourceArn != "" {
		// Check if this is a CloudFront distribution ARN and scope is CLOUDFRONT
		if scope == awstypes.ScopeCloudfront && isCloudFrontDistributionARN(resourceArn) {
			webACL, err = findWebACLByCloudFrontDistributionARN(ctx, meta.(*conns.AWSClient), resourceArn)
		} else {
			// Use GetWebACLForResource API for regional resources
			webACL, err = findWebACLByResourceARN(ctx, conn, resourceArn)
		}
		if err != nil {
			if tfresource.NotFound(err) {
				return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for resource_arn: %s", resourceArn)
			}
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL for resource_arn (%s): %s", resourceArn, err)
		}
	} else {
		// Use existing ListWebACLs + filter by name logic
		var foundWebACL awstypes.WebACLSummary
		input := &wafv2.ListWebACLsInput{
			Scope: scope,
			Limit: aws.Int32(100),
		}

		for {
			resp, err := conn.ListWebACLs(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs: %s", err)
			}

			if resp == nil || resp.WebACLs == nil {
				return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs")
			}

			for _, acl := range resp.WebACLs {
				if aws.ToString(acl.Name) == name {
					foundWebACL = acl
					break
				}
			}

			if resp.NextMarker == nil {
				break
			}
			input.NextMarker = resp.NextMarker
		}

		if foundWebACL.Id == nil {
			return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for name: %s", name)
		}

		// Get full WebACL details using GetWebACL
		getInput := &wafv2.GetWebACLInput{
			Id:    foundWebACL.Id,
			Name:  foundWebACL.Name,
			Scope: scope,
		}

		getResp, err := conn.GetWebACL(ctx, getInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL (%s): %s", aws.ToString(foundWebACL.Id), err)
		}

		webACL = getResp.WebACL
	}

	if webACL == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found")
	}

	d.SetId(aws.ToString(webACL.Id))
	d.Set(names.AttrARN, webACL.ARN)
	d.Set(names.AttrDescription, webACL.Description)
	d.Set(names.AttrName, webACL.Name)

	return diags
}

// Helper function to detect CloudFront distribution ARNs
func isCloudFrontDistributionARN(arn string) bool {
	// CloudFront distribution ARNs: arn:aws:cloudfront::account:distribution/ID
	return strings.Contains(arn, ":cloudfront::") && strings.Contains(arn, ":distribution/")
}

// Helper function to extract distribution ID from CloudFront ARN
func cloudFrontDistributionIDFromARN(arn string) (string, error) {
	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid CloudFront distribution ARN format: %s", arn)
	}
	return parts[len(parts)-1], nil
}

// Helper function to find WebACL by CloudFront distribution ARN
func findWebACLByCloudFrontDistributionARN(ctx context.Context, client *conns.AWSClient, distributionARN string) (*awstypes.WebACL, error) {
	// Extract distribution ID from ARN
	distributionID, err := cloudFrontDistributionIDFromARN(distributionARN)
	if err != nil {
		return nil, err
	}

	// Get CloudFront client
	cfConn := client.CloudFrontClient(ctx)

	// Get distribution configuration
	input := &cloudfront.GetDistributionInput{
		Id: aws.String(distributionID),
	}

	output, err := cfConn.GetDistribution(ctx, input)
	if err != nil {
		if errs.IsA[*cftypes.NoSuchDistribution](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}
		return nil, fmt.Errorf("getting CloudFront distribution (%s): %w", distributionID, err)
	}

	if output.Distribution == nil || output.Distribution.DistributionConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	webACLId := aws.ToString(output.Distribution.DistributionConfig.WebACLId)
	if webACLId == "" {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("no WebACL associated with CloudFront distribution: %s", distributionID),
		}
	}

	// Now get the actual WebACL using WAFv2 API
	wafConn := client.WAFV2Client(ctx)

	// WebACLId can be either an ARN (WAFv2) or ID (WAF Classic)
	// For WAFv2, we need to extract the name and ID from the ARN
	if strings.HasPrefix(webACLId, "arn:aws:wafv2:") {
		return findWebACLByARN(ctx, wafConn, webACLId)
	} else {
		// This would be a WAF Classic ID, not supported by this data source
		return nil, fmt.Errorf("CloudFront distribution (%s) is associated with WAF Classic WebACL (%s), which is not supported by this data source. Use aws_waf_web_acl data source instead", distributionID, webACLId)
	}
}

// Helper function to find WebACL by WAFv2 ARN
func findWebACLByARN(ctx context.Context, conn *wafv2.Client, webACLARN string) (*awstypes.WebACL, error) {
	// Parse the ARN to extract name and ID
	// WAFv2 ARN format: arn:aws:wafv2:region:account:global/webacl/name/id
	parts := strings.Split(webACLARN, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid WAFv2 WebACL ARN format: %s", webACLARN)
	}

	webACLName := parts[len(parts)-2]
	webACLID := parts[len(parts)-1]

	input := &wafv2.GetWebACLInput{
		Id:    aws.String(webACLID),
		Name:  aws.String(webACLName),
		Scope: awstypes.ScopeCloudfront, // CloudFront WebACLs are always global
	}

	output, err := conn.GetWebACL(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}
		return nil, err
	}

	if output.WebACL == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.WebACL, nil
}
