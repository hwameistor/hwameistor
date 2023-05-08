package node

import "context"

func (m *nodeManager) startDiskNodeTaskWorker(ctx context.Context) {
	<-ctx.Done()
	return
}
