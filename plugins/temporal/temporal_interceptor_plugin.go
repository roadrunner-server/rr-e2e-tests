package tests

import (
	"context"
	"os"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

type TemporalInterceptorPlugin struct {
	config Configurer
}

func (p1 *TemporalInterceptorPlugin) Init(cfg Configurer) error {
	p1.config = cfg
	return nil
}

func (p1 *TemporalInterceptorPlugin) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *TemporalInterceptorPlugin) Stop(context.Context) error {
	return nil
}

func (p1 *TemporalInterceptorPlugin) Name() string {
	return "temporal_test.incterceptor_plugin"
}

func (p *TemporalInterceptorPlugin) TemporalInterceptor() interceptor.WorkerInterceptor {
	return &workerInterceptor{}
}

type workerInterceptor struct {
	interceptor.WorkerInterceptorBase
}

func (w *workerInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	f, err := os.Create("./interceptor_test")
	if err != nil {
		panic(err)
	}
	f.Close()
	return next
}

func (w *workerInterceptor) InterceptWorkflow(
	ctx workflow.Context,
	next interceptor.WorkflowInboundInterceptor,
) interceptor.WorkflowInboundInterceptor {
	return next
}
