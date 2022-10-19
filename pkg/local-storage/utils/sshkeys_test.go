package utils

import (
	"testing"
)

func TestAddPubKeyIntoAuthorizedKeys(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{content: "test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddPubKeyIntoAuthorizedKeys(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("AddPubKeyIntoAuthorizedKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
