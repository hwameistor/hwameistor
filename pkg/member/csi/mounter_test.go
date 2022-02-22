package csi

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/utils/mount"
)

func Test_linuxMounter_GetDeviceMountPoints(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		devPath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if got := m.GetDeviceMountPoints(tt.args.devPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("linuxMounter.GetDeviceMountPoints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linuxMounter_Unmount(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		mountPoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if err := m.Unmount(tt.args.mountPoint); (err != nil) != tt.wantErr {
				t.Errorf("linuxMounter.Unmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_linuxMounter_BindMount(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		devPath    string
		mountPoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if err := m.BindMount(tt.args.devPath, tt.args.mountPoint); (err != nil) != tt.wantErr {
				t.Errorf("linuxMounter.BindMount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_linuxMounter_doBindMount(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		devPath    string
		mountPoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if err := m.doBindMount(tt.args.devPath, tt.args.mountPoint); (err != nil) != tt.wantErr {
				t.Errorf("linuxMounter.doBindMount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_linuxMounter_bindMount(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		devPath    string
		mountPoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if err := m.bindMount(tt.args.devPath, tt.args.mountPoint); (err != nil) != tt.wantErr {
				t.Errorf("linuxMounter.bindMount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_linuxMounter_isNotBindMountPoint(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		mountPoint string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if got := m.isNotBindMountPoint(tt.args.mountPoint); got != tt.want {
				t.Errorf("linuxMounter.isNotBindMountPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linuxMounter_MountRawBlock(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		devPath    string
		mountPoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if err := m.MountRawBlock(tt.args.devPath, tt.args.mountPoint); (err != nil) != tt.wantErr {
				t.Errorf("linuxMounter.MountRawBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_linuxMounter_FormatAndMount(t *testing.T) {
	type fields struct {
		mounter *mount.SafeFormatAndMount
		logger  *log.Entry
	}
	type args struct {
		devPath    string
		mountPoint string
		fsType     string
		options    []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &linuxMounter{
				mounter: tt.fields.mounter,
				logger:  tt.fields.logger,
			}
			if err := m.FormatAndMount(tt.args.devPath, tt.args.mountPoint, tt.args.fsType, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("linuxMounter.FormatAndMount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
