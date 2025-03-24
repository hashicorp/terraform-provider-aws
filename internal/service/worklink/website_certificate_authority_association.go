// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package worklink

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/worklink"
	awstypes "github.com/aws/aws-sdk-go-v2/service/worklink/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_worklink_website_certificate_authority_association", name="Website Certificate Authority Association")
func resourceWebsiteCertificateAuthorityAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebsiteCertificateAuthorityAssociationCreate,
		ReadWithoutTimeout:   resourceWebsiteCertificateAuthorityAssociationRead,
		DeleteWithoutTimeout: resourceWebsiteCertificateAuthorityAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCertificate: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"fleet_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"website_ca_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		DeprecationMessage: `The aws_worklink_website_certificate_authority_association resource has been deprecated and will be removed in a future version. Use Amazon WorkSpaces Secure Browser instead`,
	}
}

func resourceWebsiteCertificateAuthorityAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	input := &worklink.AssociateWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(d.Get("fleet_arn").(string)),
		Certificate: aws.String(d.Get(names.AttrCertificate).(string)),
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	resp, err := conn.AssociateWebsiteCertificateAuthority(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Website Certificate Authority Association: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", d.Get("fleet_arn").(string), aws.ToString(resp.WebsiteCaId)))

	return append(diags, resourceWebsiteCertificateAuthorityAssociationRead(ctx, d, meta)...)
}

func resourceWebsiteCertificateAuthorityAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	fleetArn, websiteCaID, err := decodeWebsiteCertificateAuthorityAssociationResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	output, err := findWebsiteCertificateAuthorityByARNAndID(ctx, conn, fleetArn, websiteCaID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkLink Website Certificate Authority Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	d.Set("website_ca_id", websiteCaID)
	d.Set("fleet_arn", fleetArn)
	d.Set(names.AttrCertificate, output.Certificate)
	d.Set(names.AttrDisplayName, output.DisplayName)

	return diags
}

func resourceWebsiteCertificateAuthorityAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	fleetArn, websiteCaID, err := decodeWebsiteCertificateAuthorityAssociationResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	input := &worklink.DisassociateWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(fleetArn),
		WebsiteCaId: aws.String(websiteCaID),
	}

	if _, err := conn.DisassociateWebsiteCertificateAuthority(ctx, input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    websiteCertificateAuthorityAssociationStateRefresh(ctx, conn, websiteCaID, fleetArn),
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

func findWebsiteCertificateAuthorityByARNAndID(ctx context.Context, conn *worklink.Client, arn, caID string) (*worklink.DescribeWebsiteCertificateAuthorityOutput, error) {
	input := &worklink.DescribeWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(arn),
		WebsiteCaId: aws.String(caID),
	}

	output, err := conn.DescribeWebsiteCertificateAuthority(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func websiteCertificateAuthorityAssociationStateRefresh(ctx context.Context, conn *worklink.Client, websiteCaID, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		emptyResp := &worklink.DescribeWebsiteCertificateAuthorityOutput{}

		input := worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(arn),
			WebsiteCaId: aws.String(websiteCaID),
		}
		resp, err := conn.DescribeWebsiteCertificateAuthority(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return emptyResp, "DELETED", nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, "", nil
	}
}

func decodeWebsiteCertificateAuthorityAssociationResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID(%s), expected WebsiteCaId/FleetArn", id)
	}
	fleetArn := parts[0]
	websiteCaID := parts[1]

	return fleetArn, websiteCaID, nil
}
