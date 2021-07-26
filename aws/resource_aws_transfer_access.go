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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},

			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},

			"home_directory_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.HomeDirectoryTypePath,
				ValidateFunc: validation.StringInSlice([]string{transfer.HomeDirectoryTypePath, transfer.HomeDirectoryTypeLogical}, false),
			},

			"policy": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"posix_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
						},
					},
				},
			},

			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferServerID,
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

	if v, ok := d.GetOk("home_directory_mappings"); ok {
		input.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(v.([]interface{}))
	}

	if v, ok := d.GetOk("home_directory_type"); ok {
		input.HomeDirectoryType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("posix_profile"); ok {
		input.PosixProfile = expandTransferUserPosixUser(v.([]interface{}))
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
	d.Set("home_directory", output.HomeDirectory)
	d.Set("role", output.Role)

	if err := d.Set("home_directory_mappings", flattenAwsTransferHomeDirectoryMappings(output.HomeDirectoryMappings)); err != nil {
		return fmt.Errorf("Error setting home_directory_mappings: %w", err)
	}

	if err := d.Set("posix_profile", flattenTransferUserPosixUser(output.PosixProfile)); err != nil {
		return fmt.Errorf("Error setting posix_profile: %w", err)
	}

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
			input.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(d.Get("home_directory_mappings").([]interface{}))
		}

		if d.HasChange("home_directory_type") {
			input.HomeDirectoryType = aws.String(d.Get("home_directory_type").(string))
		}

		if d.HasChange("policy") {
			input.Policy = aws.String(d.Get("policy").(string))
		}

		if d.HasChange("posix_profile") {
			input.PosixProfile = expandTransferUserPosixUser(d.Get("posix_profile").([]interface{}))
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
