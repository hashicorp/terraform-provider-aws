package aws

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

var (
	awsPrefixListEntrySetHashFunc = schema.HashResource(prefixListEntrySchema())
)

func resourceAwsPrefixList() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPrefixListCreate,
		Read:   resourceAwsPrefixListRead,
		Update: resourceAwsPrefixListUpdate,
		Delete: resourceAwsPrefixListDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem:       prefixListEntrySchema(),
				Set:        awsPrefixListEntrySetHashFunc,
			},
			"max_entries": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
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
			"tags": tagsSchema(),
		},
	}
}

func prefixListEntrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cidr_block": {
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
	}
}

func resourceAwsPrefixListCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := ec2.CreateManagedPrefixListInput{}

	input.AddressFamily = aws.String(d.Get("address_family").(string))

	if v, ok := d.GetOk("entry"); ok {
		input.Entries = expandAddPrefixListEntries(v)
	}

	input.MaxEntries = aws.Int64(int64(d.Get("max_entries").(int)))
	input.PrefixListName = aws.String(d.Get("name").(string))

	if v, ok := d.GetOk("tags"); ok {
		input.TagSpecifications = ec2TagSpecificationsFromMap(
			v.(map[string]interface{}),
			"prefix-list") // no ec2.ResourceTypePrefixList as of 01/07/20
	}

	output, err := conn.CreateManagedPrefixList(&input)
	if err != nil {
		return fmt.Errorf("failed to create managed prefix list: %v", err)
	}

	id := aws.StringValue(output.PrefixList.PrefixListId)

	log.Printf("[INFO] Created Managed Prefix List %s (%s)", d.Get("name").(string), id)

	if err := waitUntilAwsManagedPrefixListSettled(id, conn, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("prefix list %s did not settle after create: %s", id, err)
	}

	d.SetId(id)

	return resourceAwsPrefixListRead(d, meta)
}

func resourceAwsPrefixListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	id := d.Id()

	pl, ok, err := getManagedPrefixList(id, conn)
	switch {
	case err != nil:
		return err
	case !ok:
		log.Printf("[WARN] Managed Prefix List %s not found; removing from state.", id)
		d.SetId("")
		return nil
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)

	entries, err := getPrefixListEntries(id, conn, 0)
	if err != nil {
		return err
	}

	if err := d.Set("entry", flattenPrefixListEntries(entries)); err != nil {
		return fmt.Errorf("error setting attribute entry of managed prefix list %s: %s", id, err)
	}

	d.Set("max_entries", pl.MaxEntries)
	d.Set("name", pl.PrefixListName)
	d.Set("owner_id", pl.OwnerId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(pl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error settings attribute tags of managed prefix list %s: %s", id, err)
	}

	return nil
}

func resourceAwsPrefixListUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	modifyPrefixList := false

	input := ec2.ModifyManagedPrefixListInput{}

	input.PrefixListId = aws.String(id)

	if d.HasChange("name") {
		input.PrefixListName = aws.String(d.Get("name").(string))
		modifyPrefixList = true
	}

	if d.HasChange("entry") {
		pl, ok, err := getManagedPrefixList(id, conn)
		switch {
		case err != nil:
			return err
		case !ok:
			return &resource.NotFoundError{}
		}

		currentVersion := aws.Int64Value(pl.Version)

		oldEntries, err := getPrefixListEntries(id, conn, currentVersion)
		if err != nil {
			return err
		}

		newEntries := expandAddPrefixListEntries(d.Get("entry"))
		adds, removes := computePrefixListEntriesModification(oldEntries, newEntries)

		if len(adds) > 0 || len(removes) > 0 {
			if len(adds) > 0 {
				// the Modify API doesn't like empty lists
				input.AddEntries = adds
			}

			if len(removes) > 0 {
				// the Modify API doesn't like empty lists
				input.RemoveEntries = removes
			}

			input.CurrentVersion = aws.Int64(currentVersion)
			modifyPrefixList = true
		}
	}

	if modifyPrefixList {
		log.Printf("[INFO] modifying managed prefix list %s...", id)

		switch _, err := conn.ModifyManagedPrefixList(&input); {
		case isAWSErr(err, "PrefixListVersionMismatch", "prefix list has the incorrect version number"):
			return fmt.Errorf("failed to modify managed prefix list %s: conflicting change", id)
		case err != nil:
			return fmt.Errorf("failed to modify managed prefix list %s: %s", id, err)
		}

		if err := waitUntilAwsManagedPrefixListSettled(id, conn, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("prefix list did not settle after update: %s", err)
		}
	}

	if d.HasChange("tags") {
		before, after := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, id, before, after); err != nil {
			return fmt.Errorf("failed to update tags of managed prefix list %s: %s", id, err)
		}
	}

	return resourceAwsPrefixListRead(d, meta)
}

func resourceAwsPrefixListDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	input := ec2.DeleteManagedPrefixListInput{
		PrefixListId: aws.String(id),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteManagedPrefixList(&input)
		switch {
		case isManagedPrefixListModificationConflictErr(err):
			return resource.RetryableError(err)
		case isAWSErr(err, "InvalidPrefixListID.NotFound", ""):
			log.Printf("[WARN] managed prefix list %s has already been deleted", id)
			return nil
		case err != nil:
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteManagedPrefixList(&input)
	}

	if err != nil {
		return fmt.Errorf("failed to delete managed prefix list %s: %s", id, err)
	}

	if err := waitUntilAwsManagedPrefixListSettled(id, conn, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("prefix list %s did not settle after delete: %s", id, err)
	}

	return nil
}

