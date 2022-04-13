package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAMILaunchPermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAMILaunchPermissionCreate,
		ReadWithoutTimeout:   resourceAMILaunchPermissionRead,
		DeleteWithoutTimeout: resourceAMILaunchPermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				const importIDSeparator = "/"
				parts := strings.Split(d.Id(), importIDSeparator)

				switch {
				case len(parts) == 2 && parts[0] != "" && parts[1] != "":
					d.SetId(AMILaunchPermissionCreateResourceID(parts[1], parts[0], ""))
				case len(parts) == 3 && parts[0] == "group" && parts[1] != "" && parts[2] != "":
					d.SetId(AMILaunchPermissionCreateResourceID(parts[2], "", parts[1]))
				default:
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected ACCOUNT-ID%[2]sIMAGE-ID or group%[2]sGROUP-NAME%[2]sIMAGE-ID", d.Id(), importIDSeparator)
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"account_id", "group"},
			},
			"group": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.PermissionGroup_Values(), false),
				ExactlyOneOf: []string{"account_id", "group"},
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAMILaunchPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID := d.Get("image_id").(string)
	accountID := d.Get("account_id").(string)
	group := d.Get("group").(string)
	id := AMILaunchPermissionCreateResourceID(imageID, accountID, group)
	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: expandLaunchPermissions(accountID, group),
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

	imageID, accountID, group, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = FindImageLaunchPermission(ctx, conn, imageID, accountID, group)

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

	return nil
}

func resourceAMILaunchPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID, accountID, group, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: expandLaunchPermissions(accountID, group),
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

func expandLaunchPermissions(accountID, group string) []*ec2.LaunchPermission {
	apiObject := &ec2.LaunchPermission{}

	if accountID != "" {
		apiObject.UserId = aws.String(accountID)
	}

	if group != "" {
		apiObject.Group = aws.String(group)
	}

	return []*ec2.LaunchPermission{apiObject}
}

const (
	amiLaunchPermissionIDSeparator      = "-"
	amiLaunchPermissionIDGroupIndicator = "group"
)

func AMILaunchPermissionCreateResourceID(imageID, accountID, group string) string {
	parts := []string{imageID}

	if accountID != "" {
		parts = append(parts, accountID)
	} else if group != "" {
		parts = append(parts, amiLaunchPermissionIDGroupIndicator, group)
	}

	id := strings.Join(parts, amiLaunchPermissionIDSeparator)

	return id
}

func AMILaunchPermissionParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, amiLaunchPermissionIDSeparator)

	switch {
	case len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "":
		return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), parts[2], "", nil
	case len(parts) > 3 && parts[0] != "" && parts[1] != "" && parts[3] != "":
		switch parts[2] {
		case amiLaunchPermissionIDGroupIndicator:
			return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), "", strings.Join(parts[3:], amiLaunchPermissionIDSeparator), nil
		}
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected IMAGE-ID%[2]sACCOUNT-ID or IMAGE-ID%[2]sgroup%[2]sGROUP-NAME", id, amiLaunchPermissionIDSeparator)
}
