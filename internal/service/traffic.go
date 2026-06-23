package service

import (
	"time"

	nexusdb "nexus/internal/database"
	nexusmodel "nexus/internal/model"

	"gorm.io/gorm"
)

// TrafficEntry is a thin DTO that decouples the service layer from the
// protobuf-generated types.
type TrafficEntry struct {
	UserUUID string
	Upload   uint64
	Download uint64
}

// RecordTraffic performs a batch insert of traffic log rows and increments
// each affected user traffic_used counter in a single transaction.
func RecordTraffic(nodeID uint, entries []TrafficEntry) error {
	if len(entries) == 0 {
		return nil
	}

	return nexusdb.DB.Transaction(func(tx *gorm.DB) error {
		for _, e := range entries {
			// Resolve the user by UUID.
			var user nexusmodel.User
			if err := tx.Where("uuid = ?", e.UserUUID).First(&user).Error; err != nil {
				// Skip unknown users instead of aborting the whole batch.
				continue
			}

			// Insert a traffic log row.
			logEntry := nexusmodel.TrafficLog{
				UserID:     user.ID,
				NodeID:     nodeID,
				Upload:     int64(e.Upload),
				Download:   int64(e.Download),
				RecordedAt: time.Now(),
			}
			if err := tx.Create(&logEntry).Error; err != nil {
				return err
			}

			// Increment the user cumulative traffic counter.
			if err := tx.Model(&nexusmodel.User{}).
				Where("id = ?", user.ID).
				UpdateColumn("traffic_used", gorm.Expr("traffic_used + ?", e.Upload+e.Download)).
				Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// TrafficStatsRow is a single row returned by GetTrafficStats.
type TrafficStatsRow struct {
	Date       string `json:"date"`
	Upload     int64  `json:"upload"`
	Download   int64  `json:"download"`
	Total      int64  `json:"total"`
	UniqueUser int64  `json:"unique_users"`
}

// GetTrafficStats aggregates traffic over the last N days, grouped by day.
func GetTrafficStats(days int) ([]TrafficStatsRow, error) {
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	var rows []TrafficStatsRow
	err := nexusdb.DB.Raw(`
		SELECT
			date(recorded_at)                       AS date,
			COALESCE(SUM(upload), 0)                AS upload,
			COALESCE(SUM(download), 0)              AS download,
			COALESCE(SUM(upload + download), 0)     AS total,
			COUNT(DISTINCT user_id)                 AS unique_users
		FROM traffic_logs
		WHERE recorded_at >= ?
		GROUP BY date(recorded_at)
		ORDER BY date(recorded_at) DESC
	`, since).Scan(&rows).Error
	return rows, err
}
