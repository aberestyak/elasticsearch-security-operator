package controllers

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/aberestyak/elasticsearch-security-operator/api/v1alpha1"
	"github.com/aberestyak/elasticsearch-security-operator/config"
	log "github.com/sirupsen/logrus"
)

// RoleMappingAPISpec defines roleMapping API spec
type RoleMappingAPISpec struct {
	Users        []string `json:"users,omitempty"`
	BackendRoles []string `json:"backend_roles"`
}

var roleMappingLogger = log.WithFields(log.Fields{
	"component": "RoleMapping",
})

// MakeRoleMapping - create/update RoleMapping object, based on passed role
func MakeRoleMapping(role *v1alpha1.Role) error {
	roleMappingExists, existingRoleMappingSpec, err := GetExistingObject(config.AppConfig.ElasticsearchRoleMappingAPIPath, role.Name)
	if err != nil {
		return err
	}
	var apiRoleMappingObject RoleMappingAPISpec
	apiRoleMappingObject.MapAPIRoleMappingObject(role)
	apiRoleMappingJSON, _ := json.Marshal(apiRoleMappingObject)

	if !roleMappingExists {
		// Create new roleMapping
		if err := UpdateRoleMapping(role.Name, apiRoleMappingJSON); err != nil {
			return err
		}
		roleMappingLogger.Infof("Created roleMapping: %v.", role.Name)
	} else {
		// Update existing if need so
		existingRoleMapping := make(map[string]RoleMappingAPISpec, 1)
		if err := json.Unmarshal(existingRoleMappingSpec, &existingRoleMapping); err != nil {
			return err
		}
		if !reflect.DeepEqual(existingRoleMapping[role.Name], apiRoleMappingObject) {
			if err := UpdateRoleMapping(role.Name, apiRoleMappingJSON); err != nil {
				roleMappingLogger.Errorf("Error when updating roleMapping: %v", err.Error())
				return err
			}
			roleMappingLogger.Infof("Updated roleMapping: %v.", role.Name)
		}
	}
	return nil
}

// MapAPIRoleMappingObject - map passed role RoleMapping field to API
func (r *RoleMappingAPISpec) MapAPIRoleMappingObject(role *v1alpha1.Role) {
	r.BackendRoles = role.Spec.RoleMappings.BackendRoles
	r.Users = role.Spec.RoleMappings.Users
}

// UpdateRoleMapping - make request to create or update RoleMapping for "parent" role or user
func UpdateRoleMapping(name string, jsonRoleMapping []byte) error {
	_, responseResult, responseBody, err := MakeAPIRequest("PUT", config.AppConfig.ElasticsearchRoleMappingAPIPath+"/"+name, jsonRoleMapping)
	if err != nil {
		return errors.New("Error when creating new role:" + err.Error())
	}
	if responseResult != "Error" {
		return nil
	}
	return errors.New("Error when updating roleMapping" + name + "." + string(responseBody))
}
