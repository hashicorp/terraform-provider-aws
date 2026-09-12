// Copyright IBM Corp. 2014, 2025
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
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_appstream_application_entitlement_association", name="Application Entitlement Association")
func resourceApplicationEntitlementAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationEntitlementAssociationCreate,
		ReadWithoutTimeout:   resourceApplicationEntitlementAssociationRead,
		DeleteWithoutTimeout: resourceApplicationEntitlementAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"entitlement_name": {
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

func resourceApplicationEntitlementAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stackName := d.Get("stack_name").(string)
	entitlementName := d.Get("entitlement_name").(string)
	applicationIdentifier := d.Get("application_identifier").(string)
	id := applicationEntitlementAssociationCreateResourceID(stackName, entitlementName, applicationIdentifier)

	input := appstream.AssociateApplicationToEntitlementInput{
		StackName:             aws.String(stackName),
		EntitlementName:       aws.String(entitlementName),
		ApplicationIdentifier: aws.String(applicationIdentifier),
	}

	const (
		timeout = 15 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceNotFoundException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.AssociateApplicationToEntitlement(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Application Entitlement Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceApplicationEntitlementAssociationRead(ctx, d, meta)...)
}

func resourceApplicationEntitlementAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stackName, entitlementName, applicationIdentifier, err := applicationEntitlementAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = findApplicationEntitlementAssociationByThreePartKey(ctx, conn, stackName, entitlementName, applicationIdentifier)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] AppStream Application Entitlement Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Application Entitlement Association (%s): %s", d.Id(), err)
	}

	d.Set("application_identifier", applicationIdentifier)
	d.Set("entitlement_name", entitlementName)
	d.Set("stack_name", stackName)

	return diags
}

func resourceApplicationEntitlementAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stackName, entitlementName, applicationIdentifier, err := applicationEntitlementAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting AppStream Application Entitlement Association: %s", d.Id())
	input := appstream.DisassociateApplicationFromEntitlementInput{
		StackName:             aws.String(stackName),
		EntitlementName:       aws.String(entitlementName),
		ApplicationIdentifier: aws.String(applicationIdentifier),
	}
	_, err = conn.DisassociateApplicationFromEntitlement(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Application Entitlement Association (%s): %s", d.Id(), err)
	}

	return diags
}

const applicationEntitlementAssociationResourceIDSeparator = "/"

func applicationEntitlementAssociationCreateResourceID(stackName, entitlementName, applicationIdentifier string) string {
	parts := []string{stackName, entitlementName, applicationIdentifier}
	id := strings.Join(parts, applicationEntitlementAssociationResourceIDSeparator)

	return id
}

func applicationEntitlementAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, applicationEntitlementAssociationResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected StackName%[2]sEntitlementName%[2]sApplicationIdentifier", id, applicationEntitlementAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findApplicationEntitlementAssociationByThreePartKey(ctx context.Context, conn *appstream.Client, stackName, entitlementName, applicationIdentifier string) error {
	input := appstream.ListEntitledApplicationsInput{
		StackName:       aws.String(stackName),
		EntitlementName: aws.String(entitlementName),
	}
	_, err := findEntitledApplication(ctx, conn, &input, func(v awstypes.EntitledApplication) bool {
		return aws.ToString(v.ApplicationIdentifier) == applicationIdentifier
	})

	return err
}

func findEntitledApplication(ctx context.Context, conn *appstream.Client, input *appstream.ListEntitledApplicationsInput, filter tfslices.Predicate[awstypes.EntitledApplication]) (*awstypes.EntitledApplication, error) {
	output, err := findEntitledApplications(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEntitledApplications(ctx context.Context, conn *appstream.Client, input *appstream.ListEntitledApplicationsInput, filter tfslices.Predicate[awstypes.EntitledApplication]) ([]awstypes.EntitledApplication, error) {
	var output []awstypes.EntitledApplication

	for {
		page, err := conn.ListEntitledApplications(ctx, input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.EntitledApplications {
			if filter(v) {
				output = append(output, v)
			}
		}

		if aws.ToString(page.NextToken) == "" {
			break
		}

		input.NextToken = page.NextToken
	}

	if len(output) == 0 {
		return nil, &sdkretry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}
