package responses

type (
	ProductInfo struct {
	}

	GetProductCatalogueResponse struct {
		ListProducts []ProductInfo `json:"list_products"`
	}

	GetProductDetailResponse struct {
	}
)
