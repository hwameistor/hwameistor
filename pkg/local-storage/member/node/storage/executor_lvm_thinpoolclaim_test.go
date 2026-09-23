package storage

import (
	"bytes"
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

type recordingLVMExecutor struct {
	lvReport string
	calls    []exechelper.ExecParams
}

func (e *recordingLVMExecutor) RunCommand(params exechelper.ExecParams) exechelper.ExecResult {
	params.CmdArgs = append([]string(nil), params.CmdArgs...)
	e.calls = append(e.calls, params)
	if params.CmdName == "lvs" {
		return exechelper.ExecResult{
			OutBuf:   bytes.NewBufferString(e.lvReport),
			ErrBuf:   bytes.NewBuffer(nil),
			ExitCode: 0,
		}
	}
	return exechelper.ExecResult{
		OutBuf:   bytes.NewBuffer(nil),
		ErrBuf:   bytes.NewBuffer(nil),
		ExitCode: 0,
	}
}

func TestLVMExecutor_ExtendThinPoolMetadataDefault(t *testing.T) {
	twoGiB := uint(2)
	tests := []struct {
		name                 string
		currentMetadataBytes int64
		requestedMetadataGiB *uint
		newPool              bool
		wantCommand          string
		wantArgs             []string
		wantAbsentArgs       []string
	}{
		{
			name:                 "omitted metadata preserves an existing value below the default",
			currentMetadataBytes: 512 * utils.Mi,
			wantCommand:          "lvextend",
			wantArgs:             []string{"--size=101G"},
			wantAbsentArgs:       []string{"--poolmetadatasize=1G"},
		},
		{
			name:                 "omitted metadata preserves an existing value above the default",
			currentMetadataBytes: 1400 * utils.Mi,
			wantCommand:          "lvextend",
			wantArgs:             []string{"--size=101G"},
			wantAbsentArgs:       []string{"--poolmetadatasize=1G"},
		},
		{
			name:                 "explicit larger metadata extends the existing pool",
			currentMetadataBytes: utils.Gi,
			requestedMetadataGiB: &twoGiB,
			wantCommand:          "lvextend",
			wantArgs:             []string{"--size=101G", "--poolmetadatasize=2G"},
		},
		{
			name:        "new pool keeps the one gibibyte metadata default",
			newPool:     true,
			wantCommand: "lvcreate",
			wantArgs:    []string{"--poolmetadatasize=1G", "--size=101G"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lvReport := `{"report":[{"lv":[]}]}`
			if !tt.newPool {
				lvReport = fmt.Sprintf(`{"report":[{"lv":[{"lv_name":%q,"vg_name":%q,"lv_size":"%dB","lv_metadata_size":"%dB"}]}]}`,
					apisv1alpha1.ThinPoolName, "pool1", 100*utils.Gi, tt.currentMetadataBytes)
			}
			cmdExec := &recordingLVMExecutor{lvReport: lvReport}
			lvm := &lvmExecutor{cmdExec: cmdExec, logger: log.New().WithField("test", true)}
			claim := &apisv1alpha1.ThinPoolClaim{}
			claim.Spec.Description.PoolName = "pool1"
			claim.Spec.Description.Capacity = 101
			claim.Spec.Description.PoolMetadataSize = tt.requestedMetadataGiB

			if err := lvm.ExtendThinPool(claim); err != nil {
				t.Fatalf("ExtendThinPool: %v", err)
			}

			var command *exechelper.ExecParams
			for i := range cmdExec.calls {
				if cmdExec.calls[i].CmdName == tt.wantCommand {
					command = &cmdExec.calls[i]
					break
				}
			}
			if command == nil {
				t.Fatalf("calls = %#v, want a %s command", cmdExec.calls, tt.wantCommand)
			}
			for _, arg := range tt.wantArgs {
				if !hasLVMArg(command.CmdArgs, arg) {
					t.Errorf("%s args %q do not contain %q", command.CmdName, command.CmdArgs, arg)
				}
			}
			for _, arg := range tt.wantAbsentArgs {
				if hasLVMArg(command.CmdArgs, arg) {
					t.Errorf("%s args %q unexpectedly contain %q", command.CmdName, command.CmdArgs, arg)
				}
			}
		})
	}
}

func hasLVMArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
