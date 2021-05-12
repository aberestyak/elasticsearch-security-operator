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

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	securityv1alpha1 "github.com/aberestyak/elasticsearch-security-operator/api/v1alpha1"
	config "github.com/aberestyak/elasticsearch-security-operator/config"
	roles "github.com/aberestyak/elasticsearch-security-operator/internal/elasticsearch/roles"
)

const roleFinalizer = "role.security.rshbdev.ru/finalizer"

var roleControllerLogger = log.WithFields(log.Fields{
	"component": "RoleController",
})

// RoleReconciler reconciles a Role object
type RoleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=roles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=roles/finalizers,verbs=update

// Reconcile main reconcile loop
func (r *RoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	desiredRole := &securityv1alpha1.Role{}
	var err = r.Get(ctx, req.NamespacedName, desiredRole)
	if err != nil {
		if kerrors.IsNotFound(err) {
			roleControllerLogger.Info("Resource was deleted")
			return ctrl.Result{}, nil
		}
		roleControllerLogger.Errorf("Error while reading CR Role: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Call finalyzer to clean up
	isdesiredRoleToBeDeleted := desiredRole.GetDeletionTimestamp() != nil
	if isdesiredRoleToBeDeleted {
		if controllerutil.ContainsFinalizer(desiredRole, roleFinalizer) {
			if err := r.FinalizeRole(desiredRole); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(desiredRole, roleFinalizer)
			err := r.Update(ctx, desiredRole)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(desiredRole, roleFinalizer) {
		controllerutil.AddFinalizer(desiredRole, roleFinalizer)
		err = r.Update(ctx, desiredRole)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Map model to RoleAPISpec
	roleAPIObject, err := MapRoleAPIObject(desiredRole)
	if err != nil {
		roleControllerLogger.Errorf("Error when mapping models : %v", err)
		return ctrl.Result{}, err
	}

	apiRoleJSON, err := json.Marshal(roleAPIObject)
	if err != nil {
		roleControllerLogger.Errorf("Error when marshaling role object: %v", err)
		return ctrl.Result{}, err
	}
	roleExists, existingRoleSpec, err := GetExistingObject(config.AppConfig.ElasticsearchRoleAPIPath, desiredRole.Name)
	if err != nil {
		roleControllerLogger.Errorf("Error when checking role existence: %v", err.Error())
		return ctrl.Result{}, err
	}
	if !roleExists {
		// Create
		if err := CreateOrUpdateRole(r, desiredRole, apiRoleJSON); err != nil {
			roleControllerLogger.Errorf("Error when creating new role: %v", err.Error())
		}
		roleControllerLogger.Infof("Created new role: %v. Status: %v", desiredRole.Name, desiredRole.Status.Status)
	} else {
		// Update
		// Trying to get existing role
		existingRole := make(map[string]roles.RoleAPISpec, 1)
		if err := json.Unmarshal(existingRoleSpec, &existingRole); err != nil {
			roleControllerLogger.Errorf("Error when unmarshaling existing role: %v", err.Error())
			return ctrl.Result{}, err
		}
		// Compare existing and desired role spec
		if !reflect.DeepEqual(existingRole[desiredRole.Name], roleAPIObject) {
			if err := CreateOrUpdateRole(r, desiredRole, apiRoleJSON); err != nil {
				roleControllerLogger.Errorf("Error when updating role: %v", err.Error())
			}
			roleControllerLogger.Infof("Updated role: %v. Status: %v", desiredRole.Name, desiredRole.Status.Status)
		}
		// Create or update roleMapping, no matter is this create or update operation and update role status
		if err := MakeRoleMapping(desiredRole); err != nil {
			if err := SetRoleStatus(r, desiredRole, "Error", []byte(err.Error())); err != nil {
				roleControllerLogger.Errorf("Error when setting role status: %v", err.Error())
			}
		}
	}
	return ctrl.Result{}, nil
}

// MapRoleAPIObject - map CRD model to API
func MapRoleAPIObject(role *securityv1alpha1.Role) (*roles.RoleAPISpec, error) {
	var roleAPI roles.RoleAPISpec
	buf, _ := json.Marshal(role.Spec)
	if err := json.Unmarshal(buf, &roleAPI); err != nil {
		return nil, err
	}
	return &roleAPI, nil
}

// SetRoleStatus - parse http response code, set status and update CR
func SetRoleStatus(r *RoleReconciler, role *securityv1alpha1.Role, responseResult string, responseBody []byte) error {
	role.Status = securityv1alpha1.RoleStatus{
		Status: responseResult,
		Error: func(response string, responseBody []byte) string {
			if response == "Error" {
				return string(responseBody)
			}
			return ""
		}(responseResult, responseBody),
	}
	if err := r.Client.Status().Update(context.TODO(), role); err != nil {
		return errors.New("Error when setting status: " + err.Error())
	}
	return nil
}

// CreateOrUpdateRole - make PUT request to create or update Role
func CreateOrUpdateRole(r *RoleReconciler, role *securityv1alpha1.Role, jsonRole []byte) error {
	_, responseResult, responseBody, err := MakeAPIRequest("PUT", config.AppConfig.ElasticsearchRoleAPIPath+"/"+role.Name, jsonRole)
	if err != nil {
		return errors.New("Error when creating new role: " + err.Error())
	}
	if err := SetRoleStatus(r, role, responseResult, responseBody); err != nil {
		return err
	}
	roleControllerLogger.Infof("Updated role: %v", role.Name)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Role{}).
		Complete(r)
}

// FinalizeRole delete role
func (r *RoleReconciler) FinalizeRole(role *securityv1alpha1.Role) error {
	jsonRole, _ := json.Marshal(role.Spec)
	_, _, _, err := MakeAPIRequest("DELETE", config.AppConfig.ElasticsearchRoleAPIPath+"/"+role.Name, jsonRole)
	if err != nil {
		roleControllerLogger.Errorf("Error when finalyzing role: %v", err.Error())
		return err
	}
	roleControllerLogger.Infof("Successfully finalized role: %v", role.Name)
	return nil
}
