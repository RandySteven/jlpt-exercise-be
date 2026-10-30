package ai_chat_workflow

import (
	"context"
	"time"
)

func (w *ai_chatWorkflow) publishChat(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {

	return nil, nil
}

func (w *ai_chatWorkflow) consume(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	errChan := make(chan error, 1)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			err := w.workflow.GetSignalResult(nil, sgConsumeChat, &execData)
			if err != nil {
				errChan <- err
				continue
			}
			if execData != nil {
				return execData, nil
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errChan:
			return nil, err
		}
	}
}
