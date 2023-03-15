package node

import "context"

func (m *nodeManager) startDiskClaimTaskWorker(ctx context.Context) {
	<-ctx.Done()
	return
}
