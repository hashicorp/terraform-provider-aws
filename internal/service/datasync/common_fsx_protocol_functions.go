// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"github.com/aws/aws-sdk-go/service/datasync"
)

func expandProtocol(l []interface{}) *datasync.FsxProtocol {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	Protocol := &datasync.FsxProtocol{
		NFS: expandNFS(m["nfs"].([]interface{})),
		SMB: expandSMB(m["smb"].([]interface{})),
	}

	return Protocol
}

func flattenProtocol(protocol *datasync.FsxProtocol) []interface{} {
	if protocol == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"nfs": flattenNFS(protocol.NFS),
		"smb": flattenSMB(protocol.SMB),
	}

	return []interface{}{m}
}

func expandNFS(l []interface{}) *datasync.FsxProtocolNfs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	Protocol := &datasync.FsxProtocolNfs{
		MountOptions: expandNFSMountOptions(m["mount_options"].([]interface{})),
	}

	return Protocol
}

func expandSMB(l []interface{}) *datasync.FsxProtocolSmb {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	Protocol := &datasync.FsxProtocolSmb{
		MountOptions: expandSMBMountOptions(m["mount_options"].([]interface{})),
	}

	return Protocol
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

	return []interface{}{m}
}
