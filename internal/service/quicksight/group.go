package quicksight

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
				ValidateFunc: validation.StringInSlice([]string{
					"default",
				}, false),
			},
		},
	}
}

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID := meta.(*conns.AWSClient).AccountID
	namespace := d.Get("namespace").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.CreateGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(d.Get("group_name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateGroup(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating QuickSight Group: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(resp.Group.GroupName)))

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID, namespace, groupName, err := GroupParseID(d.Id())
	if err != nil {
		return err
	}

	descOpts := &quicksight.DescribeGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(groupName),
	}

	resp, err := conn.DescribeGroup(descOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] QuickSight Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing QuickSight Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.Group.Arn)
	d.Set("aws_account_id", awsAccountID)
	d.Set("group_name", resp.Group.GroupName)
	d.Set("description", resp.Group.Description)
	d.Set("namespace", namespace)

	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID, namespace, groupName, err := GroupParseID(d.Id())
	if err != nil {
		return err
	}

	updateOpts := &quicksight.UpdateGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(groupName),
	}

	if v, ok := d.GetOk("description"); ok {
		updateOpts.Description = aws.String(v.(string))
	}

	_, err = conn.UpdateGroup(updateOpts)
	if err != nil {
		return fmt.Errorf("error updating QuickSight Group %s: %w", d.Id(), err)
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).QuickSightConn

	awsAccountID, namespace, groupName, err := GroupParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteGroupInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(groupName),
	}

	if _, err := conn.DeleteGroup(deleteOpts); err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight Group %s: %s", d.Id(), err)
	}

	return nil
}

func GroupParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/GROUP_NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
