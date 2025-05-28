package redis

import "github.com/victorximenis/multitenant/core"

// Compile-time check to ensure TenantCache implements core.TenantCache interface
var _ core.TenantCache = (*TenantCache)(nil)
