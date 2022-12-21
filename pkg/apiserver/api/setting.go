package api

type DrbdEnableSetting struct {
	Enabledrbd bool   `json:"enabledrbd"`
	State      State  `json:"state"`
	Version    string `json:"version"`
}

type DrbdEnableSettingRspBody struct {
	DrbdEnableSetting *DrbdEnableSetting `json:"data,omitempty"`
}

type RspFailBody struct {
	ErrCode int    `json:"errcode"`
	Desc    string `json:"description"`
}
