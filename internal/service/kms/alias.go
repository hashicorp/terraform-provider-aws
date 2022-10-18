package kms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliasCreate,
		Read:   resourceAliasRead,
		Update: resourceAliasUpdate,
		Delete: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validNameForResource,
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validNameForResource,
			},

			"target_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"target_key_id": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentKeyARNOrID,
			},
		},
	}
}

func resourceAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	namePrefix := d.Get("name_prefix").(string)
	if namePrefix == "" {
		namePrefix = AliasNamePrefix
	}
	name := create.Name(d.Get("name").(string), namePrefix)

	input := &kms.CreateAliasInput{
		AliasName:   aws.String(name),
		TargetKeyId: aws.String(d.Get("target_key_id").(string)),
	}

	// KMS is eventually consistent.
	log.Printf("[DEBUG] Creating KMS Alias: %s", input)

	_, err := tfresource.RetryWhenAWSErrCodeEquals(KeyRotationUpdatedTimeout, func() (interface{}, error) {
		return conn.CreateAlias(input)
	}, kms.ErrCodeNotFoundException)

	if err != nil {
		return fmt.Errorf("error creating KMS Alias (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceAliasRead(d, meta)
}

func resourceAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(PropagationTimeout, func() (interface{}, error) {
		return FindAliasByName(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading KMS Alias (%s): %w", d.Id(), err)
	}

	alias := outputRaw.(*kms.AliasListEntry)
	aliasARN := aws.StringValue(alias.AliasArn)
	targetKeyID := aws.StringValue(alias.TargetKeyId)
	targetKeyARN, err := AliasARNToKeyARN(aliasARN, targetKeyID)

	if err != nil {
		return err
	}

	d.Set("arn", aliasARN)
	d.Set("name", alias.AliasName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(alias.AliasName)))
	d.Set("target_key_arn", targetKeyARN)
	d.Set("target_key_id", targetKeyID)

	return nil
}

func resourceAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	if d.HasChange("target_key_id") {
		input := &kms.UpdateAliasInput{
			AliasName:   aws.String(d.Id()),
			TargetKeyId: aws.String(d.Get("target_key_id").(string)),
		}

		log.Printf("[DEBUG] Updating KMS Alias: %s", input)
		_, err := conn.UpdateAlias(input)

		if err != nil {
			return fmt.Errorf("error updating KMS Alias (%s): %w", d.Id(), err)
		}
	}

	return resourceAliasRead(d, meta)
}

func resourceAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	log.Printf("[DEBUG] Deleting KMS Alias: (%s)", d.Id())
	_, err := conn.DeleteAlias(&kms.DeleteAliasInput{
		AliasName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting KMS Alias (%s): %w", d.Id(), err)
	}

	return nil
}

func suppressEquivalentKeyARNOrID(k, old, new string, d *schema.ResourceData) bool {
	return KeyARNOrIDEqual(old, new)
}
