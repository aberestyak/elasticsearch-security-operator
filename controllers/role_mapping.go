package controllers

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/aberestyak/elasticsearch-security-operator/api/v1alpha1"
	"github.com/aberestyak/elasticsearch-security-operator/config"
	log "github.com/sirupsen/logrus"
)

type RoleMappingAPISpec struct {
	Users        []string `json:"users,omitempty"`
	BackendRoles []string `json:"backend_roles"`
}

var roleMappingLogger = log.WithFields(log.Fields{
	"component": "RoleMapping",
})

// MakeRoleMapping - create/update RoleMapping object, based on passed role
func MakeRoleMapping(role *v1alpha1.Role) error {
	roleMappingExists, existingRoleMappingSpec, err := GetExistingObject(config.AppConfig.ElasticsearchRoleMappingApiPath, role.Name)
	if err != nil {
		return err
	}
	var apiRoleMappingObject RoleMappingAPISpec
	apiRoleMappingObject.MapAPIRoleMappingObject(role)
	apiRoleMappingJson, _ := json.Marshal(apiRoleMappingObject)

	if !roleMappingExists {
		// Create new roleMapping
		if err := UpdateRoleMapping(role.Name, apiRoleMappingJson); err != nil {
			return err
		}
		roleMappingLogger.Infof("Created roleMapping: %v.", role.Name)
	} else {
		// Update existing if need so
		existingRoleMapping := make(map[string]RoleMappingAPISpec, 1)
		json.Unmarshal(existingRoleMappingSpec, &existingRoleMapping)
		if !reflect.DeepEqual(existingRoleMapping[role.Name], apiRoleMappingObject) {
			if err := UpdateRoleMapping(role.Name, apiRoleMappingJson); err != nil {
				roleMappingLogger.Errorf("Error when updating roleMapping: %v", err.Error())
				return err
			}
			roleMappingLogger.Infof("Updated roleMapping: %v.", role.Name)
		}
	}
	return nil
}

// MakeRoleMapping - map passed role RoleMapping field to API
func (r *RoleMappingAPISpec) MapAPIRoleMappingObject(role *v1alpha1.Role) {
	r.BackendRoles = role.Spec.RoleMappings.BackendRoles
	r.Users = role.Spec.RoleMappings.Users
}

// UpdateRoleMapping - make request to create or update RoleMapping for "parent" role or user
func UpdateRoleMapping(name string, jsonRoleMapping []byte) error {
	_, responseResult, responseBody, err := MakeAPIRequest("PUT", config.AppConfig.ElasticsearchRoleMappingApiPath+"/"+name, jsonRoleMapping)
	if err != nil {
		return errors.New("Error when creating new role:" + err.Error())
	}
	if responseResult != "Error" {
		return nil
	} else {
		return errors.New("Error when updating roleMapping" + name + "." + string(responseBody))
	}

}
