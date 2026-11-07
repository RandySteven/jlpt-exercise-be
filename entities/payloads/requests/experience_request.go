package requests

type (
	GetProductCatalogueRequest struct {
		ProductTypes []string `json:"product_types"`
	}

	GetProductDetailRequest struct {
		ProductCategory string `json:"product_category"`
		ID              uint64 `json:"id"`
	}
)
