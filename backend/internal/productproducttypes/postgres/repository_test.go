package postgres

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/productproducttypes"
)

func TestProductProductTypeQueries(t *testing.T) {
	t.Parallel()

	createInput := productproducttypes.CreateProductProductTypeInput{
		ProductID:     "product-id",
		ProductTypeID: "product-type-id",
	}
	tests := []struct {
		name      string
		query     bob.Query
		fragments []string
		wantArgs  []any
	}{
		{
			name:  "product exists",
			query: productExistsQuery("product-id"),
			fragments: []string{
				`SELECT (CAST("id" AS text))`,
				`FROM products`,
				`WHERE ("id" = $1)`,
			},
			wantArgs: []any{"product-id"},
		},
		{
			name:  "list",
			query: listProductProductTypesQuery("product-id"),
			fragments: []string{
				`SELECT (CAST("product_id" AS text)), (CAST("product_type_id" AS text))`,
				`FROM product_product_types`,
				`WHERE ("product_id" = $1)`,
				`ORDER BY "product_type_id"`,
			},
			wantArgs: []any{"product-id"},
		},
		{
			name:  "create",
			query: createProductProductTypeQuery(createInput),
			fragments: []string{
				`INSERT INTO product_product_types`,
				`("product_id", "product_type_id")`,
				`VALUES ($1, $2)`,
				`RETURNING (CAST("product_id" AS text)), (CAST("product_type_id" AS text))`,
			},
			wantArgs: []any{"product-id", "product-type-id"},
		},
		{
			name:  "delete",
			query: deleteProductProductTypeQuery("product-id", "product-type-id"),
			fragments: []string{
				`DELETE FROM product_product_types`,
				`WHERE (("product_id" = $1) AND ("product_type_id" = $2))`,
			},
			wantArgs: []any{"product-id", "product-type-id"},
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

func TestProductProductTypeFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[productproducttypes.ProductProductType]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{"product_id", "product_type_id"}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("columns = %#v, want %#v", columns, want)
	}
}

func TestMapErrorPreservesConstraintErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want error
	}{
		{name: "foreign key", code: "23503", want: productproducttypes.ErrReferenceNotFound},
		{name: "unique composite key", code: "23505", want: productproducttypes.ErrConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := mapError(&pgconn.PgError{Code: test.code})
			if !errors.Is(err, test.want) {
				t.Fatalf("mapError() = %v, want %v", err, test.want)
			}
		})
	}
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
