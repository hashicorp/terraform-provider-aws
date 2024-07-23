// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_ssh_key", name="SSH Key")
func resourceSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSSHKeyCreate,
		ReadWithoutTimeout:   resourceSSHKeyRead,
		DeleteWithoutTimeout: resourceSSHKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"body": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					old = cleanSSHKey(old)
					new = cleanSSHKey(new)
					return strings.Trim(old, "\n") == strings.Trim(new, "\n")
				},
			},
			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validServerID,
			},
			"ssh_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserName,
			},
		},
	}
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	userName := d.Get(names.AttrUserName).(string)
	serverID := d.Get("server_id").(string)
	input := &transfer.ImportSshPublicKeyInput{
		ServerId:         aws.String(serverID),
		SshPublicKeyBody: aws.String(d.Get("body").(string)),
		UserName:         aws.String(userName),
	}

	output, err := conn.ImportSshPublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "importing Transfer SSH Key: %s", err)
	}

	d.SetId(sshKeyCreateResourceID(serverID, userName, aws.ToString(output.SshPublicKeyId)))

	return append(diags, resourceSSHKeyRead(ctx, d, meta)...)
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, userName, sshKeyID, err := sshKeyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	user, sshKey, err := findUserSSHKeyByThreePartKey(ctx, conn, serverID, userName, sshKeyID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer SSH Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer SSH Key (%s): %s", d.Id(), err)
	}

	d.Set("body", sshKey.SshPublicKeyBody)
	d.Set("server_id", serverID)
	d.Set("ssh_key_id", sshKey.SshPublicKeyId)
	d.Set(names.AttrUserName, user.UserName)

	return diags
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, userName, sshKeyID, err := sshKeyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Transfer SSH Key: %s", d.Id())
	_, err = conn.DeleteSshPublicKey(ctx, &transfer.DeleteSshPublicKeyInput{
		UserName:       aws.String(userName),
		ServerId:       aws.String(serverID),
		SshPublicKeyId: aws.String(sshKeyID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer SSH Key (%s): %s", d.Id(), err)
	}

	return diags
}

const sshKeyResourceIDSeparator = "/"

func sshKeyCreateResourceID(serverID, userName, sshKeyID string) string {
	parts := []string{serverID, userName, sshKeyID}
	id := strings.Join(parts, sshKeyResourceIDSeparator)

	return id
}

func sshKeyParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, sshKeyResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected SERVERID%[2]sUSERNAME%[2]sSSHKEYID", id, sshKeyResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findUserSSHKeyByThreePartKey(ctx context.Context, conn *transfer.Client, serverID, userName, sshKeyID string) (*awstypes.DescribedUser, *awstypes.SshPublicKey, error) {
	user, err := findUserByTwoPartKey(ctx, conn, serverID, userName)

	if err != nil {
		return nil, nil, err
	}

	sshKey, err := tfresource.AssertSingleValueResult(tfslices.Filter(user.SshPublicKeys, func(v awstypes.SshPublicKey) bool {
		return aws.ToString(v.SshPublicKeyId) == sshKeyID
	}))

	if err != nil {
		return nil, nil, err
	}

	if aws.ToString(sshKey.SshPublicKeyBody) == "" {
		return nil, nil, tfresource.NewEmptyResultError(nil)
	}

	return user, sshKey, nil
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
