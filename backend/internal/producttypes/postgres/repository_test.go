package postgres

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/producttypes"
)

func TestProductTypeQueries(t *testing.T) {
	t.Parallel()

	createInput := producttypes.CreateProductTypeInput{
		Name: "Printable", Slug: "printable", Description: "Printable products",
	}
	updateInput := producttypes.UpdateProductTypeInput{
		Name: "Digital Printable", Slug: "digital-printable", Description: "Digital products",
	}
	tests := []struct {
		name      string
		query     bob.Query
		fragments []string
		wantArgs  []any
	}{
		{
			name:  "list",
			query: listProductTypesQuery(),
			fragments: []string{
				`SELECT (CAST("id" AS text)), "name", "slug", "description"`,
				`FROM product_types`,
				`ORDER BY "name", "id"`,
			},
		},
		{
			name:      "get",
			query:     getProductTypeQuery("product-type-id"),
			fragments: []string{`FROM product_types`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"product-type-id"},
		},
		{
			name:      "create",
			query:     createProductTypeQuery(createInput),
			fragments: []string{`INSERT INTO product_types`, `VALUES ($1, $2, $3)`, `RETURNING`},
			wantArgs:  []any{"Printable", "printable", "Printable products"},
		},
		{
			name:  "update",
			query: updateProductTypeQuery("product-type-id", updateInput),
			fragments: []string{
				`UPDATE product_types`,
				`"name" = $1`,
				`"slug" = $2`,
				`"description" = $3`,
				`WHERE ("id" = $4)`,
				`RETURNING`,
			},
			wantArgs: []any{
				"Digital Printable", "digital-printable", "Digital products", "product-type-id",
			},
		},
		{
			name:      "delete",
			query:     deleteProductTypeQuery("product-type-id"),
			fragments: []string{`DELETE FROM product_types`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"product-type-id"},
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

func TestProductTypeFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[producttypes.ProductType]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{"id", "name", "slug", "description"}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("columns = %#v, want %#v", columns, want)
	}
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
