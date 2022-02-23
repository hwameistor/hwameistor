package scheduler

import (
	"fmt"
	"testing"
)

func TestGetKubeconfigPath(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKubeconfigPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKubeconfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetKubeconfigPath() = %v, want %v", got, tt.want)
			}
			fmt.Printf("got = %s\n, err = %v\n", got, err)
		})
	}
}
