package productrelationships

import "github.com/zdenaforero/svg-piggies/backend/internal/products"

type ProductRelationship struct {
	ID               string `json:"id"`
	ProductID        string `json:"productId"`
	RelatedProductID string `json:"relatedProductId"`
	DisplayOrder     int    `json:"displayOrder"`
}

type CreateProductRelationshipInput struct {
	ProductID        string `json:"productId"`
	RelatedProductID string `json:"relatedProductId"`
	DisplayOrder     int    `json:"displayOrder"`
}

type CreateProductRelationshipsInput struct {
	ProductID         string   `json:"productId"`
	RelatedProductIDs []string `json:"relatedProductIds"`
}

type UpdateProductRelationshipInput struct {
	RelatedProductID string `json:"relatedProductId"`
	DisplayOrder     int    `json:"displayOrder"`
}

type ReplaceProductRelationshipsInput struct {
	ProductID         string   `json:"productId"`
	RelatedProductIDs []string `json:"relatedProductIds"`
}

type PopulatedProductRelationship struct {
	RelationshipID string           `json:"relationshipId"`
	DisplayOrder   int              `json:"displayOrder"`
	Product        products.Product `json:"product"`
}

type ProductWithRelationships struct {
	Product       products.Product               `json:"product"`
	Relationships []PopulatedProductRelationship `json:"relationships"`
}
