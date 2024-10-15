// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ami_launch_permission", name="AMI Launch Permission")
func resourceAMILaunchPermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAMILaunchPermissionCreate,
		ReadWithoutTimeout:   resourceAMILaunchPermissionRead,
		DeleteWithoutTimeout: resourceAMILaunchPermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAMILaunchPermissionImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrAccountID, "group", "organization_arn", "organizational_unit_arn"},
			},
			"group": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PermissionGroup](),
				ExactlyOneOf:     []string{names.AttrAccountID, "group", "organization_arn", "organizational_unit_arn"},
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organization_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{names.AttrAccountID, "group", "organization_arn", "organizational_unit_arn"},
			},
			"organizational_unit_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{names.AttrAccountID, "group", "organization_arn", "organizational_unit_arn"},
			},
		},
	}
}

func resourceAMILaunchPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	imageID := d.Get("image_id").(string)
	accountID := d.Get(names.AttrAccountID).(string)
	group := d.Get("group").(string)
	organizationARN := d.Get("organization_arn").(string)
	organizationalUnitARN := d.Get("organizational_unit_arn").(string)
	id := amiLaunchPermissionCreateResourceID(imageID, accountID, group, organizationARN, organizationalUnitARN)
	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(string(awstypes.ImageAttributeNameLaunchPermission)),
		ImageId:   aws.String(imageID),
		LaunchPermission: &awstypes.LaunchPermissionModifications{
			Add: expandLaunchPermissions(accountID, group, organizationARN, organizationalUnitARN),
		},
	}

	_, err := conn.ModifyImageAttribute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AMI Launch Permission (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAMILaunchPermissionRead(ctx, d, meta)...)
}

func resourceAMILaunchPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	imageID, accountID, group, organizationARN, organizationalUnitARN, err := amiLaunchPermissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findImageLaunchPermission(ctx, conn, imageID, accountID, group, organizationARN, organizationalUnitARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AMI Launch Permission %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AMI Launch Permission (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set("group", group)
	d.Set("image_id", imageID)
	d.Set("organization_arn", organizationARN)
	d.Set("organizational_unit_arn", organizationalUnitARN)

	return diags
}

func resourceAMILaunchPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	imageID, accountID, group, organizationARN, organizationalUnitARN, err := amiLaunchPermissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(string(awstypes.ImageAttributeNameLaunchPermission)),
		ImageId:   aws.String(imageID),
		LaunchPermission: &awstypes.LaunchPermissionModifications{
			Remove: expandLaunchPermissions(accountID, group, organizationARN, organizationalUnitARN),
		},
	}

	log.Printf("[INFO] Deleting AMI Launch Permission: %s", d.Id())
	_, err = conn.ModifyImageAttribute(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound, errCodeInvalidAMIIDUnavailable) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AMI Launch Permission (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAMILaunchPermissionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	const importIDSeparator = "/"
	parts := strings.Split(d.Id(), importIDSeparator)

	// Heuristic to identify the permission type.
	var ok bool
	if n := len(parts); n >= 2 {
		if permissionID, imageID := strings.Join(parts[:n-1], importIDSeparator), parts[n-1]; permissionID != "" && imageID != "" {
			if regexache.MustCompile(`^\d{12}$`).MatchString(permissionID) {
				// AWS account ID.
				d.SetId(amiLaunchPermissionCreateResourceID(imageID, permissionID, "", "", ""))
				ok = true
			} else if arn.IsARN(permissionID) {
				if v, _ := arn.Parse(permissionID); v.Service == "organizations" {
					// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsorganizations.html#awsorganizations-resources-for-iam-policies.
					if strings.HasPrefix(v.Resource, "organization/") {
						// Organization ARN.
						d.SetId(amiLaunchPermissionCreateResourceID(imageID, "", "", permissionID, ""))
						ok = true
					} else if strings.HasPrefix(v.Resource, "ou/") {
						// Organizational unit ARN.
						d.SetId(amiLaunchPermissionCreateResourceID(imageID, "", "", "", permissionID))
						ok = true
					}
				}
			} else {
				// Group name.
				d.SetId(amiLaunchPermissionCreateResourceID(imageID, "", permissionID, "", ""))
				ok = true
			}
		}
	}

	if !ok {
		return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected [ACCOUNT-ID|GROUP-NAME|ORGANIZATION-ARN|ORGANIZATIONAL-UNIT-ARN]%[2]sIMAGE-ID", d.Id(), importIDSeparator)
	}

	return []*schema.ResourceData{d}, nil
}

const (
	amiLaunchPermissionIDSeparator                   = "-"
	amiLaunchPermissionIDGroupIndicator              = "group"
	amiLaunchPermissionIDOrganizationIndicator       = "org"
	amiLaunchPermissionIDOrganizationalUnitIndicator = "ou"
)

func amiLaunchPermissionCreateResourceID(imageID, accountID, group, organizationARN, organizationalUnitARN string) string {
	parts := []string{imageID}

	if accountID != "" {
		parts = append(parts, accountID)
	} else if group != "" {
		parts = append(parts, amiLaunchPermissionIDGroupIndicator, group)
	} else if organizationARN != "" {
		parts = append(parts, amiLaunchPermissionIDOrganizationIndicator, organizationARN)
	} else if organizationalUnitARN != "" {
		parts = append(parts, amiLaunchPermissionIDOrganizationalUnitIndicator, organizationalUnitARN)
	}

	id := strings.Join(parts, amiLaunchPermissionIDSeparator)

	return id
}

func amiLaunchPermissionParseResourceID(id string) (string, string, string, string, string, error) {
	parts := strings.Split(id, amiLaunchPermissionIDSeparator)

	switch {
	case len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "":
		return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), parts[2], "", "", "", nil
	case len(parts) > 3 && parts[0] != "" && parts[1] != "" && parts[3] != "":
		switch parts[2] {
		case amiLaunchPermissionIDGroupIndicator:
			return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), "", strings.Join(parts[3:], amiLaunchPermissionIDSeparator), "", "", nil
		case amiLaunchPermissionIDOrganizationIndicator:
			return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), "", "", strings.Join(parts[3:], amiLaunchPermissionIDSeparator), "", nil
		case amiLaunchPermissionIDOrganizationalUnitIndicator:
			return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), "", "", "", strings.Join(parts[3:], amiLaunchPermissionIDSeparator), nil
		}
	}

	return "", "", "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected IMAGE-ID%[2]sACCOUNT-ID or IMAGE-ID%[2]s%[3]s%[2]sGROUP-NAME or IMAGE-ID%[2]s%[4]s%[2]sORGANIZATION-ARN or IMAGE-ID%[2]s%[5]s%[2]sORGANIZATIONAL-UNIT-ARN", id, amiLaunchPermissionIDSeparator, amiLaunchPermissionIDGroupIndicator, amiLaunchPermissionIDOrganizationIndicator, amiLaunchPermissionIDOrganizationalUnitIndicator)
}

func expandLaunchPermissions(accountID, group, organizationARN, organizationalUnitARN string) []awstypes.LaunchPermission {
	apiObject := awstypes.LaunchPermission{}

	if accountID != "" {
		apiObject.UserId = aws.String(accountID)
	}

	if group != "" {
		apiObject.Group = awstypes.PermissionGroup(group)
	}

	if organizationARN != "" {
		apiObject.OrganizationArn = aws.String(organizationARN)
	}

	if organizationalUnitARN != "" {
		apiObject.OrganizationalUnitArn = aws.String(organizationalUnitARN)
	}

	return []awstypes.LaunchPermission{apiObject}
}