func expandAddPrefixListEntries(input interface{}) []*ec2.AddPrefixListEntry {
	if input == nil {
		return nil
	}

	list := input.(*schema.Set).List()
	result := make([]*ec2.AddPrefixListEntry, 0, len(list))

	for _, entry := range list {
		m := entry.(map[string]interface{})

		output := ec2.AddPrefixListEntry{}

		output.Cidr = aws.String(m["cidr_block"].(string))

		if v, ok := m["description"]; ok {
			output.Description = aws.String(v.(string))
		}

		result = append(result, &output)
	}

	return result
}

func flattenPrefixListEntries(entries []*ec2.PrefixListEntry) *schema.Set {
	list := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		m := make(map[string]interface{}, 2)
		m["cidr_block"] = aws.StringValue(entry.Cidr)

		if entry.Description != nil {
			m["description"] = aws.StringValue(entry.Description)
		}

		list = append(list, m)
	}

	return schema.NewSet(awsPrefixListEntrySetHashFunc, list)
}

func getManagedPrefixList(
	id string,
	conn *ec2.EC2,
) (*ec2.ManagedPrefixList, bool, error) {
	input := ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeManagedPrefixLists(&input)
	switch {
	case isAWSErr(err, "InvalidPrefixListID.NotFound", ""):
		return nil, false, nil
	case err != nil:
		return nil, false, fmt.Errorf("describe managed prefix list %s: %v", id, err)
	case len(output.PrefixLists) != 1:
		return nil, false, nil
	}

	return output.PrefixLists[0], true, nil
}

func getPrefixListEntries(
	id string,
	conn *ec2.EC2,
	version int64,
) ([]*ec2.PrefixListEntry, error) {
	input := ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	if version > 0 {
		input.TargetVersion = aws.Int64(version)
	}

	result := []*ec2.PrefixListEntry(nil)
	switch err := conn.GetManagedPrefixListEntriesPages(
		&input,
		func(output *ec2.GetManagedPrefixListEntriesOutput, last bool) bool {
			result = append(result, output.Entries...)
			return true
		}); {
	case err != nil:
		return nil, fmt.Errorf("failed to get entries in prefix list %s: %v", id, err)
	}

	return result, nil
}

func computePrefixListEntriesModification(
	oldEntries []*ec2.PrefixListEntry,
	newEntries []*ec2.AddPrefixListEntry,
) ([]*ec2.AddPrefixListEntry, []*ec2.RemovePrefixListEntry) {
	adds := map[string]string{} // CIDR => Description

	removes := map[string]struct{}{} // set of CIDR
	for _, oldEntry := range oldEntries {
		oldCIDR := aws.StringValue(oldEntry.Cidr)
		removes[oldCIDR] = struct{}{}
	}

	for _, newEntry := range newEntries {
		newCIDR := aws.StringValue(newEntry.Cidr)
		newDescription := aws.StringValue(newEntry.Description)

		for _, oldEntry := range oldEntries {
			oldCIDR := aws.StringValue(oldEntry.Cidr)
			oldDescription := aws.StringValue(oldEntry.Description)

			if oldCIDR == newCIDR {
				delete(removes, oldCIDR)

				if oldDescription != newDescription {
					adds[oldCIDR] = newDescription
				}

				goto nextNewEntry
			}
		}

		// reach this point when no matching oldEntry found
		adds[newCIDR] = newDescription

	nextNewEntry:
	}

	addList := make([]*ec2.AddPrefixListEntry, 0, len(adds))
	for cidr, description := range adds {
		addList = append(addList, &ec2.AddPrefixListEntry{
			Cidr:        aws.String(cidr),
			Description: aws.String(description),
		})
	}
	sort.Slice(addList, func(i, j int) bool {
		return aws.StringValue(addList[i].Cidr) < aws.StringValue(addList[j].Cidr)
	})

	removeList := make([]*ec2.RemovePrefixListEntry, 0, len(removes))
	for cidr := range removes {
		removeList = append(removeList, &ec2.RemovePrefixListEntry{
			Cidr: aws.String(cidr),
		})
	}
	sort.Slice(removeList, func(i, j int) bool {
		return aws.StringValue(removeList[i].Cidr) < aws.StringValue(removeList[j].Cidr)
	})

	return addList, removeList
}

func waitUntilAwsManagedPrefixListSettled(
	id string,
	conn *ec2.EC2,
	timeout time.Duration,
) error {
	log.Printf("[INFO] Waiting for managed prefix list %s to settle...", id)

	err := resource.Retry(timeout, func() *resource.RetryError {
		settled, err := isAwsManagedPrefixListSettled(id, conn)
		switch {
		case err != nil:
			return resource.NonRetryableError(err)
		case !settled:
			return resource.RetryableError(errors.New("resource not yet settled"))
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		return fmt.Errorf("timed out: %s", err)
	}

	return nil
}

func isAwsManagedPrefixListSettled(id string, conn *ec2.EC2) (bool, error) {
	pl, ok, err := getManagedPrefixList(id, conn)
	switch {
	case err != nil:
		return false, err
	case !ok:
		return true, nil
	}

	switch state := aws.StringValue(pl.State); state {
	case ec2.PrefixListStateCreateComplete, ec2.PrefixListStateModifyComplete, ec2.PrefixListStateDeleteComplete:
		return true, nil
	case ec2.PrefixListStateCreateInProgress, ec2.PrefixListStateModifyInProgress, ec2.PrefixListStateDeleteInProgress:
		return false, nil
	case ec2.PrefixListStateCreateFailed, ec2.PrefixListStateModifyFailed, ec2.PrefixListStateDeleteFailed:
		return false, fmt.Errorf("terminal state %s indicates failure", state)
	default:
		return false, fmt.Errorf("unexpected state %s", state)
	}
}
