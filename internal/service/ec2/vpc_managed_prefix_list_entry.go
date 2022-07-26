package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceManagedPrefixListEntry() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceManagedPrefixListEntryCreate,
		Read:   resourceManagedPrefixListEntryRead,
		Delete: resourceManagedPrefixListEntryDelete,
		Importer: &schema.ResourceImporter{
			State: resourceManagedPrefixListEntryImport,
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

func resourceManagedPrefixListEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidr := d.Get("cidr").(string)
	plID := d.Get("prefix_list_id").(string)
	id := ManagedPrefixListEntryCreateID(plID, cidr)

	addPrefixListEntry := &ec2.AddPrefixListEntry{Cidr: aws.String(cidr)}

	if v, ok := d.GetOk("description"); ok {
		addPrefixListEntry.Description = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		mutexKey := fmt.Sprintf("vpc-managed-prefix-list-%s", plID)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		pl, err := FindManagedPrefixListByID(conn, plID)

		if err != nil {
			return nil, fmt.Errorf("error reading VPC Managed Prefix List (%s): %w", plID, err)
		}

		input := &ec2.ModifyManagedPrefixListInput{
			AddEntries:     []*ec2.AddPrefixListEntry{addPrefixListEntry},
			CurrentVersion: pl.Version,
			PrefixListId:   aws.String(plID),
		}

		return conn.ModifyManagedPrefixList(input)
	}, "IncorrectState", "PrefixListVersionMismatch")

	if err != nil {
		return fmt.Errorf("error creating VPC Managed Prefix List Entry (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitManagedPrefixListModified(conn, plID); err != nil {
		return fmt.Errorf("error waiting for VPC Managed Prefix List Entry (%s) create: %w", d.Id(), err)
	}

	return resourceManagedPrefixListEntryRead(d, meta)
}

func resourceManagedPrefixListEntryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	plID, cidr, err := ManagedPrefixListEntryParseID(d.Id())

	if err != nil {
		return err
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ManagedPrefixListEntryCreateTimeout, func() (interface{}, error) {
		return FindManagedPrefixListEntryByIDAndCIDR(conn, plID, cidr)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Managed Prefix List Entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Managed Prefix List Entry (%s): %w", d.Id(), err)
	}

	entry := outputRaw.(*ec2.PrefixListEntry)

	d.Set("cidr", entry.Cidr)
	d.Set("description", entry.Description)

	return nil
}

func resourceManagedPrefixListEntryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	plID, cidr, err := ManagedPrefixListEntryParseID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing VPC Managed Prefix List Entry ID (%s): %w", d.Id(), err)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		mutexKey := fmt.Sprintf("vpc-managed-prefix-list-%s", plID)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		pl, err := FindManagedPrefixListByID(conn, plID)

		if err != nil {
			return nil, fmt.Errorf("error reading VPC Managed Prefix List (%s): %w", plID, err)
		}

		input := &ec2.ModifyManagedPrefixListInput{
			CurrentVersion: pl.Version,
			PrefixListId:   aws.String(plID),
			RemoveEntries:  []*ec2.RemovePrefixListEntry{{Cidr: aws.String(cidr)}},
		}

		return conn.ModifyManagedPrefixList(input)
	}, "IncorrectState", "PrefixListVersionMismatch")

	if err != nil {
		return fmt.Errorf("error deleting VPC Managed Prefix List Entry (%s): %w", d.Id(), err)
	}

	_, err = WaitManagedPrefixListModified(conn, plID)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Managed Prefix List Entry (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourceManagedPrefixListEntryImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	plID, cidr, err := ManagedPrefixListEntryParseID(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("cidr", cidr)
	d.Set("prefix_list_id", plID)

	return []*schema.ResourceData{d}, nil
}
