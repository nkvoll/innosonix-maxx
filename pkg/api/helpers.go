package api

import (
	"encoding/json"
	"time"
)

type structWithValueField struct {
	Value string `json:"value"`
}

const format = "2006-01-02T15:04:05+0200"

// Work around SettingsTimeCurrent not unmarshalling correctly
// err msg: parsing time "\"2022-10-21T19:38:28+0200\"" as "\"2006-01-02T15:04:05Z07:00\"": cannot parse "+0200\"" as "Z07:00"
func (t *SettingsTimeCurrent) UnmarshalJSON(data []byte) error {
	tme, err := time.Parse(format, string(data))
	if err != nil {
		t.Value = &tme
		return nil
	}

	h := structWithValueField{}

	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	tme, err = time.Parse(format, h.Value)
	if err != nil {
		return err
	}

	t.Value = &tme

	return nil
}

// Work around SettingsTimezone differences between the settings api and the datapoll settings equivalent
func (t *SettingsTimezone) UnmarshalJSON(data []byte) error {
	h := structWithValueField{}

	// not wrapped in an object
	if len(data) > 0 && data[0] != '{' {
		// in datapoll, the timezone is not wrapped
		h.Value = string(data)
		return nil
	}

	// wrapped in a object with a Value field
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	t.Value = &h.Value

	return nil
}
