package provider

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resources map[string]*schema.Resource
var dataSources map[string]*schema.Resource
var resourcesMu sync.Mutex
var registrationComplete bool

func RegisterResource(name string, r *schema.Resource) {
	resourcesMu.Lock()
	defer resourcesMu.Unlock()

	if resources == nil {
		resources = map[string]*schema.Resource{}
	}
	resources[name] = r
}

func RegisterDataSource(name string, r *schema.Resource) {
	resourcesMu.Lock()
	defer resourcesMu.Unlock()

	if dataSources == nil {
		dataSources = map[string]*schema.Resource{}
	}
	dataSources[name] = r
}

func RegistrationComplete() {
	registrationComplete = true
}
