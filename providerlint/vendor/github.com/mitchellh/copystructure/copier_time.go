package copystructure

import (
	"reflect"
	"time"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	Copiers[reflect.TypeOf(time.Time{})] = timeCopier
}

func timeCopier(v interface{}) (interface{}, error) {
	// Just... copy it.
	return v.(time.Time), nil
}
