package csi

import (
	"reflect"
	"testing"
)

func Test_parseParameters(t *testing.T) {
	type args struct {
		req RequestParameterHandler
	}
	tests := []struct {
		name    string
		args    args
		want    *volumeParameters
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParameters(tt.args.req)
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
