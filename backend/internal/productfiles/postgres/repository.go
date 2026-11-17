package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/productfiles"
)

type Repository struct{ provider database.Provider }

func NewRepository(provider database.Provider) (*Repository, error) {
	if provider == nil {
		return nil, errors.New("database provider is required")
	}
	return &Repository{provider: provider}, nil
}

const columns = `id::text, product_id::text, file_key, file_name, display_name, format, content_type,
 file_size_bytes, r2_etag, availability_status, confirmation_fingerprint, fingerprint_key_version,
 display_order, created_at, updated_at`

const consumptionColumns = `file_key, product_id::text, token_digest, confirmation_fingerprint,
 fingerprint_key_version, expires_at, response_snapshot`

func scanConsumption(row pgx.Row) (productfiles.Consumption, error) {
	var result productfiles.Consumption
	var digest *string
	var snapshot []byte
	err := row.Scan(&result.FileKey, &result.ProductID, &digest, &result.ConfirmationFingerprint,
		&result.FingerprintKeyVersion, &result.ExpiresAt, &snapshot)
	if err != nil {
		return result, err
	}
	if digest != nil {
		result.TokenDigest = *digest
	}
	if err := json.Unmarshal(snapshot, &result.Snapshot); err != nil {
		return result, fmt.Errorf("decode consumed upload snapshot: %w", err)
	}
	return result, nil
}

func scan(row pgx.Row) (productfiles.ProductFile, error) {
	var f productfiles.ProductFile
	err := row.Scan(&f.ID, &f.ProductID, &f.FileKey, &f.FileName, &f.DisplayName, &f.Format,
		&f.ContentType, &f.FileSizeBytes, &f.R2ETag, &f.AvailabilityStatus, &f.ConfirmationFingerprint,
		&f.FingerprintKeyVersion, &f.DisplayOrder, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func release(c *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = c.Release(ctx)
}

func mapError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return productfiles.ErrConflict
		case "23503":
			return productfiles.ErrReferenceNotFound
		case "23502", "23514", "22P02":
			return productfiles.ErrInvalidInput
		}
	}
	return fmt.Errorf("product files database operation: %w", err)
}

func (r *Repository) List(ctx context.Context, productID string) ([]productfiles.ProductFile, error) {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(c)
	rows, err := c.Query(ctx, `SELECT `+columns+` FROM product_files WHERE product_id=$1 ORDER BY display_order, id`, productID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	files := []productfiles.ProductFile{}
	for rows.Next() {
		f, err := scan(rows)
		if err != nil {
			return nil, mapError(err)
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}
	return files, nil
}

func (r *Repository) GetByKey(ctx context.Context, key string) (productfiles.ProductFile, error) {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return productfiles.ProductFile{}, err
	}
	defer release(c)
	f, err := scan(c.QueryRow(ctx, `SELECT `+columns+` FROM product_files WHERE file_key=$1`, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return productfiles.ProductFile{}, productfiles.ErrFileNotFound
	}
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	return f, nil
}

func (r *Repository) getConsumption(ctx context.Context, field, value string) (productfiles.Consumption, error) {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return productfiles.Consumption{}, err
	}
	defer release(c)
	query := `SELECT ` + consumptionColumns + ` FROM product_file_consumptions WHERE ` + field + `=$1`
	result, err := scanConsumption(c.QueryRow(ctx, query, value))
	if errors.Is(err, pgx.ErrNoRows) {
		return productfiles.Consumption{}, productfiles.ErrFileNotFound
	}
	if err != nil {
		return productfiles.Consumption{}, mapError(err)
	}
	return result, nil
}

func (r *Repository) GetConsumptionByDigest(ctx context.Context, digest string) (productfiles.Consumption, error) {
	return r.getConsumption(ctx, "token_digest", digest)
}

func (r *Repository) GetConsumptionByKey(ctx context.Context, key string) (productfiles.Consumption, error) {
	return r.getConsumption(ctx, "file_key", key)
}

func (r *Repository) GetByIDAndProduct(ctx context.Context, productID, fileID string) (productfiles.ProductFile, error) {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return productfiles.ProductFile{}, err
	}
	defer release(c)
	f, err := scan(c.QueryRow(ctx, `SELECT `+columns+` FROM product_files WHERE id=$1 AND product_id=$2`, fileID, productID))
	if errors.Is(err, pgx.ErrNoRows) {
		return productfiles.ProductFile{}, productfiles.ErrFileNotFound
	}
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	return f, nil
}

