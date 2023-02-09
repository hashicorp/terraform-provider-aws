package worklink

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceWebsiteCertificateAuthorityAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebsiteCertificateAuthorityAssociationCreate,
		ReadWithoutTimeout:   resourceWebsiteCertificateAuthorityAssociationRead,
		DeleteWithoutTimeout: resourceWebsiteCertificateAuthorityAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"fleet_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"website_ca_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWebsiteCertificateAuthorityAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkConn()

	input := &worklink.AssociateWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(d.Get("fleet_arn").(string)),
		Certificate: aws.String(d.Get("certificate").(string)),
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	resp, err := conn.AssociateWebsiteCertificateAuthorityWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Website Certificate Authority Association: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", d.Get("fleet_arn").(string), aws.StringValue(resp.WebsiteCaId)))

	return append(diags, resourceWebsiteCertificateAuthorityAssociationRead(ctx, d, meta)...)
}

func resourceWebsiteCertificateAuthorityAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkConn()

	fleetArn, websiteCaID, err := DecodeWebsiteCertificateAuthorityAssociationResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	input := &worklink.DescribeWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(fleetArn),
		WebsiteCaId: aws.String(websiteCaID),
	}

	resp, err := conn.DescribeWebsiteCertificateAuthorityWithContext(ctx, input)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] WorkLink Website Certificate Authority Association (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	d.Set("website_ca_id", websiteCaID)
	d.Set("fleet_arn", fleetArn)
	d.Set("certificate", resp.Certificate)
	d.Set("display_name", resp.DisplayName)

	return diags
}

func resourceWebsiteCertificateAuthorityAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkConn()

	fleetArn, websiteCaID, err := DecodeWebsiteCertificateAuthorityAssociationResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	input := &worklink.DisassociateWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(fleetArn),
		WebsiteCaId: aws.String(websiteCaID),
	}

	if _, err := conn.DisassociateWebsiteCertificateAuthorityWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    WebsiteCertificateAuthorityAssociationStateRefresh(ctx, conn, websiteCaID, fleetArn),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WorkLink Website Certificate Authority Association (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func WebsiteCertificateAuthorityAssociationStateRefresh(ctx context.Context, conn *worklink.WorkLink, websiteCaID, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &worklink.DescribeWebsiteCertificateAuthorityOutput{}

		resp, err := conn.DescribeWebsiteCertificateAuthorityWithContext(ctx, &worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(arn),
			WebsiteCaId: aws.String(websiteCaID),
		})
		if tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
			return emptyResp, "DELETED", nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, "", nil
	}
}

func DecodeWebsiteCertificateAuthorityAssociationResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID(%s), expected WebsiteCaId/FleetArn", id)
	}
	fleetArn := parts[0]
	websiteCaID := parts[1]

	return fleetArn, websiteCaID, nil
}
