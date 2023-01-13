package rest

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
)

func TestNew(t *testing.T) {
	type args struct {
		name      string
		namespace string
		httpPort  int
		member    apis.LocalStorageMember
		cli       client.Client
	}
	var wantServer = &restServer{}
	wantServer.name = "test_server"
	wantServer.namespace = "test"
	wantServer.httpPort = 8080
	wantServer.logger = log.WithField("Module", "RESTServer")

	tests := []struct {
		name string
		args args
		want Server
	}{
		{
			args: args{name: "test_server", namespace: "test", httpPort: 8080},
			want: wantServer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.name, tt.args.namespace, tt.args.httpPort, tt.args.member, tt.args.cli); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
