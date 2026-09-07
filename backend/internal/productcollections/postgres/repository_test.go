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
	"github.com/zdenaforero/svg-piggies/backend/internal/productcollections"
)

func TestProductCollectionQueries(t *testing.T) {
	t.Parallel()

	createInput := productcollections.CreateProductCollectionInput{
		ProductID:    "product-id",
		CollectionID: "collection-id",
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
			query: listProductCollectionsQuery("product-id"),
			fragments: []string{
				`SELECT (CAST("product_id" AS text)), (CAST("collection_id" AS text))`,
				`FROM product_collections`,
				`WHERE ("product_id" = $1)`,
				`ORDER BY "collection_id"`,
			},
			wantArgs: []any{"product-id"},
		},
		{
			name:  "create",
			query: createProductCollectionQuery(createInput),
			fragments: []string{
				`INSERT INTO product_collections`,
				`("product_id", "collection_id")`,
				`VALUES ($1, $2)`,
				`RETURNING (CAST("product_id" AS text)), (CAST("collection_id" AS text))`,
			},
			wantArgs: []any{"product-id", "collection-id"},
		},
		{
			name:  "delete",
			query: deleteProductCollectionQuery("product-id", "collection-id"),
			fragments: []string{
				`DELETE FROM product_collections`,
				`WHERE (("product_id" = $1) AND ("collection_id" = $2))`,
			},
			wantArgs: []any{"product-id", "collection-id"},
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

func TestProductCollectionFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[productcollections.ProductCollection]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{"product_id", "collection_id"}
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
		{name: "foreign key", code: "23503", want: productcollections.ErrReferenceNotFound},
		{name: "unique composite key", code: "23505", want: productcollections.ErrConflict},
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
