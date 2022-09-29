package utils

import (
	"testing"
)

func TestWriteDataIntoSysFSFile(t *testing.T) {
	type args struct {
		content     string
		sysFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{content: "test", sysFilePath: "/test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteDataIntoSysFSFile(tt.args.content, tt.args.sysFilePath); (err != nil) != tt.wantErr {
				t.Errorf("WriteDataIntoSysFSFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
