package observer

import (
	"context"

	"dag-observatory/demo-go/internal/application/dagruntime/port"
)

type OTelObserver struct{}

func NewOTelObserver() *OTelObserver {
	return &OTelObserver{}
}

func (o *OTelObserver) OnCompile(ctx context.Context, info port.CompileInfo) {}
func (o *OTelObserver) OnCycleStart(ctx context.Context, info port.CycleInfo) {}
func (o *OTelObserver) OnCycleEnd(ctx context.Context, info port.CycleResult) {}
func (o *OTelObserver) OnNodeStart(ctx context.Context, info port.NodeInfo) {}
func (o *OTelObserver) OnNodeEnd(ctx context.Context, info port.NodeResult) {}
func (o *OTelObserver) OnError(ctx context.Context, info port.ErrorInfo) {}

