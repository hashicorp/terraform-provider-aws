package elasticache

import (
	"context"
	"fmt"

	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var minMemcachedTransitEncryptionVersion = gversion.Must(gversion.NewVersion("1.6.12"))

func CustomizeDiffValidateTransitEncryptionEnabled(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	engine := diff.Get("engine").(string)

	transitEncryptionEnabled, ok := diff.GetOk("transit_encryption_enabled")
	if !ok || !transitEncryptionEnabled.(bool) {
		return nil
	}

	if engine == engineRedis {
		return fmt.Errorf("aws_elasticache_cluster does not support transit encryption using the redis engine, use aws_elasticache_replication_group instead")
	}

	engineVersion, ok := diff.GetOk("engine_version")
	if !ok {
		return nil
	}

	version, err := normalizeEngineVersion(engineVersion.(string))
	if err != nil {
		return err
	}

	if version.LessThan(minMemcachedTransitEncryptionVersion) {
		return fmt.Errorf("Transit encryption is not supported for memcached version %v", version)
	}
	return nil
}
