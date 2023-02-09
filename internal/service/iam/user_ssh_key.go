package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUserSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserSSHKeyCreate,
		ReadWithoutTimeout:   resourceUserSSHKeyRead,
		UpdateWithoutTimeout: resourceUserSSHKeyUpdate,
		DeleteWithoutTimeout: resourceUserSSHKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserSSHKeyImport,
		},

		Schema: map[string]*schema.Schema{
			"encoding": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(iam.EncodingType_Values(), false),
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
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
			"ssh_public_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	username := d.Get("username").(string)
	input := &iam.UploadSSHPublicKeyInput{
		SSHPublicKeyBody: aws.String(d.Get("public_key").(string)),
		UserName:         aws.String(username),
	}

	output, err := conn.UploadSSHPublicKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "uploading IAM User SSH Key (%s): %s", username, err)
	}

	d.SetId(aws.StringValue(output.SSHPublicKey.SSHPublicKeyId))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindSSHPublicKeyByThreePartKey(ctx, conn, d.Id(), d.Get("encoding").(string), username)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IAM User SSH Key (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("status"); ok {
		input := &iam.UpdateSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(d.Id()),
			Status:         aws.String(v.(string)),
			UserName:       aws.String(username),
		}

		_, err := conn.UpdateSSHPublicKeyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User SSH Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserSSHKeyRead(ctx, d, meta)...)
}

func resourceUserSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	encoding := d.Get("encoding").(string)
	key, err := FindSSHPublicKeyByThreePartKey(ctx, conn, d.Id(), encoding, d.Get("username").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM User SSH Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User SSH Key (%s): %s", d.Id(), err)
	}

	d.Set("fingerprint", key.Fingerprint)
	publicKey := aws.StringValue(key.SSHPublicKeyBody)
	if encoding == iam.EncodingTypeSsh {
		publicKey = cleanSSHKey(publicKey)
	}
	d.Set("public_key", publicKey)
	d.Set("ssh_public_key_id", key.SSHPublicKeyId)
	d.Set("status", key.Status)

	return diags
}

func resourceUserSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	input := &iam.UpdateSSHPublicKeyInput{
		SSHPublicKeyId: aws.String(d.Id()),
		Status:         aws.String(d.Get("status").(string)),
		UserName:       aws.String(d.Get("username").(string)),
	}

	_, err := conn.UpdateSSHPublicKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM User SSH Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserSSHKeyRead(ctx, d, meta)...)
}

func resourceUserSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	log.Printf("[DEBUG] Deleting IAM User SSH Key: %s", d.Id())
	_, err := conn.DeleteSSHPublicKeyWithContext(ctx, &iam.DeleteSSHPublicKeyInput{
		SSHPublicKeyId: aws.String(d.Id()),
		UserName:       aws.String(d.Get("username").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User SSH Key (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceUserSSHKeyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func FindSSHPublicKeyByThreePartKey(ctx context.Context, conn *iam.IAM, id, encoding, username string) (*iam.SSHPublicKey, error) {
	input := &iam.GetSSHPublicKeyInput{
		Encoding:       aws.String(encoding),
		SSHPublicKeyId: aws.String(id),
		UserName:       aws.String(username),
	}

	output, err := conn.GetSSHPublicKeyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SSHPublicKey == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SSHPublicKey, nil
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
