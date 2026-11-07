package usecases

import (
	"context"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
)

type (
	ExperienceUsecase interface {
		GetAllProducts(ctx context.Context, request *requests.GetProductCatalogueRequest) (response *responses.GetProductCatalogueResponse, customErr *apperror.CustomError)
		GetProductDetail(ctx context.Context, request *requests.GetProductDetailRequest) (response *responses.GetProductDetailResponse, customErr *apperror.CustomError)
	}

	experienceUsecase struct {
	}
)

// GetAllProducts implements [ExperienceUsecase].
func (e *experienceUsecase) GetAllProducts(ctx context.Context, request *requests.GetProductCatalogueRequest) (response *responses.GetProductCatalogueResponse, customErr *apperror.CustomError) {
	return nil, nil
}

// GetProductDetail implements [ExperienceUsecase].
func (e *experienceUsecase) GetProductDetail(ctx context.Context, request *requests.GetProductDetailRequest) (response *responses.GetProductDetailResponse, customErr *apperror.CustomError) {
	return nil, nil
}

var _ ExperienceUsecase = &experienceUsecase{}
