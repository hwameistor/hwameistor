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
	tests := []struct {
		name string
		args args
		want *csi.ControllerServiceCapability
	}{
		// TODO: Add test cases.
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
	tests := []struct {
		name string
		args args
		want *csi.NodeServiceCapability
	}{
		// TODO: Add test cases.
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
	tests := []struct {
		name string
		args args
		want *csi.PluginCapability
	}{
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
