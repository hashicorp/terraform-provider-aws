package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2ManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ManagedPrefixListCreate,
		Read:   resourceAwsEc2ManagedPrefixListRead,
		Update: resourceAwsEc2ManagedPrefixListUpdate,
		Delete: resourceAwsEc2ManagedPrefixListDelete,

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
				Type:     schema.TypeSet,
				Optional: true,
				// Computed:   true,
				// ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
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
				},
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
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ManagedPrefixListCreate(d *schema.ResourceData, meta interface{}) error {
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
		return fmt.Errorf("failed to create managed prefix list: %w", err)
	}

	d.SetId(aws.StringValue(output.PrefixList.PrefixListId))

	log.Printf("[INFO] Created Managed Prefix List %s (%s)", d.Get("name").(string), d.Id())

	if _, err := waiter.ManagedPrefixListCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("managed prefix list %s failed to create: %w", d.Id(), err)
	}

	return resourceAwsEc2ManagedPrefixListRead(d, meta)
}

func resourceAwsEc2ManagedPrefixListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	id := d.Id()

	pl, ok, err := getManagedPrefixList(id, conn)
	if err != nil {
		return fmt.Errorf("failed to get managed prefix list %s: %w", id, err)
	}

	if !ok {
		log.Printf("[WARN] Managed Prefix List %s not found; removing from state.", id)
		d.SetId("")
		return nil
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)

	entries, err := getPrefixListEntries(id, conn, 0)
	if err != nil {
		return fmt.Errorf("error listing entries of EC2 Managed Prefix List (%s): %w", d.Id(), err)
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

	d.Set("version", pl.Version)

	return nil
}

func resourceAwsEc2ManagedPrefixListUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	input := ec2.ModifyManagedPrefixListInput{}

	input.PrefixListId = aws.String(id)

	if d.HasChangeExcept("tags") {
		input.PrefixListName = aws.String(d.Get("name").(string))
		currentVersion := int64(d.Get("version").(int))
		wait := false

		oldAttr, newAttr := d.GetChange("entry")
		os := oldAttr.(*schema.Set)
		ns := newAttr.(*schema.Set)

		if addEntries := ns.Difference(os); addEntries.Len() > 0 {
			input.AddEntries = expandAddPrefixListEntries(addEntries)
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		if removeEntries := os.Difference(ns); removeEntries.Len() > 0 {
			input.RemoveEntries = expandRemovePrefixListEntries(removeEntries)
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		log.Printf("[INFO] modifying managed prefix list %s...", id)

		_, err := conn.ModifyManagedPrefixList(&input)

		if isAWSErr(err, "PrefixListVersionMismatch", "prefix list has the incorrect version number") {
			return fmt.Errorf("failed to modify managed prefix list %s: conflicting change", id)
		}

		if err != nil {
			return fmt.Errorf("failed to modify managed prefix list %s: %s", id, err)
		}

		if wait {
			if _, err := waiter.ManagedPrefixListModified(conn, d.Id()); err != nil {
				return fmt.Errorf("failed to modify managed prefix list %s: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		before, after := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, id, before, after); err != nil {
			return fmt.Errorf("failed to update tags of managed prefix list %s: %s", id, err)
		}
	}

	return resourceAwsEc2ManagedPrefixListRead(d, meta)
}

func resourceAwsEc2ManagedPrefixListDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	input := ec2.DeleteManagedPrefixListInput{
		PrefixListId: aws.String(id),
	}

	_, err := conn.DeleteManagedPrefixList(&input)

	if tfawserr.ErrCodeEquals(err, "InvalidPrefixListID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	if err := waiter.ManagedPrefixListDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("failed to delete managed prefix list %s: %w", d.Id(), err)
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

func expandRemovePrefixListEntries(input interface{}) []*ec2.RemovePrefixListEntry {
	if input == nil {
		return nil
	}

	list := input.(*schema.Set).List()
	result := make([]*ec2.RemovePrefixListEntry, 0, len(list))

	for _, entry := range list {
		m := entry.(map[string]interface{})
		output := ec2.RemovePrefixListEntry{}
		output.Cidr = aws.String(m["cidr_block"].(string))
		result = append(result, &output)
	}

	return result
}

func flattenPrefixListEntries(entries []*ec2.PrefixListEntry) []interface{} {
	list := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		m := make(map[string]interface{}, 2)
		m["cidr_block"] = aws.StringValue(entry.Cidr)

		if entry.Description != nil {
			m["description"] = aws.StringValue(entry.Description)
		}

		list = append(list, m)
	}

	return list
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
