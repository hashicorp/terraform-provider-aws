package wafv2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	WebACLAssociationCreateTimeout = 5 * time.Minute
)

func ResourceWebACLAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebACLAssociationCreate,
		Read:   resourceWebACLAssociationRead,
		Delete: resourceWebACLAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				webAclArn, resourceArn, err := resourceACLAssociationDecodeID(d.Id())
				if err != nil {
					return nil, fmt.Errorf("Error reading resource ID: %s", err)
				}
				d.Set("resource_arn", resourceArn)
				d.Set("web_acl_arn", webAclArn)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"web_acl_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceWebACLAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	resourceArn := d.Get("resource_arn").(string)
	webAclArn := d.Get("web_acl_arn").(string)
	params := &wafv2.AssociateWebACLInput{
		ResourceArn: aws.String(resourceArn),
		WebACLArn:   aws.String(webAclArn),
	}

	err := resource.Retry(WebACLAssociationCreateTimeout, func() *resource.RetryError {
		_, err := conn.AssociateWebACL(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFUnavailableEntityException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.AssociateWebACL(params)
	}

	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s,%s", webAclArn, resourceArn))

	return resourceWebACLAssociationRead(d, meta)
}

func resourceWebACLAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	resourceArn := d.Get("resource_arn").(string)
	webAclArn := d.Get("web_acl_arn").(string)
	params := &wafv2.GetWebACLForResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	resp, err := conn.GetWebACLForResource(params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] WAFv2 Web ACL (%s) not found, removing from state", webAclArn)
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.WebACL == nil {
		log.Printf("[WARN] WAFv2 Web ACL associated resource (%s) not found, removing from state", resourceArn)
		d.SetId("")
	}

	return nil
}

func resourceWebACLAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	log.Printf("[INFO] Deleting WAFv2 Web ACL Association %s", d.Id())

	params := &wafv2.DisassociateWebACLInput{
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	_, err := conn.DisassociateWebACL(params)
	if err != nil {
		return fmt.Errorf("Error disassociating WAFv2 Web ACL: %s", err)
	}

	return nil
}

func resourceACLAssociationDecodeID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID (%s), expected WEB-ACL-ARN,RESOURCE-ARN", id)
	}

	return parts[0], parts[1], nil
}
