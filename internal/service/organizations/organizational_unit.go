package organizations

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationalUnitCreate,
		Read:   resourceOrganizationalUnitRead,
		Update: resourceOrganizationalUnitUpdate,
		Delete: resourceOrganizationalUnitDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOrganizationalUnitCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Create the organizational unit
	createOpts := &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String(d.Get("name").(string)),
		ParentId: aws.String(d.Get("parent_id").(string)),
		Tags:     Tags(tags.IgnoreAWS()),
	}

	var err error
	var resp *organizations.CreateOrganizationalUnitOutput
	err = resource.Retry(4*time.Minute, func() *resource.RetryError {
		resp, err = conn.CreateOrganizationalUnit(createOpts)

		if tfawserr.ErrCodeEquals(err, organizations.ErrCodeFinalizingOrganizationException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreateOrganizationalUnit(createOpts)
	}

	if err != nil {
		return fmt.Errorf("error creating Organizations Organizational Unit: %w", err)
	}

	// Store the ID
	d.SetId(aws.StringValue(resp.OrganizationalUnit.Id))

	return resourceOrganizationalUnitRead(d, meta)
}

func resourceOrganizationalUnitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeOrganizationalUnit(describeOpts)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Organizations Organizational Unit (%s): %w", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("error reading Organizations Organizational Unit (%s): empty response", d.Id())
	}

	ou := resp.OrganizationalUnit
	if ou == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Organizations Organizational Unit (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	parentId, err := resourceOrganizationalUnitGetParentID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing Organizations Organizational Unit (%s) parents: %w", d.Id(), err)
	}

	var accounts []*organizations.Account
	input := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(d.Id()),
	}

	err = conn.ListAccountsForParentPages(input, func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
		accounts = append(accounts, page.Accounts...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Organizations Organizational Unit (%s) accounts: %w", d.Id(), err)
	}

	if err := d.Set("accounts", flattenOrganizationalUnitAccounts(accounts)); err != nil {
		return fmt.Errorf("error setting accounts: %w", err)
	}

	d.Set("arn", ou.Arn)
	d.Set("name", ou.Name)
	d.Set("parent_id", parentId)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Organizations Organizational Unit (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceOrganizationalUnitUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	if d.HasChange("name") {
		updateOpts := &organizations.UpdateOrganizationalUnitInput{
			Name:                 aws.String(d.Get("name").(string)),
			OrganizationalUnitId: aws.String(d.Id()),
		}

		_, err := conn.UpdateOrganizationalUnit(updateOpts)
		if err != nil {
			return fmt.Errorf("error updating Organizations Organizational Unit (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Organizations Organizational Unit (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceOrganizationalUnitDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	input := &organizations.DeleteOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationalUnit(input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Organizations Organizational Unit (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceOrganizationalUnitGetParentID(conn *organizations.Organizations, childId string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	}
	var parents []*organizations.Parent

	err := conn.ListParentsPages(input, func(page *organizations.ListParentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		parents = append(parents, page.Parents...)

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if len(parents) == 0 {
		return "", nil
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	parent := parents[0]
	return aws.StringValue(parent.Id), nil
}

func flattenOrganizationalUnitAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(accounts) == 0 {
		return nil
	}

	var result []map[string]interface{}

	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"arn":   aws.StringValue(account.Arn),
			"email": aws.StringValue(account.Email),
			"id":    aws.StringValue(account.Id),
			"name":  aws.StringValue(account.Name),
		})
	}

	return result
}
