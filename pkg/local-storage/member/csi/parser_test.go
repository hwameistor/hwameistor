package csi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func Test_parseParameters(t *testing.T) {
	type args struct {
		req RequestParameterHandler
	}
	var param = make(map[string]string)
	param[apisv1alpha1.VolumeParameterPoolClassKey] = "HDD"
	param[apisv1alpha1.VolumeParameterPoolTypeKey] = "Regular"
	param[apisv1alpha1.VolumeParameterReplicaNumberKey] = "1"
	param[apisv1alpha1.VolumeParameterConvertible] = "true"
	param[pvcNamespaceKey] = "default"
	param[pvcNameKey] = "pvc-1"

	var req = &csi.CreateVolumeRequest{}
	req.Parameters = param

	var param2 = make(map[string]string)
	param2[apisv1alpha1.VolumeParameterPoolClassKey] = "HDD"
	param2[apisv1alpha1.VolumeParameterPoolTypeKey] = "Regular"
	param2[apisv1alpha1.VolumeParameterReplicaNumberKey] = "2"
	param2[apisv1alpha1.VolumeParameterConvertible] = "true"
	param2[pvcNamespaceKey] = "default"
	param[pvcNameKey] = "pvc-2"

	var req2 = &csi.CreateVolumeRequest{}
	req2.Parameters = param2

	tests := []struct {
		name    string
		args    args
		want    *volumeParameters
		wantErr bool
	}{
		{
			args:    args{req: req},
			want:    nil,
			wantErr: true,
		},
		{
			args:    args{req: req2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParameters(tt.args.req.(*csi.CreateVolumeRequest))
			fmt.Printf("got = %+v", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
