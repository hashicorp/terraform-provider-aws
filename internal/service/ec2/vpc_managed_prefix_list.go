// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_managed_prefix_list", name="Managed Prefix List")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceManagedPrefixListCreate,
		ReadWithoutTimeout:   resourceManagedPrefixListRead,
		UpdateWithoutTimeout: resourceManagedPrefixListUpdate,
		DeleteWithoutTimeout: resourceManagedPrefixListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf(names.AttrVersion, func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("entry")
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(managedPrefixListAddressFamily_Values(), false),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"entry": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						names.AttrDescription: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
					},
				},
			},
			"max_entries": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceManagedPrefixListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ec2.CreateManagedPrefixListInput{
		AddressFamily:     aws.String(d.Get("address_family").(string)),
		ClientToken:       aws.String(id.UniqueId()),
		MaxEntries:        aws.Int32(int32(d.Get("max_entries").(int))),
		PrefixListName:    aws.String(name),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypePrefixList),
	}

	if v, ok := d.GetOk("entry"); ok && v.(*schema.Set).Len() > 0 {
		input.Entries = expandAddPrefixListEntries(v.(*schema.Set).List())
	}

	output, err := conn.CreateManagedPrefixList(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Managed Prefix List (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PrefixList.PrefixListId))

	if _, err := waitManagedPrefixListCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Managed Prefix List (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceManagedPrefixListRead(ctx, d, meta)...)
}

func resourceManagedPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	pl, err := findManagedPrefixListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Managed Prefix List %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Managed Prefix List (%s): %s", d.Id(), err)
	}

	prefixListEntries, err := findManagedPrefixListEntriesByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Managed Prefix List (%s) Entries: %s", d.Id(), err)
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set(names.AttrARN, pl.PrefixListArn)
	if err := d.Set("entry", flattenPrefixListEntries(prefixListEntries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting entry: %s", err)
	}
	d.Set("max_entries", pl.MaxEntries)
	d.Set(names.AttrName, pl.PrefixListName)
	d.Set(names.AttrOwnerID, pl.OwnerId)
	d.Set(names.AttrVersion, pl.Version)

	setTagsOut(ctx, pl.Tags)

	return diags
}

func resourceManagedPrefixListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// MaxEntries & Entry cannot change in the same API call.
	//   If MaxEntry is increasing, complete before updating entry(s)
	//   If MaxEntry is decreasing, complete after updating entry(s)
	maxEntryChangedDecrease := false
	var newMaxEntryInt int32

	if d.HasChange("max_entries") {
		oldMaxEntry, newMaxEntry := d.GetChange("max_entries")
		newMaxEntryInt = int32(d.Get("max_entries").(int))

		if newMaxEntry.(int) < oldMaxEntry.(int) {
			maxEntryChangedDecrease = true
		} else {
			err := updateMaxEntry(ctx, conn, d.Id(), newMaxEntryInt)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Managed Prefix List (%s) increased MaxEntries : %s", d.Id(), err)
			}
		}
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "max_entries") {
		input := &ec2.ModifyManagedPrefixListInput{
			PrefixListId: aws.String(d.Id()),
		}

		input.PrefixListName = aws.String(d.Get(names.AttrName).(string))
		currentVersion := int64(d.Get(names.AttrVersion).(int))
		wait := false

		oldAttr, newAttr := d.GetChange("entry")
		os := oldAttr.(*schema.Set)
		ns := newAttr.(*schema.Set)

		if addEntries := ns.Difference(os); addEntries.Len() > 0 {
			input.AddEntries = expandAddPrefixListEntries(addEntries.List())
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		if removeEntries := os.Difference(ns); removeEntries.Len() > 0 {
			input.RemoveEntries = expandRemovePrefixListEntries(removeEntries.List())
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		// Prevent the following error on description-only updates:
		//   InvalidParameterValue: Request cannot contain Cidr #.#.#.#/# in both AddPrefixListEntries and RemovePrefixListEntries
		// Attempting to just delete the RemoveEntries item causes:
		//   InvalidRequest: The request received was invalid.
		// Therefore it seems we must issue two ModifyManagedPrefixList calls,
		// one with a collection of all description-only removals and the
		// second one will add them all back.
		if len(input.AddEntries) > 0 && len(input.RemoveEntries) > 0 {
			descriptionOnlyRemovals := []awstypes.RemovePrefixListEntry{}
			removals := []awstypes.RemovePrefixListEntry{}

			for _, removeEntry := range input.RemoveEntries {
				inAddAndRemove := false

				for _, addEntry := range input.AddEntries {
					if aws.ToString(addEntry.Cidr) == aws.ToString(removeEntry.Cidr) {
						inAddAndRemove = true
						break
					}
				}

				if inAddAndRemove {
					descriptionOnlyRemovals = append(descriptionOnlyRemovals, removeEntry)
				} else {
					removals = append(removals, removeEntry)
				}
			}

			if len(descriptionOnlyRemovals) > 0 {
				_, err := conn.ModifyManagedPrefixList(ctx, &ec2.ModifyManagedPrefixListInput{
					CurrentVersion: input.CurrentVersion,
					PrefixListId:   aws.String(d.Id()),
					RemoveEntries:  descriptionOnlyRemovals,
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating EC2 Managed Prefix List (%s): %s", d.Id(), err)
				}

				managedPrefixList, err := waitManagedPrefixListModified(ctx, conn, d.Id())

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for EC2 Managed Prefix List (%s) update: %s", d.Id(), err)
				}

				input.CurrentVersion = managedPrefixList.Version
			}

			if len(removals) > 0 {
				input.RemoveEntries = removals
			} else {
				// Prevent this error if RemoveEntries is list with no elements after removals:
				//   InvalidRequest: The request received was invalid.
				input.RemoveEntries = nil
			}
		}

		_, err := conn.ModifyManagedPrefixList(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Managed Prefix List (%s): %s", d.Id(), err)
		}

		if wait {
			if _, err := waitManagedPrefixListModified(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EC2 Managed Prefix List (%s) update: %s", d.Id(), err)
			}
		}
	}

	// Only decrease MaxEntries after entry(s) have had opportunity to be removed
	if maxEntryChangedDecrease {
		err := updateMaxEntry(ctx, conn, d.Id(), newMaxEntryInt)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Managed Prefix List (%s) decreased MaxEntries : %s", d.Id(), err)
		}
	}

	return append(diags, resourceManagedPrefixListRead(ctx, d, meta)...)
}

func resourceManagedPrefixListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Managed Prefix List: %s", d.Id())
	_, err := conn.DeleteManagedPrefixList(ctx, &ec2.DeleteManagedPrefixListInput{
		PrefixListId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Managed Prefix List (%s): %s", d.Id(), err)
	}

	if _, err := waitManagedPrefixListDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Managed Prefix List (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func updateMaxEntry(ctx context.Context, conn *ec2.Client, id string, maxEntries int32) error {
	_, err := conn.ModifyManagedPrefixList(ctx, &ec2.ModifyManagedPrefixListInput{
		PrefixListId: aws.String(id),
		MaxEntries:   aws.Int32(maxEntries),
	})

	if err != nil {
		return fmt.Errorf("updating MaxEntries for EC2 Managed Prefix List (%s): %s", id, err)
	}

	_, err = waitManagedPrefixListModified(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("waiting for EC2 Managed Prefix List (%s) MaxEntries update: %s", id, err)
	}

	return nil
}

func expandAddPrefixListEntry(tfMap map[string]interface{}) awstypes.AddPrefixListEntry {
	apiObject := awstypes.AddPrefixListEntry{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	return apiObject
}

func expandAddPrefixListEntries(tfList []interface{}) []awstypes.AddPrefixListEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.AddPrefixListEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandAddPrefixListEntry(tfMap))
	}

	return apiObjects
}

func expandRemovePrefixListEntry(tfMap map[string]interface{}) awstypes.RemovePrefixListEntry {
	apiObject := awstypes.RemovePrefixListEntry{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	return apiObject
}

func expandRemovePrefixListEntries(tfList []interface{}) []awstypes.RemovePrefixListEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.RemovePrefixListEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandRemovePrefixListEntry(tfMap))
	}

	return apiObjects
}

func flattenPrefixListEntry(apiObject awstypes.PrefixListEntry) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Cidr; v != nil {
		tfMap["cidr"] = aws.ToString(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	return tfMap
}

func flattenPrefixListEntries(apiObjects []awstypes.PrefixListEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPrefixListEntry(apiObject))
	}

	return tfList
}
