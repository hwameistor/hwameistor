package csi

import (
	"reflect"
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

func Test_newControllerServiceCapability(t *testing.T) {
	type args struct {
		cap csi.ControllerServiceCapability_RPC_Type
	}
	var wantCap = &csi.ControllerServiceCapability{}
	wantCap.Type = &csi.ControllerServiceCapability_Rpc{
		Rpc: &csi.ControllerServiceCapability_RPC{
			Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		},
	}
	tests := []struct {
		name string
		args args
		want *csi.ControllerServiceCapability
	}{
		{
			args: args{cap: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT},
			want: wantCap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newControllerServiceCapability(tt.args.cap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newControllerServiceCapability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newNodeServiceCapability(t *testing.T) {
	type args struct {
		cap csi.NodeServiceCapability_RPC_Type
	}
	var wantCap = &csi.NodeServiceCapability{}
	wantCap.Type = &csi.NodeServiceCapability_Rpc{
		Rpc: &csi.NodeServiceCapability_RPC{
			Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		},
	}
	tests := []struct {
		name string
		args args
		want *csi.NodeServiceCapability
	}{
		{
			args: args{cap: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME},
			want: wantCap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newNodeServiceCapability(tt.args.cap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNodeServiceCapability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newPluginCapability(t *testing.T) {
	type args struct {
		cap csi.PluginCapability_Service_Type
	}
	var wantCap = &csi.PluginCapability{}
	wantCap.Type = &csi.PluginCapability_Service_{
		Service: &csi.PluginCapability_Service{
			Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
		},
	}
	tests := []struct {
		name string
		args args
		want *csi.PluginCapability
	}{
		{
			args: args{cap: csi.PluginCapability_Service_CONTROLLER_SERVICE},
			want: wantCap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPluginCapability(tt.args.cap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newPluginCapability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEndpoint(t *testing.T) {
	type args struct {
		ep string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			args:    args{ep: "unix://127.0.0.1:8080"},
			want:    "unix",
			want1:   "127.0.0.1:8080",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseEndpoint(tt.args.ep)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseEndpoint() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseEndpoint() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_pathExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			args:    args{path: "/data/test"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pathExists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("pathExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("pathExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVolumeMetrics(t *testing.T) {
	type args struct {
		mntPoint string
	}
	tests := []struct {
		name    string
		args    args
		want    *VolumeMetrics
		wantErr bool
	}{
		//{
		//	args:  args{mntPoint: "/data/test"},
		//	want:  nil,
		//	wantErr: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVolumeMetrics(tt.args.mntPoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVolumeMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVolumeMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isStringInArray(t *testing.T) {
	type args struct {
		str  string
		strs []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			args: args{str: "test1", strs: []string{"test1", "test2"}},
			want: true,
		},
		{
			args: args{str: "test3", strs: []string{"test1", "test2"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStringInArray(tt.args.str, tt.args.strs); got != tt.want {
				t.Errorf("isStringInArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
