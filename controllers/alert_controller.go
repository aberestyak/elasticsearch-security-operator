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

	config "github.com/aberestyak/elasticsearch-security-operator/config"
	alerts "github.com/aberestyak/elasticsearch-security-operator/internal/elasticsearch/alerts"
	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	securityv1alpha1 "github.com/aberestyak/elasticsearch-security-operator/api/v1alpha1"
)

const alertFinalizer = "alert.security.rshbdev.ru/finalizer"

var alertControllerLogger = log.WithFields(log.Fields{
	"component": "AlertController",
})

// AlertReconciler reconciles a Alert object
type AlertReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=alerts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=alerts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=security.rshbdev.ru,resources=alerts/finalizers,verbs=update

// Reconcile main reconcile loop
func (r *AlertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	desiredAlert := &securityv1alpha1.Alert{}
	var err = r.Get(ctx, req.NamespacedName, desiredAlert)
	if err != nil {
		if kerrors.IsNotFound(err) {
			alertControllerLogger.Info("Resource was deleted")
			return ctrl.Result{}, nil
		}
		alertControllerLogger.Errorf("Error while reading CR Alert: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Call finalyzer to clean up
	isdesiredAlertToBeDeleted := desiredAlert.GetDeletionTimestamp() != nil
	if isdesiredAlertToBeDeleted {
		if controllerutil.ContainsFinalizer(desiredAlert, alertFinalizer) {
			if err := r.FinalizeAlert(desiredAlert); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(desiredAlert, alertFinalizer)
			err := r.Update(ctx, desiredAlert)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(desiredAlert, alertFinalizer) {
		controllerutil.AddFinalizer(desiredAlert, alertFinalizer)
		err = r.Update(ctx, desiredAlert)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Map model to RoleAPISpec
	alertAPIObject, err := MapAlertAPIObject(desiredAlert)
	if err != nil {
		roleControllerLogger.Errorf("Error when mapping models : %v", err)
		return ctrl.Result{}, err
	}

	// Sanitize query json's
	for i, query := range alertAPIObject.Inputs {
		alertAPIObject.Inputs[i].Search.Query = SanitizeQuery(query.Search.Query)
	}

	jsonAlert, err := json.Marshal(alertAPIObject)
	if err != nil {
		alertControllerLogger.Errorf("Error when marhaling alert object: %v", err)
		return ctrl.Result{}, err
	}
	// New object created
	if desiredAlert.Status.Monitor.ID == "" {
		alertID, responseResult, responseBody, err := MakeAPIRequest("POST", config.AppConfig.ElasticsearchAlertAPIPath, jsonAlert)
		if err != nil {
			alertControllerLogger.Errorf("Error when creating new alert: %v", err.Error())
			return ctrl.Result{}, err
		}
		if err := SetAlertStatus(r, desiredAlert, responseResult, responseBody, alertID); err != nil {
			return ctrl.Result{}, err
		}
		alertControllerLogger.Infof("Created new alert: %v. Status: %v", desiredAlert.Name, desiredAlert.Status.Monitor.Status)

		// Modified existing object
	} else {
		// Don't update new and deleted alerts
		if desiredAlert.Generation > 1 && !isdesiredAlertToBeDeleted {
			alertID, responseResult, responseBody, err := MakeAPIRequest("PUT", config.AppConfig.ElasticsearchAlertAPIPath+"/"+desiredAlert.Status.Monitor.ID, jsonAlert)
			if err != nil {
				alertControllerLogger.Errorf("Error when updating alert: %v", err.Error())
				return ctrl.Result{}, err
			}
			if err := SetAlertStatus(r, desiredAlert, responseResult, responseBody, alertID); err != nil {
				return ctrl.Result{}, err
			}
			alertControllerLogger.Infof("Updated alert: %v. Status: %v", desiredAlert.Name, desiredAlert.Status.Monitor.Status)
		}
	}
	return ctrl.Result{}, nil
}

// MapAlertAPIObject - map CRD model to API
func MapAlertAPIObject(alert *securityv1alpha1.Alert) (*alerts.AlertAPISpec, error) {
	var alertAPI alerts.AlertAPISpec
	buf, _ := json.Marshal(alert.Spec)
	if err := json.Unmarshal(buf, &alertAPI); err != nil {
		return nil, err
	}
	return &alertAPI, nil
}

// SetAlertStatus set status
func SetAlertStatus(r *AlertReconciler, alert *securityv1alpha1.Alert, responseResult string, responseBody []byte, alertID string) error {
	alert.Status.Monitor = securityv1alpha1.StatusMonitor{
		Name:   alert.Name,
		ID:     alertID,
		Status: responseResult,
		Error: func(response string, responseBody []byte) string {
			if response == "Error" {
				return string(responseBody)
			}
			return ""
		}(responseResult, responseBody),
	}
	if err := r.Client.Status().Update(context.TODO(), alert); err != nil {
		return errors.New("Error when updating alert status: " + err.Error())
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Alert{}).
		Complete(r)
}

// FinalizeAlert delete alert
func (r *AlertReconciler) FinalizeAlert(alert *securityv1alpha1.Alert) error {
	if alert.Status.Monitor.ID != "" {
		_, _, _, err := MakeAPIRequest("DELETE", config.AppConfig.ElasticsearchAlertAPIPath+"/"+alert.Status.Monitor.ID, nil)
		if err != nil {
			alertControllerLogger.Errorf("Error when finalyzing alert: %v", err.Error())
			return err
		}
	}
	alertControllerLogger.Infof("Successfully finalized alert: %v", alert.Name)
	return nil
}
