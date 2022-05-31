package cloud9

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEnvironmentMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvironmentMembershipCreate,
		Read:   resourceEnvironmentMembershipRead,
		Update: resourceEnvironmentMembershipUpdate,
		Delete: resourceEnvironmentMembershipDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permissions": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloud9.Permissions_Values(), false),
			},
			"user_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceEnvironmentMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn

	envId := d.Get("environment_id").(string)
	userArn := d.Get("user_arn").(string)
	input := &cloud9.CreateEnvironmentMembershipInput{
		EnvironmentId: aws.String(envId),
		Permissions:   aws.String(d.Get("permissions").(string)),
		UserArn:       aws.String(userArn),
	}

	_, err := conn.CreateEnvironmentMembership(input)

	if err != nil {
		return fmt.Errorf("error creating Cloud9 Environment Membership: %w", err)
	}

	d.SetId(fmt.Sprintf("%s#%s", envId, userArn))

	return resourceEnvironmentMembershipRead(d, meta)
}

func resourceEnvironmentMembershipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn

	envId, userArn, err := DecodeEnviornmentMemberId(d.Id())
	if err != nil {
		return err
	}

	env, err := FindEnvironmentMembershipByID(conn, envId, userArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cloud9 EC2 Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Cloud9 EC2 Environment (%s): %w", d.Id(), err)
	}

	d.Set("environment_id", env.EnvironmentId)
	d.Set("user_arn", env.UserArn)
	d.Set("user_id", env.UserId)
	d.Set("permissions", env.Permissions)

	return nil
}

func resourceEnvironmentMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn

	input := cloud9.UpdateEnvironmentMembershipInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		Permissions:   aws.String(d.Get("permissions").(string)),
		UserArn:       aws.String(d.Get("user_arn").(string)),
	}

	log.Printf("[INFO] Updating Cloud9 Environment Membership: %#v", input)
	_, err := conn.UpdateEnvironmentMembership(&input)

	if err != nil {
		return fmt.Errorf("error updating Cloud9 Environment Membership (%s): %w", d.Id(), err)
	}

	return resourceEnvironmentMembershipRead(d, meta)
}

func resourceEnvironmentMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn

	_, err := conn.DeleteEnvironmentMembership(&cloud9.DeleteEnvironmentMembershipInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		UserArn:       aws.String(d.Get("user_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Cloud9 EC2 Environment (%s): %w", d.Id(), err)
	}

	return nil
}

func DecodeEnviornmentMemberId(id string) (string, string, error) {
	idParts := strings.Split(id, "#")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected ENVIRONMENT-ID#USER-ARN", id)
	}
	envId := idParts[0]
	userArn := idParts[1]

	return envId, userArn, nil
}
