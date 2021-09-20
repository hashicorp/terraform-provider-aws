package resource

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func unmarshalJSON(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(v)
}
