package server

import (
	"database/sql"
	"time"
)

func CleanupInactiveGuests(db *sql.DB, inactiveFor time.Duration) error {
	cutoff := time.Now().Add(-inactiveFor)

	_, _ = db.Exec(`
		DELETE FROM sessions
		WHERE expires_at < ?
	`, time.Now())

	_, err := db.Exec(`
		DELETE FROM users
		WHERE is_guest = 1
		AND (
			last_seen_at IS NULL
			OR last_seen_at < ?
		)
	`, cutoff)

	return err
}
