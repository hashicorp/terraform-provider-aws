// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package wafv2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
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
			if retry.NotFound(err) {
				return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for resource_arn: %s", resourceArn)
			}
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL for resource_arn (%s): %s", resourceArn, err)
		}
	} else {
		// Use existing ListWebACLs + filter by name logic
		var foundWebACL awstypes.WebACLSummary
		input := wafv2.ListWebACLsInput{
			Scope: scope,
			Limit: aws.Int32(100),
		}

		err := listWebACLsPages(ctx, conn, &input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, acl := range page.WebACLs {
				if aws.ToString(acl.Name) == name {
					foundWebACL = acl
					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "list WAFv2 WebACLs: %s", err)
		}

		if foundWebACL.Id == nil {
			return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for name: %s", name)
		}

		// Get full WebACL details using GetWebACL
		getResp, err := findWebACLByThreePartKey(ctx, conn, aws.ToString(foundWebACL.Id), aws.ToString(foundWebACL.Name), string(scope))

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
func isCloudFrontDistributionARN(s string) bool {
	// CloudFront distribution ARNs: arn:partition:cloudfront::account:distribution/ID
	return strings.Contains(s, ":cloudfront::") && strings.Contains(s, ":distribution/") && arn.IsARN(s)
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

	output, err := tfcloudfront.FindDistributionByID(ctx, client.CloudFrontClient(ctx), distributionID)

	if err != nil {
		return nil, fmt.Errorf("getting CloudFront distribution (%s): %w", distributionID, err)
	}

	webACLARN := aws.ToString(output.Distribution.DistributionConfig.WebACLId)
	if webACLARN == "" {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("no WebACL associated with CloudFront distribution: %s", distributionID),
		}
	}

	// Now get the actual WebACL using WAFv2 API
	wafConn := client.WAFV2Client(ctx)

	if !strings.Contains(webACLARN, ":wafv2:") || !arn.IsARN(webACLARN) {
		// This would be a WAF Classic ID, not supported by this data source
		return nil, fmt.Errorf("CloudFront distribution (%s) is associated with WAF Classic WebACL (%s), which is not supported by this data source. Use aws_waf_web_acl data source instead", distributionID, webACLARN)
	}

	// Parse the ARN to extract name and ID
	// WAFv2 ARN format: arn:partition:wafv2:region:account:global/webacl/name/id
	parts := strings.Split(webACLARN, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid WAFv2 WebACL ARN format: %s", webACLARN)
	}

	webACLName := parts[len(parts)-2]
	webACLID := parts[len(parts)-1]

	var webACLOut *wafv2.GetWebACLOutput
	if webACLOut, err = findWebACLByThreePartKey(ctx, wafConn, webACLID, webACLName, string(awstypes.ScopeCloudfront)); err != nil {
		return nil, fmt.Errorf("finding WAFv2 WebACL (%s): %w", webACLARN, err)
	}
	if webACLOut == nil {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("no WAFv2 WebACL found: %s", webACLARN),
		}
	}

	return webACLOut.WebACL, nil
}
