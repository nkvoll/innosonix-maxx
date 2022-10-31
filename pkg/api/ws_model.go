package api

// WSLevelResponse models the messages received from the /socket/level API.
type WSLevelResponse struct {
	Level     [][]float64 `json:"level"`
	Reduction [][]float64 `json:"reduction"`
	PSU       float64     `json:"psu"`
	Thermo    float64     `json:"thermo"`
}

// WSDatapollResponse is a partial model of the messages received from the /socket/datapoll API.
type WSDatapollResponse struct {
	Settings *Settings `json:"settings"`
}
