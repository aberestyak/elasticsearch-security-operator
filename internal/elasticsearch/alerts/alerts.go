package esapialerts

import "encoding/json"

// AlertAPISpec defines ES alerts API
type AlertAPISpec struct {
	Name     string           `json:"name"`
	Type     string           `json:"type"`
	Enabled  bool             `json:"enabled"`
	Schedule MonitorSchedule  `json:"schedule"`
	Inputs   []MonitorInput   `json:"inputs"`
	Triggers []MonitorTrigger `json:"triggers"`
}

// MonitorTrigger defines triggers and required actions
type MonitorTrigger struct {
	Name      string           `json:"name"`
	Severity  string           `json:"severity"`
	Condition TriggerCondition `json:"condition"`
	Actions   []TriggerAction  `json:"actions"`
}

// TriggerAction defines alerting destination and templates
type TriggerAction struct {
	Name            string       `json:"name"`
	Destination     string       `json:"destination_id"`
	SubjectTemplate TextTemplate `json:"subject_template"`
	MessageTemplate TextTemplate `json:"message_template"`
}

// TriggerThrottle defines alerting throttle
type TriggerThrottle struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

// TextTemplate defines alert text template
type TextTemplate struct {
	Source string `json:"source"`
	Lang   string `json:"lang"`
}

// TriggerCondition defines condition to trigger alert
type TriggerCondition struct {
	Script ConditionScript `json:"script"`
}

// ConditionScript defines language and script to execute
type ConditionScript struct {
	Source string `json:"source"`
	Lang   string `json:"lang"`
}

// MonitorInput defines search queries
type MonitorInput struct {
	Search InputSearch `json:"search"`
}

// InputSearch defines search queries and indices
type InputSearch struct {
	Indices []string        `json:"indices"`
	Query   json.RawMessage `json:"query"`
}

// MonitorSchedule defines schedule period
type MonitorSchedule struct {
	Period SchedulePeroid `json:"period"`
	// TODO: add other schedule types
}

// SchedulePeroid defines schedule time period
type SchedulePeroid struct {
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}
