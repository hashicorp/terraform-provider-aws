// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package privatestate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func NewWriteOnlyValueStore(private PrivateState, key string) *WriteOnlyValueStore {
	return &WriteOnlyValueStore{
		key:     key,
		private: private,
	}
}

type WriteOnlyValueStore struct {
	key     string
	private PrivateState
}

func (w *WriteOnlyValueStore) EqualValue(ctx context.Context, value types.String) (bool, diag.Diagnostics) {
	bytes, diags := w.private.GetKey(ctx, w.key)
	if diags.HasError() {
		return false, diags
	}

	var s string
	if err := tfjson.DecodeFromBytes(bytes, &s); err != nil {
		diags.AddError("decoding private state", err.Error())
		return false, diags
	}

	return s == sha256Hash(value.ValueString()), diags
}

func (w *WriteOnlyValueStore) HasValue(ctx context.Context) (bool, diag.Diagnostics) {
	bytes, diags := w.private.GetKey(ctx, w.key)
	return len(bytes) > 0, diags
}

func (w *WriteOnlyValueStore) SetValue(ctx context.Context, val types.String) diag.Diagnostics {
	if val.IsNull() {
		return w.private.SetKey(ctx, w.key, []byte(""))
	}

	return w.private.SetKey(ctx, w.key, []byte(strconv.Quote(sha256Hash(val.ValueString()))))
}

func sha256Hash(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}
