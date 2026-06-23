package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// SearchReceivers lists every transferable user wallet, with optional search.
func (t *TransactionRepo) SearchReceivers(
	ctx context.Context,
	callerEmail,
	query string,
	page,
	limit int,
) ([]model.ReceiverResult, int, error) {
	offset := (page - 1) * limit

	baseSQL := `
		FROM users u
		LEFT JOIN profiles p ON p.user_id = u.id
		JOIN wallets w ON w.user_id = u.id
		WHERE u.email != $1
	`

	args := []any{callerEmail}
	conditions := make([]string, 0)
	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if query != "" {
		placeholder := addArg("%" + query + "%")
		conditions = append(conditions, `(
			COALESCE(p.full_name, '') ILIKE `+placeholder+` OR
			COALESCE(p.phone, '') ILIKE `+placeholder+` OR
			u.email ILIKE `+placeholder+` OR
			w.label ILIKE `+placeholder+` OR
			w.id::text ILIKE `+placeholder+`
		)`)
	}

	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " AND " + strings.Join(conditions, " AND ")
	}

	var total int
	countSQL := `SELECT COUNT(*) ` + baseSQL + whereSQL
	if err := t.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("SearchReceivers count: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, limit, offset)
	dataSQL := `
		SELECT u.id, u.email, p.full_name, p.phone, p.photo, w.id, w.label
	` + baseSQL + whereSQL + fmt.Sprintf(`
		ORDER BY COALESCE(p.full_name, u.email) ASC, w.label ASC
		LIMIT $%d OFFSET $%d`, len(queryArgs)-1, len(queryArgs))

	rows, err := t.db.Query(ctx, dataSQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("SearchReceivers query: %w", err)
	}
	defer rows.Close()

	var results []model.ReceiverResult
	for rows.Next() {
		var r model.ReceiverResult
		if err := rows.Scan(
			&r.UserID,
			&r.Email,
			&r.FullName,
			&r.Phone,
			&r.Photo,
			&r.WalletID,
			&r.WalletLabel,
		); err != nil {
			return nil, 0, fmt.Errorf("SearchReceivers scan: %w", err)
		}

		results = append(results, r)
	}

	return results, total, rows.Err()
}
