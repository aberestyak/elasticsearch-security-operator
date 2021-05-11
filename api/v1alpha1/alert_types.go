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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlertSpec defines the desired state of Alert
type AlertSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name string `json:"name"`
	//+kubebuilder:default:monitor
	Type     string           `json:"type"`
	Enabled  bool             `json:"enabled"`
	Schedule MonitorSchedule  `json:"schedule"`
	Inputs   []MonitorInput   `json:"inputs"`
	Triggers []MonitorTrigger `json:"triggers"`
}

type MonitorTrigger struct {
	Name      string           `json:"name"`
	Severity  string           `json:"severity"`
	Condition TriggerCondition `json:"condition"`
	Actions   []TriggerAction  `json:"actions"`
}

type TriggerAction struct {
	Name string `json:"name"`
	//+optional
	Destination     string       `json:"destination_id"`
	SubjectTemplate TextTemplate `json:"subject_template"`
	MessageTemplate TextTemplate `json:"message_template"`
	// TODO:
	//+optional
	//ThrottleEnabled bool `json:"throttle_enabled,omitempty"`
	//+optional
	//Throttle TriggerThrottle `json:"throttle,omitempty"`
}

type TriggerThrottle struct {
	//+kubebuilder:default:=1
	Value int `json:"value"`
	//+kubebuilder:default:=MINUTES
	//+kubebuilder:validation:Enum=HOURS;MINUTES;DAYS
	Unit string `json:"unit"`
}

type TextTemplate struct {
	Source string `json:"source"`
	//+kubebuilder:validation:Enum=mustache;painless
	Lang string `json:"lang"`
}

type TriggerCondition struct {
	Script ConditionScript `json:"script"`
}

type ConditionScript struct {
	Source string `json:"source"`
	//+kubebuilder:validation:Enum=painless
	Lang string `json:"lang"`
}

type MonitorInput struct {
	Search InputSearch `json:"search"`
}

type InputSearch struct {
	Indices []string `json:"indices"`
	Query   string   `json:"query"`
}

type MonitorSchedule struct {
	Period SchedulePeroid `json:"period"`
	// TODO: add other schedule types
}

type SchedulePeroid struct {
	Interval int `json:"interval"`
	//+kubebuilder:validation:Enum=HOURS;MINUTES;DAYS
	Unit string `json:"unit"`
}

// AlertStatus defines the observed state of Alert
type AlertStatus struct {
	Monitor StatusMonitor `json:"monitor"`
}
type StatusMonitor struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status string `json:"state"`
	//+optional
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Enabled",type=boolean,JSONPath=`.spec.enabled`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.monitor.state`

// Alert is the Schema for the alerts API
type Alert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertSpec   `json:"spec,omitempty"`
	Status AlertStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AlertList contains a list of Alert
type AlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Alert{}, &AlertList{})
}
