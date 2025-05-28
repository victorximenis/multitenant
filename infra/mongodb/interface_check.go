package mongodb

import "github.com/victorximenis/multitenant/core"

// Compile-time check to ensure TenantRepository implements core.TenantRepository interface
var _ core.TenantRepository = (*TenantRepository)(nil)
