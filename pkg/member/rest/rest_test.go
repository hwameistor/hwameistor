package rest

import (
	"reflect"
	"testing"

	"github.com/HwameiStor/local-storage/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNew(t *testing.T) {
	type args struct {
		name      string
		namespace string
		httpPort  int
		member    apis.LocalStorageMember
		cli       client.Client
	}
	tests := []struct {
		name string
		args args
		want Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.name, tt.args.namespace, tt.args.httpPort, tt.args.member, tt.args.cli); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
