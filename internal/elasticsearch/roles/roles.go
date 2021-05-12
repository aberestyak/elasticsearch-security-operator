package esapiroles

import (
	"encoding/json"
)

// RoleAPISpec defines ES roles API
type RoleAPISpec struct {
	Description       string              `json:"description,omitempty"`
	ClusterPermissons []string            `json:"cluster_permissions,omitempty"`
	IndexPermissions  []IndexPermissions  `json:"index_permissions"`
	TenantPermissions []TenantPermissions `json:"tenant_permissions,omitempty"`
}

// IndexPermissions defines permissions to specified indices
type IndexPermissions struct {
	IndexPatterns  []string        `json:"index_patterns"`
	DLS            json.RawMessage `json:"dls,omitempty"`
	FLS            []string        `json:"fls,omitempty"`
	MaskedFields   []string        `json:"masked_fields,omitempty"`
	AllowedActions []string        `json:"allowed_actions"`
}

// TenantPermissions defines permissions to specified tenants
type TenantPermissions struct {
	TenantPatterns []string `json:"tenant_patterns"`
	AllowedActions []string `json:"allowed_actions"`
}
