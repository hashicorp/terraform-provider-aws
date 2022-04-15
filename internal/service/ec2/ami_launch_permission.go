package ec2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAMILaunchPermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAMILaunchPermissionCreate,
		ReadWithoutTimeout:   resourceAMILaunchPermissionRead,
		DeleteWithoutTimeout: resourceAMILaunchPermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAMILaunchPermissionImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"account_id", "group", "organization_arn", "organizational_unit_arn"},
			},
			"group": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.PermissionGroup_Values(), false),
				ExactlyOneOf: []string{"account_id", "group", "organization_arn", "organizational_unit_arn"},
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
				ExactlyOneOf: []string{"account_id", "group", "organization_arn", "organizational_unit_arn"},
			},
			"organizational_unit_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{"account_id", "group", "organization_arn", "organizational_unit_arn"},
			},
		},
	}
}

func resourceAMILaunchPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID := d.Get("image_id").(string)
	accountID := d.Get("account_id").(string)
	group := d.Get("group").(string)
	organizationARN := d.Get("organization_arn").(string)
	organizationalUnitARN := d.Get("organizational_unit_arn").(string)
	id := AMILaunchPermissionCreateResourceID(imageID, accountID, group, organizationARN, organizationalUnitARN)
	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: expandLaunchPermissions(accountID, group, organizationARN, organizationalUnitARN),
		},
	}

	log.Printf("[DEBUG] Creating AMI Launch Permission: %s", input)
	_, err := conn.ModifyImageAttributeWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating AMI Launch Permission (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceAMILaunchPermissionRead(ctx, d, meta)
}

func resourceAMILaunchPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID, accountID, group, organizationARN, organizationalUnitARN, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = FindImageLaunchPermission(ctx, conn, imageID, accountID, group, organizationARN, organizationalUnitARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AMI Launch Permission %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AMI Launch Permission (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("group", group)
	d.Set("image_id", imageID)
	d.Set("organization_arn", organizationARN)
	d.Set("organizational_unit_arn", organizationalUnitARN)

	return nil
}

func resourceAMILaunchPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID, accountID, group, organizationARN, organizationalUnitARN, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: expandLaunchPermissions(accountID, group, organizationARN, organizationalUnitARN),
		},
	}

	log.Printf("[INFO] Deleting AMI Launch Permission: %s", d.Id())
	_, err = conn.ModifyImageAttributeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidAMIIDNotFound, ErrCodeInvalidAMIIDUnavailable) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting AMI Launch Permission (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAMILaunchPermissionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	const importIDSeparator = "/"
	parts := strings.Split(d.Id(), importIDSeparator)

	// Heuristic to identify the permission type.
	var ok bool
	if n := len(parts); n >= 2 {
		if permissionID, imageID := strings.Join(parts[:n-1], importIDSeparator), parts[n-1]; permissionID != "" && imageID != "" {
			if regexp.MustCompile(`^\d{12}$`).MatchString(permissionID) {
				// AWS account ID.
				d.SetId(AMILaunchPermissionCreateResourceID(imageID, permissionID, "", "", ""))
				ok = true
			} else if arn.IsARN(permissionID) {
				if v, _ := arn.Parse(permissionID); v.Service == "organizations" {
					// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsorganizations.html#awsorganizations-resources-for-iam-policies.
					if strings.HasPrefix(v.Resource, "organization/") {
						// Organization ARN.
						d.SetId(AMILaunchPermissionCreateResourceID(imageID, "", "", permissionID, ""))
						ok = true
					} else if strings.HasPrefix(v.Resource, "ou/") {
						// Organizational unit ARN.
						d.SetId(AMILaunchPermissionCreateResourceID(imageID, "", "", "", permissionID))
						ok = true
					}
				}
			} else {
				// Group name.
				d.SetId(AMILaunchPermissionCreateResourceID(imageID, "", permissionID, "", ""))
				ok = true
			}
		}
	}

	if !ok {
		return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected [ACCOUNT-ID|GROUP-NAME|ORGANIZATION-ARN|ORGANIZATIONAL-UNIT-ARN]%[2]sIMAGE-ID", d.Id(), importIDSeparator)
	}

	return []*schema.ResourceData{d}, nil
}

func expandLaunchPermissions(accountID, group, organizationARN, organizationalUnitARN string) []*ec2.LaunchPermission {
	apiObject := &ec2.LaunchPermission{}

	if accountID != "" {
		apiObject.UserId = aws.String(accountID)
	}

	if group != "" {
		apiObject.Group = aws.String(group)
	}

	if organizationARN != "" {
		apiObject.OrganizationArn = aws.String(organizationARN)
	}

	if organizationalUnitARN != "" {
		apiObject.OrganizationalUnitArn = aws.String(organizationalUnitARN)
	}

	return []*ec2.LaunchPermission{apiObject}
}

const (
	amiLaunchPermissionIDSeparator                   = "-"
	amiLaunchPermissionIDGroupIndicator              = "group"
	amiLaunchPermissionIDOrganizationIndicator       = "org"
	amiLaunchPermissionIDOrganizationalUnitIndicator = "ou"
)

func AMILaunchPermissionCreateResourceID(imageID, accountID, group, organizationARN, organizationalUnitARN string) string {
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

func AMILaunchPermissionParseResourceID(id string) (string, string, string, string, string, error) {
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
