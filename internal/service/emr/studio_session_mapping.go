package emr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStudioSessionMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceStudioSessionMappingCreate,
		Read:   resourceStudioSessionMappingRead,
		Update: resourceStudioSessionMappingUpdate,
		Delete: resourceStudioSessionMappingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"identity_id", "identity_name"},
			},
			"identity_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"identity_id", "identity_name"},
			},
			"identity_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(emr.IdentityType_Values(), false),
			},
			"session_policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"studio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStudioSessionMappingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn

	var id string
	studioId := d.Get("studio_id").(string)
	identityType := d.Get("identity_type").(string)
	input := &emr.CreateStudioSessionMappingInput{
		IdentityType:     aws.String(identityType),
		SessionPolicyArn: aws.String(d.Get("session_policy_arn").(string)),
		StudioId:         aws.String(studioId),
	}

	if v, ok := d.GetOk("identity_id"); ok {
		input.IdentityId = aws.String(v.(string))
		id = v.(string)
	}

	if v, ok := d.GetOk("identity_name"); ok {
		input.IdentityName = aws.String(v.(string))
		id = v.(string)
	}

	_, err := conn.CreateStudioSessionMapping(input)
	if err != nil {
		return fmt.Errorf("error creating EMR Studio Session Mapping: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", studioId, identityType, id))

	return resourceStudioSessionMappingRead(d, meta)
}

func resourceStudioSessionMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn

	studioId, identityType, identityId, err := readStudioSessionMapping(d.Id())
	if err != nil {
		return err
	}

	input := &emr.UpdateStudioSessionMappingInput{
		SessionPolicyArn: aws.String(d.Get("session_policy_arn").(string)),
		IdentityType:     aws.String(identityType),
		StudioId:         aws.String(studioId),
		IdentityId:       aws.String(identityId),
	}

	_, err = conn.UpdateStudioSessionMapping(input)
	if err != nil {
		return fmt.Errorf("error updating EMR Studio Session Mapping: %w", err)
	}

	return resourceStudioSessionMappingRead(d, meta)
}

func resourceStudioSessionMappingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn

	mapping, err := FindStudioSessionMappingByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Studio Session Mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EMR Studio Session Mapping (%s): %w", d.Id(), err)
	}

	d.Set("identity_type", mapping.IdentityType)
	d.Set("identity_id", mapping.IdentityId)
	d.Set("identity_name", mapping.IdentityName)
	d.Set("studio_id", mapping.StudioId)
	d.Set("session_policy_arn", mapping.SessionPolicyArn)

	return nil
}

func resourceStudioSessionMappingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn
	studioId, identityType, identityId, err := readStudioSessionMapping(d.Id())
	if err != nil {
		return err
	}

	input := &emr.DeleteStudioSessionMappingInput{
		IdentityType: aws.String(identityType),
		StudioId:     aws.String(studioId),
		IdentityId:   aws.String(identityId),
	}

	log.Printf("[INFO] Deleting EMR Studio Session Mapping: %s", d.Id())
	_, err = conn.DeleteStudioSessionMapping(input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "Studio session mapping does not exist.") {
			return nil
		}
		return fmt.Errorf("error deleting EMR Studio Session Mapping (%s): %w", d.Id(), err)
	}

	return nil
}
