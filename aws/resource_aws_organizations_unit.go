package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsOrganizationsUnit() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOrganizationsUnitCreate,
		Read:   resourceAwsOrganizationsUnitRead,
		Update: resourceAwsOrganizationsUnitUpdate,
		Delete: resourceAwsOrganizationsUnitDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"parent_id": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateOrganizationalUnit.html#organizations-CreateOrganizationalUnit-request-ParentId"),
			},
		},
	}
}

func resourceAwsOrganizationsUnitCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	// Create the organizational unit
	createOpts := &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String(d.Get("name").(string)),
		ParentId: aws.String(d.Get("parent_id").(string)),
	}

	log.Printf("[DEBUG] Organizational Unit create config: %#v", createOpts)

	var err error
	var resp *organizations.CreateOrganizationalUnitOutput
	err = resource.Retry(4*time.Minute, func() *resource.RetryError {
		resp, err = conn.CreateOrganizationalUnit(createOpts)

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeFinalizingOrganizationException, "") {
				log.Printf("[DEBUG] Trying to create organizational unit again: %q", err.Error())
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Error creating organizational unit: %s", err)
	}
	log.Printf("[DEBUG] Organizational Unit create response: %#v", resp)

	// Store the ID
	ouId := resp.OrganizationalUnit.Id
	d.SetId(*ouId)

	return resourceAwsOrganizationsUnitRead(d, meta)
}

func resourceAwsOrganizationsUnitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn
	describeOpts := &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeOrganizationalUnit(describeOpts)
	if err != nil {
		if isAWSErr(err, organizations.ErrCodeOrganizationalUnitNotFoundException, "") {
			log.Printf("[WARN] Organizational Unit does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	ou := resp.OrganizationalUnit
	if ou == nil {
		log.Printf("[WARN] Organizational Unit does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	parentId, err := resourceAwsOrganizationsGetParentID(conn, d.Id())
	if err != nil {
		log.Printf("[WARN] Unable to find parent organizational unit, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", ou.Arn)
	d.Set("name", ou.Name)
	d.Set("parent_id", parentId)
	return nil
}

func resourceAwsOrganizationsUnitUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("name") {
		conn := meta.(*AWSClient).organizationsconn

		updateOpts := &organizations.UpdateOrganizationalUnitInput{
			Name:                 aws.String(d.Get("name").(string)),
			OrganizationalUnitId: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Organizational Unit update config: %#v", updateOpts)
		resp, err := conn.UpdateOrganizationalUnit(updateOpts)
		if err != nil {
			return fmt.Errorf("Error creating organizational unit: %s", err)
		}
		log.Printf("[DEBUG] Organizational Unit update response: %#v", resp)
	}

	return nil
}

func resourceAwsOrganizationsUnitDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.DeleteOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Removing AWS organizational unit from organization: %s", input)
	_, err := conn.DeleteOrganizationalUnit(input)
	if err != nil {
		if isAWSErr(err, organizations.ErrCodeOrganizationalUnitNotFoundException, "") {
			return nil
		}
		return err
	}
	return nil
}

func resourceAwsOrganizationsGetParentID(conn *organizations.Organizations, childId string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	}
	resp, err := conn.ListParents(input)
	if err != nil {
		return "", err
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	parent := resp.Parents[0]
	return aws.StringValue(parent.Id), nil
}
