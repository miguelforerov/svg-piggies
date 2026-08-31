package productproducttypes

type ProductProductType struct {
	ProductID     string `json:"productId"`
	ProductTypeID string `json:"productTypeId"`
}

type CreateProductProductTypeInput struct {
	ProductID     string `json:"productId"`
	ProductTypeID string `json:"productTypeId"`
}
