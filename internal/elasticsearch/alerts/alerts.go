package elasticsearch_api_alerts

import "encoding/json"

type AlertAPISpec struct {
	Name     string           `json:"name"`
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
	Name            string       `json:"name"`
	Destination     string       `json:"destination_id"`
	SubjectTemplate TextTemplate `json:"subject_template"`
	MessageTemplate TextTemplate `json:"message_template"`
}

type TriggerThrottle struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type TextTemplate struct {
	Source string `json:"source"`
	Lang   string `json:"lang"`
}

type TriggerCondition struct {
	Script ConditionScript `json:"script"`
}

type ConditionScript struct {
	Source string `json:"source"`
	Lang   string `json:"lang"`
}

type MonitorInput struct {
	Search InputSearch `json:"search"`
}

type InputSearch struct {
	Indices []string        `json:"indices"`
	Query   json.RawMessage `json:"query"`
}

type MonitorSchedule struct {
	Period SchedulePeroid `json:"period"`
	// TODO: add other schedule types
}

type SchedulePeroid struct {
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}
