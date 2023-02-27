package wafv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	webACLAssociationCreateTimeout = 5 * time.Minute
)

func ResourceWebACLAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLAssociationCreate,
		ReadWithoutTimeout:   resourceWebACLAssociationRead,
		DeleteWithoutTimeout: resourceWebACLAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceWebACLAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn()

	webACLARN := d.Get("web_acl_arn").(string)
	resourceARN := d.Get("resource_arn").(string)
	id := WebACLAssociationCreateResourceID(webACLARN, resourceARN)
	input := &wafv2.AssociateWebACLInput{
		ResourceArn: aws.String(resourceARN),
		WebACLArn:   aws.String(webACLARN),
	}

	log.Printf("[INFO] Creating WAFv2 WebACL Association: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, webACLAssociationCreateTimeout, func() (interface{}, error) {
		return conn.AssociateWebACLWithContext(ctx, input)
	}, wafv2.ErrCodeWAFUnavailableEntityException)

	if err != nil {
		return diag.Errorf("creating WAFv2 WebACL Association (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceWebACLAssociationRead(ctx, d, meta)
}

func resourceWebACLAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn()

	_, resourceARN, err := WebACLAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	webACL, err := FindWebACLByResourceARN(ctx, conn, resourceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 WebACL Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading WAFv2 WebACL Association (%s): %s", d.Id(), err)
	}

	d.Set("resource_arn", resourceARN)
	d.Set("web_acl_arn", webACL.ARN)

	return nil
}

func resourceWebACLAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn()

	_, resourceARN, err := WebACLAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting WAFv2 WebACL Association: %s", d.Id())
	_, err = conn.DisassociateWebACLWithContext(ctx, &wafv2.DisassociateWebACLInput{
		ResourceArn: aws.String(resourceARN),
	})

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting WAFv2 WebACL Association (%s): %s", d.Id(), err)
	}

	return nil
}

func FindWebACLByResourceARN(ctx context.Context, conn *wafv2.WAFV2, arn string) (*wafv2.WebACL, error) {
	input := &wafv2.GetWebACLForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetWebACLForResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WebACL == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.WebACL, nil
}

const webACLAssociationIDSeparator = ","

func WebACLAssociationCreateResourceID(webACLARN, resourceARN string) string {
	parts := []string{webACLARN, resourceARN}
	id := strings.Join(parts, webACLAssociationIDSeparator)

	return id
}

func WebACLAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, webACLAssociationIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected WEB-ACL-ARN%[2]sRESOURCE-ARN", id, webACLAssociationIDSeparator)
	}

	return parts[0], parts[1], nil
}
