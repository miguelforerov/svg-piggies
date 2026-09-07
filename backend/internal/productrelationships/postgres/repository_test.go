package postgres

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/productrelationships"
)

func TestProductRelationshipQueries(t *testing.T) {
	t.Parallel()

	createInput := productrelationships.CreateProductRelationshipInput{
		ProductID:        "product-id",
		RelatedProductID: "related-product-id",
		DisplayOrder:     3,
	}
	updateInput := productrelationships.UpdateProductRelationshipInput{
		RelatedProductID: "new-related-product-id",
		DisplayOrder:     5,
	}
	tests := []struct {
		name      string
		query     bob.Query
		fragments []string
		wantArgs  []any
	}{
		{
			name:      "product exists",
			query:     productExistsQuery("product-id"),
			fragments: []string{`SELECT (CAST("id" AS text))`, `FROM products`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"product-id"},
		},
		{
			name:  "locked product",
			query: lockedProductQuery("product-id"),
			fragments: []string{
				`FROM products`, `WHERE ("id" = $1)`, `FOR UPDATE`,
			},
			wantArgs: []any{"product-id"},
		},
		{
			name:  "list",
			query: listRelationshipsQuery("product-id"),
			fragments: []string{
				`FROM product_relationships`,
				`WHERE ("product_id" = $1)`,
				`ORDER BY "display_order", "id"`,
			},
			wantArgs: []any{"product-id"},
		},
		{
			name:  "get",
			query: getRelationshipQuery("product-id", "relationship-id"),
			fragments: []string{
				`FROM product_relationships`,
				`WHERE (("product_id" = $1) AND ("id" = $2))`,
			},
			wantArgs: []any{"product-id", "relationship-id"},
		},
		{
			name:      "create",
			query:     createRelationshipQuery(createInput),
			fragments: []string{`INSERT INTO product_relationships`, `VALUES ($1, $2, $3)`, `RETURNING`},
			wantArgs:  []any{"product-id", "related-product-id", 3},
		},
		{
			name:  "next display order",
			query: nextDisplayOrderQuery("product-id"),
			fragments: []string{
				`SELECT COALESCE((MAX("display_order") + $1), $2)`,
				`FROM product_relationships`,
				`WHERE ("product_id" = $3)`,
			},
			wantArgs: []any{1, 0, "product-id"},
		},
		{
			name: "create ignoring duplicate",
			query: createRelationshipIgnoringConflictQuery(
				"product-id", "related-product-id", 4,
			),
			fragments: []string{
				`INSERT INTO product_relationships`,
				`VALUES ($1, $2, $3)`,
				`ON CONFLICT ("product_id", "related_product_id") DO NOTHING`,
			},
			wantArgs: []any{"product-id", "related-product-id", 4},
		},
		{
			name:      "replace insert",
			query:     insertRelationshipQuery("product-id", "related-product-id", 0),
			fragments: []string{`INSERT INTO product_relationships`, `VALUES ($1, $2, $3)`},
			wantArgs:  []any{"product-id", "related-product-id", 0},
		},
		{
			name:  "update",
			query: updateRelationshipQuery("product-id", "relationship-id", updateInput),
			fragments: []string{
				`UPDATE product_relationships`,
				`"related_product_id" = $1`,
				`"display_order" = $2`,
				`WHERE (("product_id" = $3) AND ("id" = $4))`,
				`RETURNING`,
			},
			wantArgs: []any{"new-related-product-id", 5, "product-id", "relationship-id"},
		},
		{
			name:      "delete all for product",
			query:     deleteRelationshipsByProductQuery("product-id"),
			fragments: []string{`DELETE FROM product_relationships`, `WHERE ("product_id" = $1)`},
			wantArgs:  []any{"product-id"},
		},
		{
			name:  "delete one",
			query: deleteRelationshipQuery("product-id", "relationship-id"),
			fragments: []string{
				`DELETE FROM product_relationships`,
				`WHERE (("product_id" = $1) AND ("id" = $2))`,
			},
			wantArgs: []any{"product-id", "relationship-id"},
		},
		{
			name:  "populated relationships",
			query: populatedRelationshipsQuery("product-id"),
			fragments: []string{
				`FROM product_relationships AS "pr"`,
				`INNER JOIN products AS "p" ON ("p"."id" = "pr"."related_product_id")`,
				`WHERE ("pr"."product_id" = $1)`,
				`ORDER BY "pr"."display_order", "pr"."id"`,
			},
			wantArgs: []any{"product-id"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			query, args, err := bob.Build(context.Background(), test.query)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			query = normalizeSQL(query)
			for _, fragment := range test.fragments {
				if !strings.Contains(query, fragment) {
					t.Fatalf("query %q does not contain %q", query, fragment)
				}
			}
			if !reflect.DeepEqual(args, test.wantArgs) {
				t.Fatalf("args = %#v, want %#v", args, test.wantArgs)
			}
		})
	}
}

func TestProductRelationshipFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[productrelationships.ProductRelationship]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{"id", "product_id", "related_product_id", "display_order"}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("columns = %#v, want %#v", columns, want)
	}
}

func TestPopulatedRelationshipFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[populatedRelationshipRow]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{
		"relationship_id",
		"display_order",
		"product_id",
		"product_title",
		"product_slug",
		"product_description",
		"product_price",
		"product_status",
		"product_created_at",
		"product_updated_at",
	}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("columns = %#v, want %#v", columns, want)
	}
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
