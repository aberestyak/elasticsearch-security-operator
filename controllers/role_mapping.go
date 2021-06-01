package controllers

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/aberestyak/elasticsearch-security-operator/api/v1alpha1"
	"github.com/aberestyak/elasticsearch-security-operator/config"
	rolemappings "github.com/aberestyak/elasticsearch-security-operator/internal/elasticsearch/rolemappings"
	log "github.com/sirupsen/logrus"
)

var roleMappingLogger = log.WithFields(log.Fields{
	"component": "RoleMapping",
})

// CreateRoleMapping - create/update RoleMapping object, based on passed role
func CreateRoleMapping(role *v1alpha1.Role) error {
	roleMappingExists, existingRoleMappingSpec, err := GetExistingObject(config.AppConfig.ElasticsearchRoleMappingAPIPath, role.Name)
	if err != nil {
		return err
	}

	apiRoleMappingObject := MapAPIRoleMappingObject(role)
	apiRoleMappingJSON, _ := json.Marshal(apiRoleMappingObject)

	if !roleMappingExists {
		// Create new roleMapping
		if err := UpdateRoleMapping(role.Name, apiRoleMappingJSON); err != nil {
			return err
		}
		roleMappingLogger.Infof("Created roleMapping: %v.", role.Name)
	} else {
		// Update existing if need
		existingRoleMapping := make(map[string]rolemappings.RoleMappingAPISpec, 1)
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

// MapAPIRoleMappingObject - map passed role's RoleMapping field to roleMappings API
func MapAPIRoleMappingObject(role *v1alpha1.Role) *rolemappings.RoleMappingAPISpec {
	return &rolemappings.RoleMappingAPISpec{
		BackendRoles: role.Spec.RoleMappings.BackendRoles,
		Users:        role.Spec.RoleMappings.Users}
}

// UpdateRoleMapping - make request to create or update RoleMapping for "parent" role
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