func (r *Repository) CreateConfirmed(ctx context.Context, input productfiles.ProductFile, consumption productfiles.Consumption) (productfiles.ProductFile, error) {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return productfiles.ProductFile{}, err
	}
	defer release(c)
	tx, err := c.Begin(ctx)
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := lockProduct(ctx, tx, input.ProductID); err != nil {
		return productfiles.ProductFile{}, err
	}
	var consumedKey string
	err = tx.QueryRow(ctx, `SELECT file_key FROM product_file_consumptions WHERE file_key=$1`, input.FileKey).Scan(&consumedKey)
	if err == nil {
		return productfiles.ProductFile{}, productfiles.ErrConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return productfiles.ProductFile{}, mapError(err)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM product_files WHERE product_id=$1`, input.ProductID).Scan(&count); err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	if input.DisplayOrder > count {
		input.DisplayOrder = count
	}
	if _, err := tx.Exec(ctx, `UPDATE product_files SET display_order=display_order+1 WHERE product_id=$1 AND display_order >= $2`, input.ProductID, input.DisplayOrder); err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	f, err := scan(tx.QueryRow(ctx, `INSERT INTO product_files
	 (product_id,file_key,file_name,display_name,format,content_type,file_size_bytes,r2_etag,
	 availability_status,confirmation_fingerprint,fingerprint_key_version,display_order)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING `+columns,
		input.ProductID, input.FileKey, input.FileName, input.DisplayName, input.Format, input.ContentType,
		input.FileSizeBytes, input.R2ETag, input.AvailabilityStatus, input.ConfirmationFingerprint,
		input.FingerprintKeyVersion, input.DisplayOrder))
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	if consumption.FileKey != f.FileKey || consumption.ProductID != f.ProductID ||
		consumption.ConfirmationFingerprint != f.ConfirmationFingerprint ||
		consumption.FingerprintKeyVersion != f.FingerprintKeyVersion || consumption.TokenDigest == "" || consumption.ExpiresAt == nil {
		return productfiles.ProductFile{}, productfiles.ErrInvalidInput
	}
	snapshot, err := json.Marshal(productfiles.ConfirmationSnapshot{ID: f.ID, ProductID: f.ProductID,
		FileName: f.FileName, DisplayName: f.DisplayName, Format: f.Format, ContentType: f.ContentType,
		FileSizeBytes: f.FileSizeBytes, AvailabilityStatus: f.AvailabilityStatus, DisplayOrder: f.DisplayOrder,
		CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt})
	if err != nil {
		return productfiles.ProductFile{}, fmt.Errorf("encode confirmation snapshot: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO product_file_consumptions
	 (file_key,product_id,product_file_id,token_digest,confirmation_fingerprint,
	 fingerprint_key_version,expires_at,response_snapshot)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, consumption.FileKey, consumption.ProductID,
		f.ID, consumption.TokenDigest, consumption.ConfirmationFingerprint,
		consumption.FingerprintKeyVersion, consumption.ExpiresAt, string(snapshot))
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	return f, nil
}

func (r *Repository) UpdateMetadata(ctx context.Context, input productfiles.UpdateMetadataInput) (productfiles.ProductFile, error) {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return productfiles.ProductFile{}, err
	}
	defer release(c)
	tx, err := c.Begin(ctx)
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := lockProduct(ctx, tx, input.ProductID); err != nil {
		return productfiles.ProductFile{}, err
	}
	ids, err := orderedIDs(ctx, tx, input.ProductID)
	if err != nil {
		return productfiles.ProductFile{}, err
	}
	current := -1
	for i, id := range ids {
		if id == input.FileID {
			current = i
			break
		}
	}
	if current < 0 {
		return productfiles.ProductFile{}, productfiles.ErrFileNotFound
	}
	ids = append(ids[:current], ids[current+1:]...)
	position := input.DisplayOrder
	if position > len(ids) {
		position = len(ids)
	}
	ids = append(ids, "")
	copy(ids[position+1:], ids[position:])
	ids[position] = input.FileID
	if err := updateOrder(ctx, tx, input.ProductID, ids); err != nil {
		return productfiles.ProductFile{}, err
	}
	f, err := scan(tx.QueryRow(ctx, `UPDATE product_files SET display_name=$1
	 WHERE id=$2 AND product_id=$3 RETURNING `+columns, input.DisplayName, input.FileID, input.ProductID))
	if errors.Is(err, pgx.ErrNoRows) {
		return productfiles.ProductFile{}, productfiles.ErrFileNotFound
	}
	if err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return productfiles.ProductFile{}, mapError(err)
	}
	return f, nil
}

func (r *Repository) DeleteByIDAndProduct(ctx context.Context, productID, fileID string) error {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(c)
	tx, err := c.Begin(ctx)
	if err != nil {
		return mapError(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := lockProduct(ctx, tx, productID); err != nil {
		return err
	}
	var currentKey string
	var currentFingerprint string
	var currentVersion int
	err = tx.QueryRow(ctx, `SELECT file_key, confirmation_fingerprint, fingerprint_key_version
	 FROM product_files WHERE id=$1 AND product_id=$2 FOR UPDATE`, fileID, productID).
		Scan(&currentKey, &currentFingerprint, &currentVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return productfiles.ErrFileNotFound
	}
	if err != nil {
		return mapError(err)
	}
	var consumedID string
	err = tx.QueryRow(ctx, `SELECT c.product_file_id::text
	 FROM product_file_consumptions c
	 JOIN product_files f ON f.id=$3 AND f.product_id=$2
	 WHERE c.file_key=$1 AND c.product_id=$2 AND c.product_file_id=$3
	 AND c.confirmation_fingerprint=$4 AND c.fingerprint_key_version=$5
	 AND c.response_snapshot->>'id'=$3::text AND c.response_snapshot->>'productId'=$2::text
	 AND c.response_snapshot->>'fileName'=f.file_name
	 AND c.response_snapshot->>'format'=f.format
	 AND c.response_snapshot->>'contentType'=f.content_type
	 AND c.response_snapshot->>'fileSizeBytes'=f.file_size_bytes::text
	 AND (c.response_snapshot->>'createdAt')::timestamptz=f.created_at`,
		currentKey, productID, fileID, currentFingerprint, currentVersion).Scan(&consumedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return productfiles.ErrConflict
	}
	if err != nil {
		return mapError(err)
	}
	tag, err := tx.Exec(ctx, `DELETE FROM product_files WHERE id=$1 AND product_id=$2`, fileID, productID)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return productfiles.ErrFileNotFound
	}
	mark, err := tx.Exec(ctx, `UPDATE product_file_consumptions SET deleted_at=COALESCE(deleted_at, now())
	 WHERE product_file_id=$1 AND file_key=$2`, fileID, currentKey)
	if err != nil {
		return mapError(err)
	}
	if mark.RowsAffected() != 1 {
		return productfiles.ErrConflict
	}
	ids, err := orderedIDs(ctx, tx, productID)
	if err != nil {
		return err
	}
	if err := updateOrder(ctx, tx, productID, ids); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return mapError(err)
	}
	return nil
}

func lockProduct(ctx context.Context, tx pgx.Tx, productID string) error {
	var id string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM products WHERE id=$1 FOR UPDATE`, productID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return productfiles.ErrReferenceNotFound
		}
		return mapError(err)
	}
	return nil
}

