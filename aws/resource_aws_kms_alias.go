package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/naming"
	tfkms "github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsKmsAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKmsAliasCreate,
		Read:   resourceAwsKmsAliasRead,
		Update: resourceAwsKmsAliasUpdate,
		Delete: resourceAwsKmsAliasDelete,

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
				ValidateFunc:  validateAwsKmsName,
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateAwsKmsName,
			},

			"target_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"target_key_id": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentKmsKeyARNOrID,
			},
		},
	}
}

func resourceAwsKmsAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	namePrefix := d.Get("name_prefix").(string)
	if namePrefix == "" {
		namePrefix = tfkms.AliasNamePrefix
	}
	name := naming.Generate(d.Get("name").(string), namePrefix)

	input := &kms.CreateAliasInput{
		AliasName:   aws.String(name),
		TargetKeyId: aws.String(d.Get("target_key_id").(string)),
	}

	// KMS is eventually consistent.
	log.Printf("[DEBUG] Creating KMS Alias: %s", input)

	_, err := tfresource.RetryWhenAwsErrCodeEquals(waiter.KeyRotationUpdatedTimeout, func() (interface{}, error) {
		return conn.CreateAlias(input)
	}, kms.ErrCodeNotFoundException)

	if err != nil {
		return fmt.Errorf("error creating KMS Alias (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceAwsKmsAliasRead(d, meta)
}

func resourceAwsKmsAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(waiter.PropagationTimeout, func() (interface{}, error) {
		return finder.AliasByName(conn, d.Id())
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
	targetKeyARN, err := tfkms.AliasARNToKeyARN(aliasARN, targetKeyID)

	if err != nil {
		return err
	}

	d.Set("arn", aliasARN)
	d.Set("name", alias.AliasName)
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(alias.AliasName)))
	d.Set("target_key_arn", targetKeyARN)
	d.Set("target_key_id", targetKeyID)

	return nil
}

func resourceAwsKmsAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

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

	return resourceAwsKmsAliasRead(d, meta)
}

func resourceAwsKmsAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

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

func suppressEquivalentKmsKeyARNOrID(k, old, new string, d *schema.ResourceData) bool {
	return tfkms.KeyARNOrIDEqual(old, new)
}
