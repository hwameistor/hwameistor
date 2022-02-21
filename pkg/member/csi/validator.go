package csi

// RequestParameterHandler interface
type RequestParameterHandler interface {
	GetParameters() map[string]string
}
