// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ec2_managed_prefix_list_entry")
func ResourceManagedPrefixListEntry() *schema.Resource {
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
			"description": {
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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	cidr := d.Get("cidr").(string)
	plID := d.Get("prefix_list_id").(string)
	id := ManagedPrefixListEntryCreateResourceID(plID, cidr)

	addPrefixListEntry := &ec2.AddPrefixListEntry{Cidr: aws.String(cidr)}

	if v, ok := d.GetOk("description"); ok {
		addPrefixListEntry.Description = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		mutexKey := fmt.Sprintf("vpc-managed-prefix-list-%s", plID)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		pl, err := FindManagedPrefixListByID(ctx, conn, plID)

		if err != nil {
			return nil, fmt.Errorf("reading VPC Managed Prefix List (%s): %w", plID, err)
		}

		input := &ec2.ModifyManagedPrefixListInput{
			AddEntries:     []*ec2.AddPrefixListEntry{addPrefixListEntry},
			CurrentVersion: pl.Version,
			PrefixListId:   aws.String(plID),
		}

		return conn.ModifyManagedPrefixListWithContext(ctx, input)
	}, errCodeIncorrectState, errCodePrefixListVersionMismatch)

	if err != nil {
		return diag.Errorf("creating VPC Managed Prefix List Entry (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := WaitManagedPrefixListModified(ctx, conn, plID); err != nil {
		return diag.Errorf("waiting for VPC Managed Prefix List Entry (%s) create: %s", d.Id(), err)
	}

	return resourceManagedPrefixListEntryRead(ctx, d, meta)
}

func resourceManagedPrefixListEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	plID, cidr, err := ManagedPrefixListEntryParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ManagedPrefixListEntryCreateTimeout, func() (interface{}, error) {
		return FindManagedPrefixListEntryByIDAndCIDR(ctx, conn, plID, cidr)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Managed Prefix List Entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading VPC Managed Prefix List Entry (%s): %s", d.Id(), err)
	}

	entry := outputRaw.(*ec2.PrefixListEntry)

	d.Set("cidr", entry.Cidr)
	d.Set("description", entry.Description)

	return nil
}

func resourceManagedPrefixListEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	plID, cidr, err := ManagedPrefixListEntryParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		mutexKey := fmt.Sprintf("vpc-managed-prefix-list-%s", plID)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		pl, err := FindManagedPrefixListByID(ctx, conn, plID)

		if err != nil {
			return nil, fmt.Errorf("reading VPC Managed Prefix List (%s): %w", plID, err)
		}

		input := &ec2.ModifyManagedPrefixListInput{
			CurrentVersion: pl.Version,
			PrefixListId:   aws.String(plID),
			RemoveEntries:  []*ec2.RemovePrefixListEntry{{Cidr: aws.String(cidr)}},
		}

		return conn.ModifyManagedPrefixListWithContext(ctx, input)
	}, errCodeIncorrectState, errCodePrefixListVersionMismatch)

	if err != nil {
		return diag.Errorf("deleting VPC Managed Prefix List Entry (%s): %s", d.Id(), err)
	}

	_, err = WaitManagedPrefixListModified(ctx, conn, plID)

	if err != nil {
		return diag.Errorf("waiting for VPC Managed Prefix List Entry (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func resourceManagedPrefixListEntryImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	plID, cidr, err := ManagedPrefixListEntryParseResourceID(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("cidr", cidr)
	d.Set("prefix_list_id", plID)

	return []*schema.ResourceData{d}, nil
}

const managedPrefixListEntryIDSeparator = ","

func ManagedPrefixListEntryCreateResourceID(prefixListID, cidrBlock string) string {
	parts := []string{prefixListID, cidrBlock}
	id := strings.Join(parts, managedPrefixListEntryIDSeparator)

	return id
}

func ManagedPrefixListEntryParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, managedPrefixListEntryIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected prefix-list-id%[2]scidr-block", id, managedPrefixListEntryIDSeparator)
}
