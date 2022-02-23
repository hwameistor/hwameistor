package csi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
)

func Test_parseParameters(t *testing.T) {
	type args struct {
		req RequestParameterHandler
	}
	var param = make(map[string]string)
	param[localstoragev1alpha1.VolumeParameterPoolClassKey] = "DISK"
	param[localstoragev1alpha1.VolumeParameterPoolTypeKey] = "DISK"
	param[localstoragev1alpha1.VolumeParameterVolumeKindKey] = "DISK"

	var req = &csi.CreateVolumeRequest{}
	req.Parameters = param

	tests := []struct {
		name    string
		args    args
		want    *volumeParameters
		wantErr bool
	}{
		{
			args:  args{req: req},
			want:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParameters(tt.args.req)
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
