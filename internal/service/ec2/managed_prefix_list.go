package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		Create: resourceManagedPrefixListCreate,
		Read:   resourceManagedPrefixListRead,
		Update: resourceManagedPrefixListUpdate,
		Delete: resourceManagedPrefixListDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("version", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("entry")
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"IPv4", "IPv6"},
					false),
			},
			"arn": {
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
						"description": {
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceManagedPrefixListCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateManagedPrefixListInput{}

	if v, ok := d.GetOk("address_family"); ok {
		input.AddressFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk("entry"); ok && v.(*schema.Set).Len() > 0 {
		input.Entries = expandEc2AddPrefixListEntries(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("max_entries"); ok {
		input.MaxEntries = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("name"); ok {
		input.PrefixListName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.TagSpecifications = ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypePrefixList)
	}

	log.Printf("[DEBUG] Creating EC2 Managed Prefix List: %s", input)
	output, err := conn.CreateManagedPrefixList(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Managed Prefix List: %w", err)
	}

	d.SetId(aws.StringValue(output.PrefixList.PrefixListId))

	if _, err := WaitManagedPrefixListCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) create: %w", d.Id(), err)
	}

	return resourceManagedPrefixListRead(d, meta)
}

func resourceManagedPrefixListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	pl, err := FindManagedPrefixListByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Managed Prefix List %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	prefixListEntries, err := FindManagedPrefixListEntriesByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 Managed Prefix List (%s) Entries: %w", d.Id(), err)
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)

	if err := d.Set("entry", flattenEc2PrefixListEntries(prefixListEntries)); err != nil {
		return fmt.Errorf("error setting entry: %w", err)
	}

	d.Set("max_entries", pl.MaxEntries)
	d.Set("name", pl.PrefixListName)
	d.Set("owner_id", pl.OwnerId)

	tags := KeyValueTags(pl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("version", pl.Version)

	return nil
}

func resourceManagedPrefixListUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyManagedPrefixListInput{
			PrefixListId: aws.String(d.Id()),
		}

		input.PrefixListName = aws.String(d.Get("name").(string))
		currentVersion := int64(d.Get("version").(int))
		wait := false

		oldAttr, newAttr := d.GetChange("entry")
		os := oldAttr.(*schema.Set)
		ns := newAttr.(*schema.Set)

		if addEntries := ns.Difference(os); addEntries.Len() > 0 {
			input.AddEntries = expandEc2AddPrefixListEntries(addEntries.List())
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		if removeEntries := os.Difference(ns); removeEntries.Len() > 0 {
			input.RemoveEntries = expandEc2RemovePrefixListEntries(removeEntries.List())
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
			descriptionOnlyRemovals := []*ec2.RemovePrefixListEntry{}
			removals := []*ec2.RemovePrefixListEntry{}

			for _, removeEntry := range input.RemoveEntries {
				inAddAndRemove := false

				for _, addEntry := range input.AddEntries {
					if aws.StringValue(addEntry.Cidr) == aws.StringValue(removeEntry.Cidr) {
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
				_, err := conn.ModifyManagedPrefixList(&ec2.ModifyManagedPrefixListInput{
					CurrentVersion: input.CurrentVersion,
					PrefixListId:   aws.String(d.Id()),
					RemoveEntries:  descriptionOnlyRemovals,
				})

				if err != nil {
					return fmt.Errorf("error updating EC2 Managed Prefix List (%s): %w", d.Id(), err)
				}

				managedPrefixList, err := WaitManagedPrefixListModified(conn, d.Id())

				if err != nil {
					return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) update: %w", d.Id(), err)
				}

				if managedPrefixList == nil {
					return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) update: empty response", d.Id())
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

		if d.HasChange("max_entries") {
			input.MaxEntries = aws.Int64(int64(d.Get("max_entries").(int)))
			wait = true
		}

		_, err := conn.ModifyManagedPrefixList(input)

		if err != nil {
			return fmt.Errorf("error updating EC2 Managed Prefix List (%s): %w", d.Id(), err)
		}

		if wait {
			if _, err := WaitManagedPrefixListModified(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) update: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Managed Prefix List (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceManagedPrefixListRead(d, meta)
}

func resourceManagedPrefixListDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Managed Prefix List: %s", d.Id())
	_, err := conn.DeleteManagedPrefixList(&ec2.DeleteManagedPrefixListInput{
		PrefixListId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPrefixListIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	if _, err := WaitManagedPrefixListDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandEc2AddPrefixListEntry(tfMap map[string]interface{}) *ec2.AddPrefixListEntry {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.AddPrefixListEntry{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	return apiObject
}

func expandEc2AddPrefixListEntries(tfList []interface{}) []*ec2.AddPrefixListEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.AddPrefixListEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2AddPrefixListEntry(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandEc2RemovePrefixListEntry(tfMap map[string]interface{}) *ec2.RemovePrefixListEntry {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.RemovePrefixListEntry{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	return apiObject
}

func expandEc2RemovePrefixListEntries(tfList []interface{}) []*ec2.RemovePrefixListEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.RemovePrefixListEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2RemovePrefixListEntry(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEc2PrefixListEntry(apiObject *ec2.PrefixListEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Cidr; v != nil {
		tfMap["cidr"] = aws.StringValue(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2PrefixListEntries(apiObjects []*ec2.PrefixListEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEc2PrefixListEntry(apiObject))
	}

	return tfList
}
