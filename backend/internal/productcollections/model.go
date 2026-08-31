package productcollections

type ProductCollection struct {
	ProductID    string `json:"productId"`
	CollectionID string `json:"collectionId"`
}

type CreateProductCollectionInput struct {
	ProductID    string `json:"productId"`
	CollectionID string `json:"collectionId"`
}
