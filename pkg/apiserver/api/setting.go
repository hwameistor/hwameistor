package api

type DrbdEnableSetting struct {
	Enable  bool   `json:"enable"`
	State   State  `json:"state"`
	Version string `json:"version"`
}

type DrbdEnableSettingRspBody struct {
	DrbdEnableSetting *DrbdEnableSetting `json:"data,omitempty"`
}

type DrbdEnableSettingReqBody struct {
	Enable bool `json:"enable,omitempty"`
}

type RspFailBody struct {
	ErrCode int    `json:"errcode"`
	Desc    string `json:"description"`
}
