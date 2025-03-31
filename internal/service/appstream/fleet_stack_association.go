// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_appstream_fleet_stack_association", name="Fleet Stack Association")
func resourceFleetStackAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFleetStackAssociationCreate,
		ReadWithoutTimeout:   resourceFleetStackAssociationRead,
		DeleteWithoutTimeout: resourceFleetStackAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"fleet_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stack_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFleetStackAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	fleetName, stackName := d.Get("fleet_name").(string), d.Get("stack_name").(string)
	id := fleetStackAssociationCreateResourceID(fleetName, stackName)
	input := appstream.AssociateFleetInput{
		FleetName: aws.String(fleetName),
		StackName: aws.String(stackName),
	}

	const (
		timeout = 15 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.ResourceNotFoundException](ctx, timeout, func() (any, error) {
		return conn.AssociateFleet(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Fleet Stack Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceFleetStackAssociationRead(ctx, d, meta)...)
}

func resourceFleetStackAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	fleetName, stackName, err := fleetStackAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = findFleetStackAssociationByTwoPartKey(ctx, conn, fleetName, stackName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream Fleet Stack Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Fleet Stack Association (%s): %s", d.Id(), err)
	}

	d.Set("fleet_name", fleetName)
	d.Set("stack_name", stackName)

	return diags
}

func resourceFleetStackAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	fleetName, stackName, err := fleetStackAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting AppStream Fleet Stack Association: %s", d.Id())
	input := appstream.DisassociateFleetInput{
		StackName: aws.String(stackName),
		FleetName: aws.String(fleetName),
	}
	_, err = conn.DisassociateFleet(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Fleet Stack Association (%s): %s", d.Id(), err)
	}

	return diags
}

const fleetStackAssociationResourceIDSeparator = "/"

func fleetStackAssociationCreateResourceID(fleetName, stackName string) string {
	parts := []string{fleetName, stackName}
	id := strings.Join(parts, fleetStackAssociationResourceIDSeparator)

	return id
}

func fleetStackAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, fleetStackAssociationResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected FleetName%[2]sStackName", id, fleetStackAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findFleetStackAssociationByTwoPartKey(ctx context.Context, conn *appstream.Client, fleetName, stackName string) error {
	input := appstream.ListAssociatedStacksInput{
		FleetName: aws.String(fleetName),
	}
	_, err := findAssociatedStack(ctx, conn, &input, func(v string) bool {
		return v == stackName
	})

	return err
}

func findAssociatedStack(ctx context.Context, conn *appstream.Client, input *appstream.ListAssociatedStacksInput, filter tfslices.Predicate[string]) (*string, error) {
	output, err := findAssociatedStacks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAssociatedStacks(ctx context.Context, conn *appstream.Client, input *appstream.ListAssociatedStacksInput, filter tfslices.Predicate[string]) ([]string, error) {
	var output []string

	err := listAssociatedStacksPages(ctx, conn, input, func(page *appstream.ListAssociatedStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Names {
			if filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
