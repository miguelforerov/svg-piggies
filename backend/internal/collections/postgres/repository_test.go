package postgres

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
	"github.com/zdenaforero/svg-piggies/backend/internal/collections"
)

func TestCollectionQueries(t *testing.T) {
	t.Parallel()

	createInput := collections.CreateCollectionInput{
		Name: "Animals", Slug: "animals", Description: "Animal illustrations",
	}
	updateInput := collections.UpdateCollectionInput{
		Name: "Wildlife", Slug: "wildlife", Description: "Wildlife illustrations",
	}
	tests := []struct {
		name      string
		query     bob.Query
		fragments []string
		wantArgs  []any
	}{
		{
			name:  "list",
			query: listCollectionsQuery(),
			fragments: []string{
				`SELECT (CAST("id" AS text)), "name", "slug", "description"`,
				`FROM collections`,
				`ORDER BY "name", "id"`,
			},
		},
		{
			name:      "get by id",
			query:     getCollectionQuery("id", "collection-id"),
			fragments: []string{`FROM collections`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"collection-id"},
		},
		{
			name:      "get by slug",
			query:     getCollectionQuery("slug", "animals"),
			fragments: []string{`FROM collections`, `WHERE ("slug" = $1)`},
			wantArgs:  []any{"animals"},
		},
		{
			name:      "create",
			query:     createCollectionQuery(createInput),
			fragments: []string{`INSERT INTO collections`, `VALUES ($1, $2, $3)`, `RETURNING`},
			wantArgs:  []any{"Animals", "animals", "Animal illustrations"},
		},
		{
			name:  "update",
			query: updateCollectionQuery("collection-id", updateInput),
			fragments: []string{
				`UPDATE collections`,
				`"name" = $1`,
				`"slug" = $2`,
				`"description" = $3`,
				`WHERE ("id" = $4)`,
				`RETURNING`,
			},
			wantArgs: []any{"Wildlife", "wildlife", "Wildlife illustrations", "collection-id"},
		},
		{
			name:      "delete",
			query:     deleteCollectionQuery("collection-id"),
			fragments: []string{`DELETE FROM collections`, `WHERE ("id" = $1)`},
			wantArgs:  []any{"collection-id"},
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

func TestCollectionFieldsMatchBobResultColumns(t *testing.T) {
	t.Parallel()

	columns, err := scan.StructMapperColumns[collections.Collection]()
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
