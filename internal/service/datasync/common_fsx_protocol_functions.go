// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
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

func flattenProtocol(protocol *datasync.FsxProtocol) []interface{} {
	if protocol == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if protocol.NFS != nil {
		m["nfs"] = flattenNFS(protocol.NFS)
	}
	if protocol.SMB != nil {
		m["smb"] = flattenSMB(protocol.SMB)
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
		MountOptions: expandSMBMountOptions(m["mount_options"].([]interface{})),
	}
	if v, ok := m["domain"].(string); ok && v != "" {
		protocol.Domain = aws.String(v)
	}
	if v, ok := m["password"].(string); ok && v != "" {
		protocol.Password = aws.String(v)
	}
	if v, ok := m["user"].(string); ok && v != "" {
		protocol.User = aws.String(v)
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

func flattenSMB(smb *datasync.FsxProtocolSmb) []interface{} {
	if smb == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"mount_options": flattenSMBMountOptions(smb.MountOptions),
	}
	if v := smb.Domain; v != nil {
		m["domain"] = aws.StringValue(v)
	}
	if v := smb.Password; v != nil {
		m["password"] = aws.StringValue(v)
	}
	if v := smb.User; v != nil {
		m["user"] = aws.StringValue(v)
	}

	return []interface{}{m}
}
