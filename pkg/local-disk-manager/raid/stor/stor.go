package stor

import (
	"encoding/json"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

const (
	defaultCmd = "storcli"

	STOR_SHOW_CTRL_COUNT = "show ctrlcount j"
)

type Stor struct {
	Path string `json:"Path"`
}

func NewStor() *Stor {
	return &Stor{Path: defaultCmd}
}

func (s *Stor) GetControllerCount() (int, error) {
	var (
		count  int
		err    error
		info   []byte
		result ControllerCounts
	)

	info, err = utils.ExecCmd(s.Path, STOR_SHOW_CTRL_COUNT)
	if err != nil {
		return count, err
	}

	err = json.Unmarshal(info, &result)
	if err != nil {
		return count, err
	}

	if result.Controllers[0].CommandStatus.Status == "Success" {
		count = result.Controllers[0].ResponseData.ControllerCount
		return count, nil

	}

	return count, fmt.Errorf("get controller count err: %s",
		result.Controllers[0].CommandStatus.Description)
}
