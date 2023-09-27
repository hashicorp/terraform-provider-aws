// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandProtocol(l []interface{}) *datasync.FsxProtocol {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	protocol := &datasync.FsxProtocol{}

	if v, ok := m["nfs"].([]interface{}); ok {
		protocol.NFS = expandNFS(v)
	}
	if v, ok := m["smb"].([]interface{}); ok {
		protocol.SMB = expandSMB(v)
	}

	return protocol
}

func flattenProtocol(protocol *datasync.FsxProtocol, d *schema.ResourceData) []interface{} {
	if protocol == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if protocol.NFS != nil {
		m["nfs"] = flattenNFS(protocol.NFS)
	}
	if protocol.SMB != nil {
		m["smb"] = flattenSMB(protocol.SMB, d)
	}

	return []interface{}{m}
}

func expandNFS(l []interface{}) *datasync.FsxProtocolNfs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	protocol := &datasync.FsxProtocolNfs{
		MountOptions: expandNFSMountOptions(m["mount_options"].([]interface{})),
	}

	return protocol
}

func expandSMB(l []interface{}) *datasync.FsxProtocolSmb {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	protocol := &datasync.FsxProtocolSmb{
		Domain:       aws.String(m["domain"].(string)),
		MountOptions: expandSMBMountOptions(m["mount_options"].([]interface{})),
		Password:     aws.String(m["password"].(string)),
		User:         aws.String(m["user"].(string)),
	}

	return protocol
}

// todo: go another level down?
func flattenNFS(nfs *datasync.FsxProtocolNfs) []interface{} {
	if nfs == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"mount_options": flattenNFSMountOptions(nfs.MountOptions),
	}

	return []interface{}{m}
}

func flattenSMB(smb *datasync.FsxProtocolSmb, d *schema.ResourceData) []interface{} {
	if smb == nil {
		return []interface{}{}
	}

	// Need to store the value for "password" from config in the state since it is write-only in the Describe API
	password := ""
	if d != nil {
		password = d.Get("protocol.0.smb.0.password").(string)
	}

	m := map[string]interface{}{
		"domain":        smb.Domain,
		"mount_options": flattenSMBMountOptions(smb.MountOptions),
		"password":      password,
		"user":          smb.User,
	}

	return []interface{}{m}
}
