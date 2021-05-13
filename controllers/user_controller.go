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

	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	securityv1alpha1 "github.com/aberestyak/elasticsearch-security-operator/api/v1alpha1"
	config "github.com/aberestyak/elasticsearch-security-operator/config"
	users "github.com/aberestyak/elasticsearch-security-operator/internal/elasticsearch/users"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var userControllerLogger = log.WithFields(log.Fields{
	"component": "UserController",
})

const userFinalizer = "user.security.rshbdev.ru/finalizer"

//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=users/finalizers,verbs=update

// Reconcile main reconcile loop
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	desiredUser := &securityv1alpha1.User{}
	var err = r.Get(ctx, req.NamespacedName, desiredUser)
	if err != nil {
		if kerrors.IsNotFound(err) {
			userControllerLogger.Info("Resource was deleted")
			return ctrl.Result{}, nil
		}
		userControllerLogger.Errorf("Error while reading CR User: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Call finalyzer to clean up
	isdesiredUserToBeDeleted := desiredUser.GetDeletionTimestamp() != nil
	if isdesiredUserToBeDeleted {
		if controllerutil.ContainsFinalizer(desiredUser, userFinalizer) {
			if err := r.FinalizeUser(desiredUser); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(desiredUser, userFinalizer)
			err := r.Update(ctx, desiredUser)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(desiredUser, userFinalizer) {
		controllerutil.AddFinalizer(desiredUser, userFinalizer)
		err = r.Update(ctx, desiredUser)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// Map model to UserAPISpec
	userAPIObject, err := MapUserAPIObject(desiredUser)
	if err != nil {
		roleControllerLogger.Errorf("Error when mapping models : %v", err)
	}
	apiUserJSON, err := json.Marshal(userAPIObject)
	if err != nil {
		roleControllerLogger.Errorf("Error when marshaling user object: %v", err)
		return ctrl.Result{}, err
	}

	// Can't get hash from elasticsearch, so can't check changed or not
	if err := CreateOrUpdateUser(r, desiredUser, apiUserJSON); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// CreateOrUpdateUser - make PUT request to create or update User
func CreateOrUpdateUser(r *UserReconciler, user *securityv1alpha1.User, jsonUser []byte) error {
	_, responseResult, responseBody, err := MakeAPIRequest("PUT", config.AppConfig.ElasticsearchUserAPIPath+"/"+user.Name, jsonUser)
	if err != nil {
		return errors.New("Error when creating new user: " + err.Error())
	}
	if err := SetUserStatus(r, user, responseResult, responseBody); err != nil {
		return err
	}
	userControllerLogger.Infof("Updated user: %v", user.Name)
	return nil
}

// SetUserStatus - parse http response code, set status and update CR
func SetUserStatus(r *UserReconciler, user *securityv1alpha1.User, responseResult string, responseBody []byte) error {
	user.Status = securityv1alpha1.UserStatus{
		Status: responseResult,
		Error: func(response string, responseBody []byte) string {
			if response == "Error" {
				return string(responseBody)
			}
			return ""
		}(responseResult, responseBody),
	}
	if err := r.Client.Status().Update(context.TODO(), user); err != nil {
		return errors.New("Error when setting status: " + err.Error())
	}
	return nil
}

// MapUserAPIObject - map CRD model to API
func MapUserAPIObject(user *securityv1alpha1.User) (*users.UserAPISpec, error) {
	var userAPI users.UserAPISpec
	buf, _ := json.Marshal(user.Spec)
	if err := json.Unmarshal(buf, &userAPI); err != nil {
		return nil, err
	}
	return &userAPI, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.User{}).
		Complete(r)
}

// FinalizeUser delete user
func (r *UserReconciler) FinalizeUser(user *securityv1alpha1.User) error {
	_, _, _, err := MakeAPIRequest("DELETE", config.AppConfig.ElasticsearchUserAPIPath+"/"+user.Name, nil)
	if err != nil {
		userControllerLogger.Errorf("Error when finalyzing user: %v", err.Error())
		return err
	}
	userControllerLogger.Infof("Successfully finalized user: %v", user.Name)
	return nil
}
