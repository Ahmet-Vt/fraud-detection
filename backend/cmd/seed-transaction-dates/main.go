package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"fraud-detection/internal/config"
	"fraud-detection/internal/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func main() {
	userStr := flag.String("user", "", "user id (UUID) — required")
	n := flag.Int("n", 20, "how many latest transactions (by created_at) to update")
	days := flag.Float64("days", 14, "spread from (now - days) to now (UTC)")
	dryRun := flag.Bool("dry-run", false, "print planned updates only")
	flag.Parse()

	if *userStr == "" {
		log.Fatal("missing -user <uuid>")
	}
	uid, err := uuid.Parse(*userStr)
	if err != nil {
		log.Fatalf("invalid -user: %v", err)
	}
	if *n < 1 {
		log.Fatal("-n must be at least 1")
	}
	if *days <= 0 {
		log.Fatal("-days must be positive")
	}

	cfg := config.Load()
	db, err := database.NewPostgres(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("sql db: %v", err)
	}
	defer sqlDB.Close()

	var ids []uuid.UUID
	if err := db.Table("transactions").
		Where("user_id = ?", uid).
		Order("created_at DESC").
		Limit(*n).
		Pluck("id", &ids).Error; err != nil {
		log.Fatalf("list transactions: %v", err)
	}

	if len(ids) == 0 {
		log.Printf("no transactions for user %s", uid)
		return
	}

	end := time.Now().UTC()
	start := end.Add(-time.Duration(*days * float64(24*time.Hour)))

	timestamps := spreadTimestamps(start, end, len(ids))
	if *dryRun {
		for i := range ids {
			fmt.Printf("%s -> %s\n", ids[i], timestamps[i].Format(time.RFC3339))
		}
		return
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for i := range ids {
			if err := tx.Exec(
				`UPDATE transactions SET created_at = ? WHERE id = ? AND user_id = ?`,
				timestamps[i], ids[i], uid,
			).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("update: %v", err)
	}

	log.Printf("updated %d transaction(s) for user %s", len(ids), uid)
}

func spreadTimestamps(start, end time.Time, n int) []time.Time {
	out := make([]time.Time, n)
	if n == 1 {
		out[0] = end
		return out
	}
	delta := end.Sub(start)
	for i := 0; i < n; i++ {
		t := end.Add(-time.Duration(float64(delta) * float64(i) / float64(n-1)))
		out[i] = t
	}
	return out
}
