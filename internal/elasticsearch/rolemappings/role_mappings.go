package esapirolemapping

// RoleMappingAPISpec defines roleMapping API spec
type RoleMappingAPISpec struct {
	Users        []string `json:"users,omitempty"`
	BackendRoles []string `json:"backend_roles,omitempty"`
}
