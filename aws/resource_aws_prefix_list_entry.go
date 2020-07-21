package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsPrefixListEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPrefixListEntryCreate,
		Read:   resourceAwsPrefixListEntryRead,
		Update: resourceAwsPrefixListEntryUpdate,
		Delete: resourceAwsPrefixListEntryDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				ss := strings.Split(d.Id(), "_")
				if len(ss) != 2 || ss[0] == "" || ss[1] == "" {
					return nil, fmt.Errorf("invalid id %s: expected pl-123456_1.0.0.0/8", d.Id())
				}

				d.Set("prefix_list_id", ss[0])
				d.Set("cidr_block", ss[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"prefix_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
		},
	}
}

func resourceAwsPrefixListEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	prefixListId := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	log.Printf(
		"[INFO] adding entry %s to prefix list %s...",
		cidrBlock, prefixListId)

	err := modifyAwsManagedPrefixListConcurrently(
		prefixListId, conn, d.Timeout(schema.TimeoutUpdate),
		ec2.ModifyManagedPrefixListInput{
			PrefixListId:   aws.String(prefixListId),
			CurrentVersion: nil, // set by modifyAwsManagedPrefixListConcurrently
			AddEntries: []*ec2.AddPrefixListEntry{
				{
					Cidr:        aws.String(cidrBlock),
					Description: aws.String(d.Get("description").(string)),
				},
			},
		},
		func(pl *ec2.ManagedPrefixList) *resource.RetryError {
			currentVersion := int(aws.Int64Value(pl.Version))

			_, ok, err := getManagedPrefixListEntryByCIDR(prefixListId, conn, currentVersion, cidrBlock)
			switch {
			case err != nil:
				return resource.NonRetryableError(err)
			case ok:
				return resource.NonRetryableError(errors.New("an entry for this cidr block already exists"))
			}

			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to add entry %s to prefix list %s: %s", cidrBlock, prefixListId, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", prefixListId, cidrBlock))

	return resourceAwsPrefixListEntryRead(d, meta)
}

func resourceAwsPrefixListEntryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	prefixListId := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	entry, ok, err := getManagedPrefixListEntryByCIDR(prefixListId, conn, 0, cidrBlock)
	switch {
	case err != nil:
		return err
	case !ok:
		log.Printf(
			"[WARN] entry %s of managed prefix list %s not found; removing from state.",
			cidrBlock, prefixListId)
		d.SetId("")
		return nil
	}

	d.Set("description", entry.Description)

	return nil
}

func resourceAwsPrefixListEntryUpdate(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("description") {
		return fmt.Errorf("all attributes except description should force new resource")
	}

	conn := meta.(*AWSClient).ec2conn
	prefixListId := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	err := modifyAwsManagedPrefixListConcurrently(
		prefixListId, conn, d.Timeout(schema.TimeoutUpdate),
		ec2.ModifyManagedPrefixListInput{
			PrefixListId:   aws.String(prefixListId),
			CurrentVersion: nil, // set by modifyAwsManagedPrefixListConcurrently
			AddEntries: []*ec2.AddPrefixListEntry{
				{
					Cidr:        aws.String(cidrBlock),
					Description: aws.String(d.Get("description").(string)),
				},
			},
		},
		nil)

	if err != nil {
		return fmt.Errorf("failed to update entry %s in prefix list %s: %s", cidrBlock, prefixListId, err)
	}

	return resourceAwsPrefixListEntryRead(d, meta)
}

func resourceAwsPrefixListEntryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	prefixListId := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	err := modifyAwsManagedPrefixListConcurrently(
		prefixListId, conn, d.Timeout(schema.TimeoutUpdate),
		ec2.ModifyManagedPrefixListInput{
			PrefixListId:   aws.String(prefixListId),
			CurrentVersion: nil, // set by modifyAwsManagedPrefixListConcurrently
			RemoveEntries: []*ec2.RemovePrefixListEntry{
				{
					Cidr: aws.String(cidrBlock),
				},
			},
		},
		nil)

	switch {
	case isResourceNotFoundError(err):
		log.Printf("[WARN] managed prefix list %s not found; removing from state", prefixListId)
		return nil
	case err != nil:
		return fmt.Errorf("failed to remove entry %s from prefix list %s: %s", cidrBlock, prefixListId, err)
	}

	return nil
}

func getManagedPrefixListEntryByCIDR(
	id string,
	conn *ec2.EC2,
	version int,
	cidr string,
) (*ec2.PrefixListEntry, bool, error) {
	input := ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	if version > 0 {
		input.TargetVersion = aws.Int64(int64(version))
	}

	result := (*ec2.PrefixListEntry)(nil)

	err := conn.GetManagedPrefixListEntriesPages(
		&input,
		func(output *ec2.GetManagedPrefixListEntriesOutput, last bool) bool {
			for _, entry := range output.Entries {
				entryCidr := aws.StringValue(entry.Cidr)
				if entryCidr == cidr {
					result = entry
					return false
				}
			}

			return true
		})

	switch {
	case isAWSErr(err, "InvalidPrefixListID.NotFound", ""):
		return nil, false, nil
	case err != nil:
		return nil, false, fmt.Errorf("failed to get entries in prefix list %s: %v", id, err)
	case result == nil:
		return nil, false, nil
	}

	return result, true, nil
}

func modifyAwsManagedPrefixListConcurrently(
	id string,
	conn *ec2.EC2,
	timeout time.Duration,
	input ec2.ModifyManagedPrefixListInput,
	check func(pl *ec2.ManagedPrefixList) *resource.RetryError,
) error {
	isModified := false
	err := resource.Retry(timeout, func() *resource.RetryError {
		if !isModified {
			pl, ok, err := getManagedPrefixList(id, conn)
			switch {
			case err != nil:
				return resource.NonRetryableError(err)
			case !ok:
				return resource.NonRetryableError(&resource.NotFoundError{})
			}

			input.CurrentVersion = pl.Version

			if check != nil {
				if err := check(pl); err != nil {
					return err
				}
			}

			switch _, err := conn.ModifyManagedPrefixList(&input); {
			case isManagedPrefixListModificationConflictErr(err):
				return resource.RetryableError(err)
			case err != nil:
				return resource.NonRetryableError(fmt.Errorf("modify failed: %s", err))
			}

			isModified = true
		}

		switch settled, err := isAwsManagedPrefixListSettled(id, conn); {
		case err != nil:
			return resource.NonRetryableError(fmt.Errorf("resource failed to settle: %s", err))
		case !settled:
			return resource.RetryableError(errors.New("resource not yet settled"))
		}

		return nil
	})

	switch {
	case isResourceTimeoutError(err):
		return fmt.Errorf("timed out: %s", err)
	case err != nil:
		return err
	}

	return nil
}

func isManagedPrefixListModificationConflictErr(err error) bool {
	return isAWSErr(err, "IncorrectState", "in the current state (modify-in-progress)") ||
		isAWSErr(err, "IncorrectState", "in the current state (create-in-progress)") ||
		isAWSErr(err, "PrefixListVersionMismatch", "") ||
		isAWSErr(err, "ConcurrentMutationLimitExceeded", "")
}
