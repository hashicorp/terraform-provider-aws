// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service_network", name="Service Network")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceServiceNetwork() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceNetworkCreate,
		ReadWithoutTimeout:   resourceServiceNetworkRead,
		UpdateWithoutTimeout: resourceServiceNetworkUpdate,
		DeleteWithoutTimeout: resourceServiceNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.AuthType](),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceServiceNetworkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := vpclattice.CreateServiceNetworkInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auth_type"); ok {
		input.AuthType = types.AuthType(v.(string))
	}

	output, err := conn.CreateServiceNetwork(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Service Network (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceServiceNetworkRead(ctx, d, meta)...)
}

func resourceServiceNetworkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	output, err := findServiceNetworkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Service Network (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("auth_type", output.AuthType)
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceServiceNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := vpclattice.UpdateServiceNetworkInput{
			ServiceNetworkIdentifier: aws.String(d.Id()),
		}

		if d.HasChanges("auth_type") {
			input.AuthType = types.AuthType(d.Get("auth_type").(string))
		}

		_, err := conn.UpdateServiceNetwork(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating VPCLattice Service Network (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceNetworkRead(ctx, d, meta)...)
}

func resourceServiceNetworkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPC Lattice Service Network: %s", d.Id())
	input := vpclattice.DeleteServiceNetworkInput{
		ServiceNetworkIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteServiceNetwork(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Service Network (%s): %s", d.Id(), err)
	}

	return diags
}

func findServiceNetworkByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkOutput, error) {
	input := vpclattice.GetServiceNetworkInput{
		ServiceNetworkIdentifier: aws.String(id),
	}

	return findServiceNetwork(ctx, conn, &input)
}

func findServiceNetwork(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetServiceNetworkInput) (*vpclattice.GetServiceNetworkOutput, error) {
	output, err := conn.GetServiceNetwork(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

// idFromIDOrARN return a resource ID from an ID or ARN.
func idFromIDOrARN(idOrARN string) string {
	// e.g. "sn-1234567890abcdefg" or
	// "arn:aws:vpc-lattice:us-east-1:123456789012:servicenetwork/sn-1234567890abcdefg".
	return idOrARN[strings.LastIndex(idOrARN, "/")+1:]
}

// suppressEquivalentIDOrARN provides custom difference suppression
// for strings that represent equal resource IDs or ARNs.
func suppressEquivalentIDOrARN(_, old, new string, _ *schema.ResourceData) bool {
	return idFromIDOrARN(old) == idFromIDOrARN(new)
}
