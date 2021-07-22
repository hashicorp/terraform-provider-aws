package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	tfkms "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateAwsKmsName,
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
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
				DiffSuppressFunc: suppressEquivalentTargetKeyIdAndARN,
			},
		},
	}
}

const (
	kmsAliasDefaultNamePrefix = "alias/"
)

func resourceAwsKmsAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	namePrefix := d.Get("name_prefix").(string)
	if namePrefix == "" {
		namePrefix = kmsAliasDefaultNamePrefix
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
		err := resourceAwsKmsAliasTargetUpdate(conn, d)
		if err != nil {
			return err
		}
		return resourceAwsKmsAliasRead(d, meta)
	}
	return nil
}

func resourceAwsKmsAliasTargetUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	name := d.Get("name").(string)
	targetKeyId := d.Get("target_key_id").(string)

	log.Printf("[DEBUG] KMS alias: %s, update target: %s", name, targetKeyId)

	req := &kms.UpdateAliasInput{
		AliasName:   aws.String(name),
		TargetKeyId: aws.String(targetKeyId),
	}
	_, err := conn.UpdateAlias(req)

	return err
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

// API by default limits results to 50 aliases
// This is how we make sure we won't miss any alias
// See http://docs.aws.amazon.com/kms/latest/APIReference/API_ListAliases.html
func findKmsAliasByName(conn *kms.KMS, name string, marker *string) (*kms.AliasListEntry, error) {
	req := kms.ListAliasesInput{
		Limit: aws.Int64(int64(100)),
	}
	if marker != nil {
		req.Marker = marker
	}

	log.Printf("[DEBUG] Listing KMS aliases: %s", req)
	resp, err := conn.ListAliases(&req)
	if err != nil {
		return nil, err
	}

	for _, entry := range resp.Aliases {
		if *entry.AliasName == name {
			return entry, nil
		}
	}
	if *resp.Truncated {
		log.Printf("[DEBUG] KMS alias list is truncated, listing more via %s", *resp.NextMarker)
		return findKmsAliasByName(conn, name, resp.NextMarker)
	}

	return nil, nil
}

func suppressEquivalentTargetKeyIdAndARN(k, old, new string, d *schema.ResourceData) bool {
	newARN, err := arn.Parse(new)
	if err != nil {
		log.Printf("[DEBUG] %q can not be parsed as an ARN: %q", new, err)
		return false
	}

	resource := strings.TrimPrefix(newARN.Resource, "key/")
	return old == resource
}
