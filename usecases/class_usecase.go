package usecases

import (
	"context"
	"log"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/logic/class_workflow"
)

type (
	ClassUsecase interface {
		AssignSenseiClass(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError)
	}

	classUsecase struct {
		workflow class_workflow.ClassWorkflow
	}
)

func (c *classUsecase) AssignSenseiClass(ctx context.Context, request *requests.AssignSenseiClassRequest) (response *responses.AssignSenseiClassResponse, customErr *apperror.CustomError) {
	if request == nil || len(request.StudentIDs) == 0 {
		return nil, apperror.InvalidRequestErr()
	}

	for index, studentID := range request.StudentIDs {
		log.Println("assigning sensei class to student for student ", studentID)
		response, customErr = c.workflow.AssignSenseiClassWorkflow(ctx, request, uint64(index))
		if customErr != nil {
			return nil, customErr
		}
	}

	return response, nil
}

func newClassUsecase(workflow class_workflow.ClassWorkflow) ClassUsecase {
	return &classUsecase{workflow: workflow}
}
