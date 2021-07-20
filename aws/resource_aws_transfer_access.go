package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsTransferAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTransferAccessCreate,
		Read:   resourceAwsTransferAccessRead,
		Update: resourceAwsTransferAccessUpdate,
		Delete: resourceAwsTransferAccessDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"external_id": {
				Type:     schema.TypeString,
				Optional: false,
			},

			"home_directory": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"home_directory_mappings": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
			},

			"home_directory_type": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(transfer.HomeDirectoryType_Values(), false),
				},
			},

			"policy": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"posix_profile": {
				Type:     schema.TypeSet,
				Optional: true,
				//TODO: this
			},

			"role": {
				Type:     schema.TypeString,
				Optional: false,
				//TODO: Min length 20
			},

			"server_id": {
				Type:     schema.TypeString,
				Optional: false,
				//TODO: Min length 19
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsTransferAccessCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	input := &transfer.CreateAccessInput{}

	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("home_directory"); ok {
		input.HomeDirectory = aws.String(v.(string))
	}

	if _, ok := d.GetOk("home_directory_mappings"); ok {
		//TODO this
	}

	if v, ok := d.GetOk("home_directory_type"); ok {
		input.HomeDirectoryType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if _, ok := d.GetOk("posix_profile"); ok {
		//TODO this
	}

	if v, ok := d.GetOk("role"); ok {
		input.Role = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_id"); ok {
		input.ServerId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Access: %s", input)
	output, err := conn.CreateAccess(input)

	if err != nil {
		return fmt.Errorf("error creating Access: %w", err)
	}

	d.SetId(aws.StringValue(output.ExternalId))

	return resourceAwsTransferAccessRead(d, meta)
}

func resourceAwsTransferAccessRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	externalId := d.Get("external_id").(string)
	serverId := d.Get("server_id").(string)
	output, err := finder.AccessByID(conn, serverId, externalId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Access with external ID (%s) not found for server (%s), removing from state", externalId, serverId)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Access with external ID (%s) for server (%s): %w", externalId, serverId, err)
	}

	d.Set("external_id", output.ExternalId)
	d.Set("policy", output.Policy)
	d.Set("home_directory_type", output.HomeDirectoryType)
	d.Set("home_directory_mappings", output.HomeDirectoryMappings)
	d.Set("posix_profile", output.PosixProfile)
	d.Set("home_directory", output.HomeDirectory)
	d.Set("role", output.Role)

	return nil
}

func resourceAwsTransferAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	input := &transfer.UpdateAccessInput{}

	if d.HasChangesExcept() {
		if d.HasChange("external_id") {
			input.ExternalId = aws.String(d.Get("external_id").(string))
		}

		if d.HasChange("home_directory") {
			input.HomeDirectory = aws.String(d.Get("home_directory").(string))
		}

		if d.HasChange("home_directory_mappings") {
			//TODO: This
		}

		if d.HasChange("home_directory_type") {
			input.HomeDirectoryType = aws.String(d.Get("home_directory_type").(string))
		}

		if d.HasChange("policy") {
			input.Policy = aws.String(d.Get("policy").(string))
		}

		if d.HasChange("posix_profile") {
			//TODO: This
		}

		if d.HasChange("role") {
			input.Role = aws.String(d.Get("role").(string))
		}

		if d.HasChange("server_id") {
			input.ServerId = aws.String(d.Get("server_id").(string))
		}

		log.Printf("[DEBUG] Updating Transfer Access: %s", input)
		_, err := conn.UpdateAccess(input)
		if err != nil {
			return fmt.Errorf("error updating Transfer Access (externalID: %s, serverID: %s): %w", aws.StringValue(input.ExternalId), aws.StringValue(input.ServerId), err)
		}
	}

	return resourceAwsTransferAccessRead(d, meta)
}

func resourceAwsTransferAccessDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	externalId := d.Get("external_id").(string)
	serverId := d.Get("server_id").(string)

	log.Printf("[DEBUG] Deleting Transfer Access: (externalID: %s, serverID: %s)", externalId, serverId)
	_, err := conn.DeleteAccess(&transfer.DeleteAccessInput{
		ExternalId: aws.String(externalId),
		ServerId:   aws.String(serverId),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Transfer Access (externalID: %s, serverID: %s): %w", externalId, serverId, err)
	}

	return nil
}
