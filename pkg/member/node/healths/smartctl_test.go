package healths

import (
	"reflect"
	"testing"

	"github.com/HwameiStor/local-storage/pkg/exechelper"
)

func TestNewSmartCtl(t *testing.T) {
	tests := []struct {
		name string
		want DiskChecker
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSmartCtl(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSmartCtl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_smartCtlr_IsDiskHealthy(t *testing.T) {
	type fields struct {
		cmdExec exechelper.Executor
		cmdName string
	}
	type args struct {
		devPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &smartCtlr{
				cmdExec: tt.fields.cmdExec,
				cmdName: tt.fields.cmdName,
			}
			got, err := sc.IsDiskHealthy(tt.args.devPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("smartCtlr.IsDiskHealthy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("smartCtlr.IsDiskHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_smartCtlr_GetLocalDisksAll(t *testing.T) {
	type fields struct {
		cmdExec exechelper.Executor
		cmdName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []DeviceInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &smartCtlr{
				cmdExec: tt.fields.cmdExec,
				cmdName: tt.fields.cmdName,
			}
			got, err := sc.GetLocalDisksAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("smartCtlr.GetLocalDisksAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("smartCtlr.GetLocalDisksAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_smartCtlr_CheckHealthForLocalDisk(t *testing.T) {
	type fields struct {
		cmdExec exechelper.Executor
		cmdName string
	}
	type args struct {
		device *DeviceInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *DiskCheckResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &smartCtlr{
				cmdExec: tt.fields.cmdExec,
				cmdName: tt.fields.cmdName,
			}
			got, err := sc.CheckHealthForLocalDisk(tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("smartCtlr.CheckHealthForLocalDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("smartCtlr.CheckHealthForLocalDisk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_smartCtlr_scanForLocalDisks(t *testing.T) {
	type fields struct {
		cmdExec exechelper.Executor
		cmdName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []DeviceInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &smartCtlr{
				cmdExec: tt.fields.cmdExec,
				cmdName: tt.fields.cmdName,
			}
			got, err := sc.scanForLocalDisks()
			if (err != nil) != tt.wantErr {
				t.Errorf("smartCtlr.scanForLocalDisks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("smartCtlr.scanForLocalDisks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_smartCtlr_checkForDisk(t *testing.T) {
	type fields struct {
		cmdExec exechelper.Executor
		cmdName string
	}
	type args struct {
		device *DeviceInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *DiskCheckResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &smartCtlr{
				cmdExec: tt.fields.cmdExec,
				cmdName: tt.fields.cmdName,
			}
			got, err := sc.checkForDisk(tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("smartCtlr.checkForDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("smartCtlr.checkForDisk() = %v, want %v", got, tt.want)
			}
		})
	}
}
