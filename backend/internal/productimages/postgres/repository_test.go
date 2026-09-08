package postgres

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/productimages"
)

func TestLockProductQuery(t *testing.T) {
	t.Parallel()

	query, args, err := bob.Build(context.Background(), lockProductQuery("product-id"))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	query = normalizeSQL(query)
	if query != `SELECT (CAST("id" AS text)) FROM products WHERE ("id" = $1) FOR UPDATE` {
		t.Fatalf("query = %q", query)
	}
	if !reflect.DeepEqual(args, []any{"product-id"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestClearPrimaryImageQuery(t *testing.T) {
	t.Parallel()

	query, args, err := bob.Build(context.Background(), clearPrimaryImageQuery("product-id"))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	query = normalizeSQL(query)
	for _, fragment := range []string{
		`UPDATE product_images`,
		`SET "is_primary" = $1`,
		`"product_id" = $2`,
		`"is_primary" = $3`,
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query %q does not contain %q", query, fragment)
		}
	}
	if !reflect.DeepEqual(args, []any{false, "product-id", true}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestInsertProductImageQuery(t *testing.T) {
	t.Parallel()

	width := 1200
	altText := "Party pig preview"
	image := productimages.NewProductImage{
		R2ObjectKey:      "products/product-id/party-pig.webp",
		OriginalFilename: "party-pig.webp",
		ContentType:      "image/webp",
		FileSizeBytes:    4096,
		WidthPX:          &width,
		AltText:          &altText,
		DisplayOrder:     2,
		IsPrimary:        true,
	}

	query, args, err := bob.Build(
		context.Background(),
		insertProductImageQuery("product-id", image),
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	query = normalizeSQL(query)
	for _, fragment := range []string{
		`INSERT INTO product_images`,
		`"r2_object_key"`,
		`VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		`RETURNING (CAST("id" AS text)), (CAST("product_id" AS text))`,
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query %q does not contain %q", query, fragment)
		}
	}
	if len(args) != 11 || args[0] != "product-id" || args[1] != image.R2ObjectKey {
		t.Fatalf("args = %#v", args)
	}
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

func TestProductImageFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[productimages.ProductImage]()
	if err != nil {
		t.Fatalf("StructMapperColumns() error = %v", err)
	}
	want := []string{
		"id",
		"product_id",
		"r2_object_key",
		"original_filename",
		"content_type",
		"file_size_bytes",
		"width_px",
		"height_px",
		"alt_text",
		"display_order",
		"is_primary",
		"r2_etag",
		"created_at",
		"updated_at",
	}
	if !reflect.DeepEqual(columns, want) {
		t.Fatalf("columns = %#v, want %#v", columns, want)
	}
}
