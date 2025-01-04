package data

type UIConfig struct {
	UID       string     `json:"uid"`
	Functions []Function `json:"functions"`
}

type Function struct {
	UID          string      `json:"uid"`
	ChannelType  string      `json:"channelType"`
	DisplayName  string      `json:"displayName"`
	FunctionType string      `json:"functionType"`
	DataPoints   []DataPoint `json:"dataPoints"`
}

type DataPoint struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

func (cfg UIConfig) Find(uid string) (Function, bool) {
	for _, fn := range cfg.Functions {
		if fn.UID == uid {
			return fn, true
		}
	}
	return Function{}, false
}
