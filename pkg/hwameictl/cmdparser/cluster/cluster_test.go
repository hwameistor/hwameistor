package cluster_test

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser/cluster"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterHelpCalled(t *testing.T) {
	cmd := &cobra.Command{}
	err := cluster.Cluster.RunE(cmd, []string{})
	assert.NoError(t, err, "Expected no error to be returned when RunE is called")
}
