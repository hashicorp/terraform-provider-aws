package iam

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUserSSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserSSHKeyCreate,
		Read:   resourceUserSSHKeyRead,
		Update: resourceUserSSHKeyUpdate,
		Delete: resourceUserSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUserSSHKeyImport,
		},

		Schema: map[string]*schema.Schema{
			"ssh_public_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("encoding").(string) == "SSH" {
						old = cleanSSHKey(old)
						new = cleanSSHKey(new)
					}
					return strings.Trim(old, "\n") == strings.Trim(new, "\n")
				},
			},

			"encoding": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					iam.EncodingTypeSsh,
					iam.EncodingTypePem,
				}, false),
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUserSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	username := d.Get("username").(string)
	publicKey := d.Get("public_key").(string)

	request := &iam.UploadSSHPublicKeyInput{
		UserName:         aws.String(username),
		SSHPublicKeyBody: aws.String(publicKey),
	}

	log.Println("[DEBUG] Create IAM User SSH Key Request:", request)
	createResp, err := conn.UploadSSHPublicKey(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM User SSH Key %s: %s", username, err)
	}

	d.SetId(aws.StringValue(createResp.SSHPublicKey.SSHPublicKeyId))

	return resourceUserSSHKeyUpdate(d, meta)
}

func resourceUserSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	username := d.Get("username").(string)
	encoding := d.Get("encoding").(string)
	request := &iam.GetSSHPublicKeyInput{
		UserName:       aws.String(username),
		SSHPublicKeyId: aws.String(d.Id()),
		Encoding:       aws.String(encoding),
	}

	var getResp *iam.GetSSHPublicKeyOutput

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		getResp, err = conn.GetSSHPublicKey(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetSSHPublicKey(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User SSH Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM User SSH Key (%s): %w", d.Id(), err)
	}

	if getResp == nil || getResp.SSHPublicKey == nil {
		return fmt.Errorf("error reading IAM User SSH Key (%s): empty response", d.Id())
	}

	publicKey := aws.StringValue(getResp.SSHPublicKey.SSHPublicKeyBody)
	if encoding == iam.EncodingTypeSsh {
		publicKey = cleanSSHKey(publicKey)
	}

	d.Set("fingerprint", getResp.SSHPublicKey.Fingerprint)
	d.Set("status", getResp.SSHPublicKey.Status)
	d.Set("ssh_public_key_id", getResp.SSHPublicKey.SSHPublicKeyId)
	d.Set("public_key", publicKey)
	return nil
}

func resourceUserSSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("status") {
		conn := meta.(*conns.AWSClient).IAMConn

		request := &iam.UpdateSSHPublicKeyInput{
			UserName:       aws.String(d.Get("username").(string)),
			SSHPublicKeyId: aws.String(d.Id()),
			Status:         aws.String(d.Get("status").(string)),
		}

		_, err := conn.UpdateSSHPublicKey(request)
		if err != nil {
			return fmt.Errorf("error updating IAM User SSH Key (%s): %w", d.Id(), err)
		}
	}
	return resourceUserSSHKeyRead(d, meta)
}

func resourceUserSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.DeleteSSHPublicKeyInput{
		UserName:       aws.String(d.Get("username").(string)),
		SSHPublicKeyId: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete IAM User SSH Key request:", request)
	if _, err := conn.DeleteSSHPublicKey(request); err != nil {
		return fmt.Errorf("error deleting IAM User SSH Key (%s): %w", d.Id(), err)
	}
	return nil
}

func resourceUserSSHKeyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), ":", 3)

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), UserName:SSHPublicKeyId:Encoding", d.Id())
	}

	username := idParts[0]
	sshPublicKeyId := idParts[1]
	encoding := idParts[2]

	d.Set("username", username)
	d.Set("ssh_public_key_id", sshPublicKeyId)
	d.Set("encoding", encoding)
	d.SetId(sshPublicKeyId)

	return []*schema.ResourceData{d}, nil
}

func cleanSSHKey(key string) string {
	// Remove comments from SSH Keys
	// Comments are anything after "ssh-rsa XXXX" where XXXX is the key.
	parts := strings.Split(key, " ")
	if len(parts) > 2 {
		parts = parts[0:2]
	}
	return strings.Join(parts, " ")
}
