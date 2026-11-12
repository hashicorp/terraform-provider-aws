// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"reflect"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
)

func TestExpandProtocolDetailsProxyConfig(t *testing.T) {
	t.Parallel()

	got := expandProtocolDetails([]any{map[string]any{
		"proxy_config": []any{map[string]any{
			"sftp_mode": string(awstypes.ProxyModeProxyProtocolV2Enforced),
		}},
	}})

	if got == nil || got.ProxyConfig == nil {
		t.Fatal("expected ProxyConfig")
	}

	if got.ProxyConfig.SftpMode != awstypes.ProxyModeProxyProtocolV2Enforced {
		t.Errorf("SftpMode = %q, want %q", got.ProxyConfig.SftpMode, awstypes.ProxyModeProxyProtocolV2Enforced)
	}
}

func TestFlattenProtocolDetailsProxyConfig(t *testing.T) {
	t.Parallel()

	got := flattenProtocolDetails(&awstypes.ProtocolDetails{
		ProxyConfig: &awstypes.ProxyConfig{
			SftpMode: awstypes.ProxyModeProxyProtocolV2Enforced,
		},
	})
	want := []any{map[string]any{
		"proxy_config": []any{map[string]any{
			"sftp_mode": awstypes.ProxyModeProxyProtocolV2Enforced,
		}},
		"set_stat_option":             awstypes.SetStatOption(""),
		"tls_session_resumption_mode": awstypes.TlsSessionResumptionMode(""),
	}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("flattenProtocolDetails() = %#v, want %#v", got, want)
	}
}
