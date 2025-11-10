// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"reflect"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
)

func Test_replicationGroupStateUpgradeFromV1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawState map[string]any
		want     map[string]any
		wantErr  bool
	}{
		{
			name:     "empty rawState",
			rawState: nil,
			want: map[string]any{
				"auth_token_update_strategy": awstypes.AuthTokenUpdateStrategyTypeRotate,
			},
		},
		{
			name: "auth_token",
			rawState: map[string]any{
				"auth_token": "foo",
			},
			want: map[string]any{
				"auth_token":                 "foo",
				"auth_token_update_strategy": awstypes.AuthTokenUpdateStrategyTypeRotate,
			},
		},
		{
			name: "cluster_mode block",
			rawState: map[string]any{
				"cluster_mode.num_node_groups":         "2",
				"cluster_mode.replicas_per_node_group": "2",
			},
			want: map[string]any{
				"auth_token_update_strategy": awstypes.AuthTokenUpdateStrategyTypeRotate,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := replicationGroupStateUpgradeFromV1(t.Context(), tt.rawState, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("replicationGroupStateUpgradeFromV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replicationGroupStateUpgradeFromV1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replicationGroupStateUpgradeFromV2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawState map[string]any
		want     map[string]any
		wantErr  bool
	}{
		{
			name:     "empty rawState",
			rawState: nil,
			want:     map[string]any{},
		},
		{
			name: "no auth_token, default auth_token_update_strategy",
			rawState: map[string]any{
				"auth_token_update_strategy": string(awstypes.AuthTokenUpdateStrategyTypeRotate),
			},
			want: map[string]any{},
		},
		{
			name: "auth_token and auth_token_update_strategy",
			rawState: map[string]any{
				"auth_token":                 "foo",
				"auth_token_update_strategy": string(awstypes.AuthTokenUpdateStrategyTypeSet),
			},
			want: map[string]any{
				"auth_token":                 "foo",
				"auth_token_update_strategy": string(awstypes.AuthTokenUpdateStrategyTypeSet),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := replicationGroupStateUpgradeFromV2(t.Context(), tt.rawState, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("replicationGroupStateUpgradeFromV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replicationGroupStateUpgradeFromV2() = %v, want %v", got, tt.want)
			}
		})
	}
}
