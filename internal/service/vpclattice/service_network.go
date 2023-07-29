// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service_network", name="ServiceNetwork")
// @Tags(identifierAttribute="arn")
func ResourceServiceNetwork() *schema.Resource {
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameServiceNetwork = "Service Network"
)

func resourceServiceNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	in := &vpclattice.CreateServiceNetworkInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auth_type"); ok {
		in.AuthType = types.AuthType(v.(string))
	}

	out, err := conn.CreateServiceNetwork(ctx, in)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameServiceNetwork, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameServiceNetwork, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id))

	return resourceServiceNetworkRead(ctx, d, meta)
}

func resourceServiceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := findServiceNetworkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice ServiceNetwork (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameServiceNetwork, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("auth_type", out.AuthType)
	d.Set("name", out.Name)

	return nil
}

func resourceServiceNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		in := &vpclattice.UpdateServiceNetworkInput{
			ServiceNetworkIdentifier: aws.String(d.Id()),
		}

		if d.HasChanges("auth_type") {
			in.AuthType = types.AuthType(d.Get("auth_type").(string))
		}

		log.Printf("[DEBUG] Updating VPCLattice ServiceNetwork (%s): %#v", d.Id(), in)
		_, err := conn.UpdateServiceNetwork(ctx, in)
		if err != nil {
			return create.DiagError(names.VPCLattice, create.ErrActionUpdating, ResNameServiceNetwork, d.Id(), err)
		}
	}

	return resourceServiceNetworkRead(ctx, d, meta)
}

func resourceServiceNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPC Lattice Service Network: %s", d.Id())
	_, err := conn.DeleteServiceNetwork(ctx, &vpclattice.DeleteServiceNetworkInput{
		ServiceNetworkIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.VPCLattice, create.ErrActionDeleting, ResNameServiceNetwork, d.Id(), err)
	}

	return nil
}

func findServiceNetworkByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkOutput, error) {
	in := &vpclattice.GetServiceNetworkInput{
		ServiceNetworkIdentifier: aws.String(id),
	}
	out, err := conn.GetServiceNetwork(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
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