func orderedIDs(ctx context.Context, tx pgx.Tx, productID string) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT id::text FROM product_files WHERE product_id=$1 ORDER BY display_order,id`, productID)
	if err != nil {
		return nil, mapError(err)
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, mapError(err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}
	return ids, nil
}

func updateOrder(ctx context.Context, tx pgx.Tx, productID string, ids []string) error {
	for position, id := range ids {
		if _, err := tx.Exec(ctx, `UPDATE product_files SET display_order=$1 WHERE id=$2 AND product_id=$3 AND display_order IS DISTINCT FROM $1`, position, id, productID); err != nil {
			return mapError(err)
		}
	}
	return nil
}

func (r *Repository) Reorder(ctx context.Context, productID string, ids []string) error {
	c, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(c)
	tx, err := c.Begin(ctx)
	if err != nil {
		return mapError(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := lockProduct(ctx, tx, productID); err != nil {
		return err
	}
	stored, err := orderedIDs(ctx, tx, productID)
	if err != nil {
		return err
	}
	current := map[string]bool{}
	for _, id := range stored {
		current[id] = true
	}
	if len(current) != len(ids) {
		return productfiles.ErrInvalidInput
	}
	for _, id := range ids {
		if !current[id] {
			return productfiles.ErrInvalidInput
		}
	}
	if err := updateOrder(ctx, tx, productID, ids); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return mapError(err)
	}
	return nil
}
