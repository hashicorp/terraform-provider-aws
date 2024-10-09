// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_managed_prefix_list_entry", name="Managed Prefix List Entry")
func resourceManagedPrefixListEntry() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceManagedPrefixListEntryCreate,
		ReadWithoutTimeout:   resourceManagedPrefixListEntryRead,
		DeleteWithoutTimeout: resourceManagedPrefixListEntryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceManagedPrefixListEntryImport,
		},

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceManagedPrefixListEntryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	cidr := d.Get("cidr").(string)
	plID := d.Get("prefix_list_id").(string)
	id := managedPrefixListEntryCreateResourceID(plID, cidr)

	addPrefixListEntry := awstypes.AddPrefixListEntry{Cidr: aws.String(cidr)}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		addPrefixListEntry.Description = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		mutexKey := fmt.Sprintf("vpc-managed-prefix-list-%s", plID)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		pl, err := findManagedPrefixListByID(ctx, conn, plID)

		if err != nil {
			return nil, fmt.Errorf("reading VPC Managed Prefix List (%s): %w", plID, err)
		}

		input := &ec2.ModifyManagedPrefixListInput{
			AddEntries:     []awstypes.AddPrefixListEntry{addPrefixListEntry},
			CurrentVersion: pl.Version,
			PrefixListId:   aws.String(plID),
		}

		return conn.ModifyManagedPrefixList(ctx, input)
	}, errCodeIncorrectState, errCodePrefixListVersionMismatch)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPC Managed Prefix List Entry (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitManagedPrefixListModified(ctx, conn, plID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Managed Prefix List Entry (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceManagedPrefixListEntryRead(ctx, d, meta)...)
}

func resourceManagedPrefixListEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	plID, cidr, err := managedPrefixListEntryParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, managedPrefixListEntryCreateTimeout, func() (interface{}, error) {
		return findManagedPrefixListEntryByIDAndCIDR(ctx, conn, plID, cidr)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Managed Prefix List Entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Managed Prefix List Entry (%s): %s", d.Id(), err)
	}

	entry := outputRaw.(*awstypes.PrefixListEntry)

	d.Set("cidr", entry.Cidr)
	d.Set(names.AttrDescription, entry.Description)

	return diags
}

func resourceManagedPrefixListEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	plID, cidr, err := managedPrefixListEntryParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		mutexKey := fmt.Sprintf("vpc-managed-prefix-list-%s", plID)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		pl, err := findManagedPrefixListByID(ctx, conn, plID)

		if err != nil {
			return nil, fmt.Errorf("reading VPC Managed Prefix List (%s): %w", plID, err)
		}

		input := &ec2.ModifyManagedPrefixListInput{
			CurrentVersion: pl.Version,
			PrefixListId:   aws.String(plID),
			RemoveEntries:  []awstypes.RemovePrefixListEntry{{Cidr: aws.String(cidr)}},
		}

		return conn.ModifyManagedPrefixList(ctx, input)
	}, errCodeIncorrectState, errCodePrefixListVersionMismatch)

	if tfawserr.ErrMessageContains(err, errCodeInvalidPrefixListModification, "does not exist.") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPC Managed Prefix List Entry (%s): %s", d.Id(), err)
	}

	_, err = waitManagedPrefixListModified(ctx, conn, plID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Managed Prefix List Entry (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceManagedPrefixListEntryImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	plID, cidr, err := managedPrefixListEntryParseResourceID(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("cidr", cidr)
	d.Set("prefix_list_id", plID)

	return []*schema.ResourceData{d}, nil
}

const managedPrefixListEntryIDSeparator = ","

func managedPrefixListEntryCreateResourceID(prefixListID, cidrBlock string) string {
	parts := []string{prefixListID, cidrBlock}
	id := strings.Join(parts, managedPrefixListEntryIDSeparator)

	return id
}

func managedPrefixListEntryParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, managedPrefixListEntryIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected prefix-list-id%[2]scidr-block", id, managedPrefixListEntryIDSeparator)
}
