/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleSpec defines the desired state of Role
type RoleSpec struct {
	ClusterPermissons []string           `json:"cluster_permissions"`
	IndexPermissions  []IndexPermissions `json:"index_permissions"`
	//+optional
	TenantPermissions []TenantPermissions `json:"tenant_permissions,omitempty"`
	RoleMappings      RoleMappings        `json:"roleMappings"`
}

type RoleMappings struct {
	BackendRoles []string `json:"backend_roles"`
	//+optional
	Users []string `json:"users,omitempty"`
}

type IndexPermissions struct {
	IndexPatterns []string `json:"index_patterns"`
	//+optional
	DLS string `json:"dls,omitempty"`
	//+optional
	FLS []string `json:"fls,omitempty"`
	//+optional
	MaskedFields   []string `json:"masked_fields,omitempty"`
	AllowedActions []string `json:"allowed_actions"`
}

type TenantPermissions struct {
	TenantPatterns []string `json:"tenant_patterns"`
	AllowedActions []string `json:"allowed_actions"`
}

// RoleStatus defines the observed state of Role
type RoleStatus struct {
	Status string `json:"state"`
	//+optional
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Role mappings",type=string,JSONPath=`.spec.roleMappings.backend_roles`
// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec,omitempty"`
	Status RoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Role{}, &RoleList{})
}
