package taskservice

import (
	"context"
)

// TaskHandler is a service that accepts and handles tasks provided by client
// Also it returns back info on the status of the request
type TaskHandler interface {
	// get the task and its status by id
	GetByID(ctx context.Context, args ParamGetByID) (ResponseGetByID, error)
	// send task to the server
	Handle(ctx context.Context, args ParamsHandle) (ResponseHandle, error)
}
