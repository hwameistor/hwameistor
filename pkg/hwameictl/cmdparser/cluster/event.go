package cluster

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/formatter"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/manager"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"os/exec"
)

var eventName string

var event = &cobra.Command{
	Use:   "event",
	Args:  cobra.ExactArgs(0),
	Short: "list Hwameistor cluster events.",
	Long: "You can use 'hwameictl cluster event' to list Hwameistor cluster events.\n" +
		"Furthermore, you can use 'hwameictl cluster event --name {eventName}' to Get details",
	Example: "hwameictl cluster event \n" +
		"hwameictl cluster event --name example-event",
	RunE: eventRunE,
}

func init() {
	// Disk list flags
	event.Flags().StringVar(&eventName, "name", "", "get event's info by event name")
}

func eventRunE(_ *cobra.Command, _ []string) error {
	m, err := manager.NewMetricsController()
	if err != nil {
		return err
	}

	if eventName == "" {
		evList := &apisv1alpha1.EventList{}
		if err := m.Client.List(context.TODO(), evList); err != nil {
			//log.WithError(err).Error("Failed to list Event")
			return err
		}

		eventsHeader := table.Row{"#", "Name", "ResourceName", "ResourceType", "Action", "Time"}
		var eventsRows []table.Row
		index := 0
		for _, event := range evList.Items {

			index++
			eventsRows = append(eventsRows, table.Row{index, event.Name, event.Spec.ResourceName, event.Spec.ResourceType,
				event.Spec.Records[0].Action, event.CreationTimestamp})

		}
		formatter.PrintTable("Disks", eventsHeader, eventsRows)
	} else {

		cmd := exec.Command("kubectl", "get", "evt", eventName, "-oyaml")

		// 执行命令并获取输出
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("执行命令出错:", err)
			return err
		}

		// 打印命令输出
		fmt.Println(string(output))
	}

	return nil
}
