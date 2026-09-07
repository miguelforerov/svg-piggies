package postgres

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/products"
)

func TestProductQueries(t *testing.T) {
	t.Parallel()

	createInput := products.CreateProductInput{
		Title:       "Party Pig",
		Slug:        "party-pig",
		Description: "Printable party pig",
		Price:       "19.95",
		Status:      products.StatusActive,
	}
	updateInput := products.UpdateProductInput{
		Title:       "Celebration Pig",
		Slug:        "celebration-pig",
		Description: "Updated printable pig",
		Price:       "24.50",
		Status:      products.StatusArchived,
	}
	tests := []struct {
		name      string
		query     bob.Query
		fragments []string
		wantArgs  []any
	}{
		{
			name:  "list",
			query: listProductsQuery(),
			fragments: []string{
				`SELECT (CAST("id" AS text)), "title", "slug", "description"`,
				`(CAST("price" AS text)), (CAST("status" AS text))`,
				`FROM products`,
				`ORDER BY "created_at" DESC, "id"`,
			},
		},
		{
			name:      "get by id",
			query:     getProductQuery("id", "product-id"),
			fragments: []string{`FROM products`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"product-id"},
		},
		{
			name:      "get by slug",
			query:     getProductQuery("slug", "party-pig"),
			fragments: []string{`FROM products`, `WHERE ("slug" = $1)`},
			wantArgs:  []any{"party-pig"},
		},
		{
			name:  "create",
			query: createProductQuery(createInput),
			fragments: []string{
				`INSERT INTO products`,
				`CAST($4 AS numeric)`,
				`CAST($5 AS product_status)`,
				`RETURNING`,
			},
			wantArgs: []any{
				"Party Pig", "party-pig", "Printable party pig", "19.95", products.StatusActive,
			},
		},
		{
			name:  "update",
			query: updateProductQuery("product-id", updateInput),
			fragments: []string{
				`UPDATE products`,
				`"title" = $1`,
				`"price" = (CAST($4 AS numeric))`,
				`"status" = (CAST($5 AS product_status))`,
				`WHERE ("id" = $6)`,
				`RETURNING`,
			},
			wantArgs: []any{
				"Celebration Pig",
				"celebration-pig",
				"Updated printable pig",
				"24.50",
				products.StatusArchived,
				"product-id",
			},
		},
		{
			name:      "delete",
			query:     deleteProductQuery("product-id"),
			fragments: []string{`DELETE FROM products`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"product-id"},
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

func TestProductFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[products.Product]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{
		"id",
		"title",
		"slug",
		"description",
		"price",
		"status",
		"created_at",
		"updated_at",
	}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("columns = %#v, want %#v", columns, want)
	}
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
