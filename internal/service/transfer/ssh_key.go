package transfer

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceSSHKey() *schema.Resource {

	return &schema.Resource{
		Create: resourceSSHKeyCreate,
		Read:   resourceSSHKeyRead,
		Delete: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserName,
			},
		},
	}
}

func resourceSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	userName := d.Get("user_name").(string)
	serverID := d.Get("server_id").(string)

	createOpts := &transfer.ImportSshPublicKeyInput{
		ServerId:         aws.String(serverID),
		UserName:         aws.String(userName),
		SshPublicKeyBody: aws.String(d.Get("body").(string)),
	}

	log.Printf("[DEBUG] Create Transfer SSH Public Key Option: %#v", createOpts)

	resp, err := conn.ImportSshPublicKey(createOpts)
	if err != nil {
		return fmt.Errorf("Error importing ssh public key: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", serverID, userName, *resp.SshPublicKeyId))

	return nil
}

func resourceSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	serverID, userName, sshKeyID, err := DecodeSSHKeyID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing Transfer SSH Public Key ID: %s", err)
	}

	descOpts := &transfer.DescribeUserInput{
		UserName: aws.String(userName),
		ServerId: aws.String(serverID),
	}

	log.Printf("[DEBUG] Describe Transfer User Option: %#v", descOpts)

	resp, err := conn.DescribeUser(descOpts)
	if err != nil {
		if tfawserr.ErrMessageContains(err, transfer.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Transfer User (%s) for Server (%s) not found, removing ssh public key (%s) from state", userName, serverID, sshKeyID)
			d.SetId("")
			return nil
		}
		return err
	}

	var body string
	for _, s := range resp.User.SshPublicKeys {
		if sshKeyID == aws.StringValue(s.SshPublicKeyId) {
			body = aws.StringValue(s.SshPublicKeyBody)
		}
	}

	if body == "" {
		log.Printf("[WARN] No such ssh public key found for User (%s) in Server (%s)", userName, serverID)
		d.SetId("")
	}

	d.Set("server_id", resp.ServerId)
	d.Set("user_name", resp.User.UserName)
	d.Set("body", body)

	return nil
}

func resourceSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	serverID, userName, sshKeyID, err := DecodeSSHKeyID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing Transfer SSH Public Key ID: %s", err)
	}

	delOpts := &transfer.DeleteSshPublicKeyInput{
		UserName:       aws.String(userName),
		ServerId:       aws.String(serverID),
		SshPublicKeyId: aws.String(sshKeyID),
	}

	log.Printf("[DEBUG] Delete Transfer SSH Public Key Option: %#v", delOpts)

	_, err = conn.DeleteSshPublicKey(delOpts)
	if err != nil {
		if tfawserr.ErrMessageContains(err, transfer.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Transfer User Ssh Key (%s): %s", d.Id(), err)
	}

	return nil
}

func DecodeSSHKeyID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected SERVERID/USERNAME/SSHKEYID", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
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
