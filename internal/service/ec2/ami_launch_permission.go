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
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected ACCOUNT-ID%[2]sIMAGE-ID", d.Id(), importIDSeparator)
				}

				d.SetId(AMILaunchPermissionCreateResourceID(parts[1], parts[0]))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	id := AMILaunchPermissionCreateResourceID(imageID, accountID)
	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: expandLaunchPermissions(accountID),
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

	imageID, accountID, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = FindImageLaunchPermission(ctx, conn, imageID, accountID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AMI Launch Permission %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AMI Launch Permission (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("image_id", imageID)

	return nil
}

func resourceAMILaunchPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID, accountID, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: expandLaunchPermissions(accountID),
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

func expandLaunchPermissions(accountID string) []*ec2.LaunchPermission {
	apiObject := &ec2.LaunchPermission{}

	if accountID != "" {
		apiObject.UserId = aws.String(accountID)
	}

	return []*ec2.LaunchPermission{apiObject}
}

const amiLaunchPermissionIDSeparator = "-"

func AMILaunchPermissionCreateResourceID(imageID, accountID string) string {
	parts := []string{imageID, accountID}
	id := strings.Join(parts, amiLaunchPermissionIDSeparator)

	return id
}

func AMILaunchPermissionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, amiLaunchPermissionIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return strings.Join([]string{parts[0], parts[1]}, amiLaunchPermissionIDSeparator), parts[2], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected IMAGE-ID%[2]sACCOUNT-ID", id, amiLaunchPermissionIDSeparator)
}
