package stor

import (
	"encoding/json"
	"fmt"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

const (
	// The storcli is the command line management software designed for the MegaRAID® product line.
	defaultCmd = "storcli"

	// Print the number of controllers connected in JSON format.
	STOR_SHOW_CTRL_COUNT = "show ctrlcount j"
)

// Stor implements the raid.Manager interface.
type Stor struct {
	Path string `json:"Path"`
}

// NewStor returns a Stor instance with default values.
func NewStor() *Stor {
	return &Stor{Path: defaultCmd}
}

// GetControllerCount implements the raid.Manager interface.
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

	if len(result.Controllers) < 1 {
		return count, fmt.Errorf("no controller found in the result: %v", result)
	}

	if result.Controllers[0].CommandStatus.Status == "Success" {
		count = result.Controllers[0].ResponseData.ControllerCount
		return count, nil

	}

	return count, fmt.Errorf("get controller count err: %s",
		result.Controllers[0].CommandStatus.Description)
}
