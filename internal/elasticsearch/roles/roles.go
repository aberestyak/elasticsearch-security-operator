package elasticsearch_api_roles

import (
	"encoding/json"
)

type RoleAPISpec struct {
	Description       string              `json:"description,omitempty"`
	ClusterPermissons []string            `json:"cluster_permissions"`
	IndexPermissions  []IndexPermissions  `json:"index_permissions"`
	TenantPermissions []TenantPermissions `json:"tenant_permissions,omitempty"`
}

type IndexPermissions struct {
	IndexPatterns  []string        `json:"index_patterns"`
	DLS            json.RawMessage `json:"dls,omitempty"`
	FLS            []string        `json:"fls,omitempty"`
	MaskedFields   []string        `json:"masked_fields,omitempty"`
	AllowedActions []string        `json:"allowed_actions"`
}

type TenantPermissions struct {
	TenantPatterns []string `json:"tenant_patterns"`
	AllowedActions []string `json:"allowed_actions"`
}
