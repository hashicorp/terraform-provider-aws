package redshift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAuthenticationProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAuthenticationProfileCreate,
		Read:   resourceAuthenticationProfileRead,
		Update: resourceAuthenticationProfileUpdate,
		Delete: resourceAuthenticationProfileDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"authentication_profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authentication_profile_content": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceAuthenticationProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	authProfileName := d.Get("authentication_profile_name").(string)

	input := redshift.CreateAuthenticationProfileInput{
		AuthenticationProfileName:    aws.String(authProfileName),
		AuthenticationProfileContent: aws.String(d.Get("authentication_profile_content").(string)),
	}

	out, err := conn.CreateAuthenticationProfile(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Authentication Profile (%s): %s", authProfileName, err)
	}

	d.SetId(aws.StringValue(out.AuthenticationProfileName))

	return resourceAuthenticationProfileRead(d, meta)
}

func resourceAuthenticationProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	out, err := FindAuthenticationProfileByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Authentication Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Authentication Profile (%s): %w", d.Id(), err)
	}

	d.Set("authentication_profile_content", out.AuthenticationProfileContent)
	d.Set("authentication_profile_name", out.AuthenticationProfileName)

	return nil
}

func resourceAuthenticationProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	input := &redshift.ModifyAuthenticationProfileInput{
		AuthenticationProfileName:    aws.String(d.Id()),
		AuthenticationProfileContent: aws.String(d.Get("authentication_profile_content").(string)),
	}

	_, err := conn.ModifyAuthenticationProfile(input)

	if err != nil {
		return fmt.Errorf("error modifying Redshift Authentication Profile (%s): %w", d.Id(), err)
	}

	return resourceAuthenticationProfileRead(d, meta)
}

func resourceAuthenticationProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	deleteInput := redshift.DeleteAuthenticationProfileInput{
		AuthenticationProfileName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift Authentication Profile: %s", d.Id())
	_, err := conn.DeleteAuthenticationProfile(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeAuthenticationProfileNotFoundFault) {
			return nil
		}
		return err
	}

	return err
}
