package database

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/Ahmet-Vt/fraud-detection/internal/domain"
	"gorm.io/gorm"
)

type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Version   string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Name      string `gorm:"type:varchar(255);not null"`
	AppliedAt time.Time
}

type MigrationEntry struct {
	Version string
	Name    string
	Up      func(tx *gorm.DB) error
	Down    func(tx *gorm.DB) error
}

type Migrator struct {
	db         *gorm.DB
	migrations []MigrationEntry
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Register(entries ...MigrationEntry) error {
	seen := make(map[string]bool)
	for _, e := range m.migrations {
		seen[e.Version] = true
	}

	for _, e := range entries {
		if e.Version == "" || e.Name == "" {
			return fmt.Errorf("migration entry must have non-empty Version and Name")
		}
		if e.Up == nil {
			return fmt.Errorf("migration %s_%s must have a non-nil Up function", e.Version, e.Name)
		}
		if seen[e.Version] {
			return fmt.Errorf("duplicate migration version: %s", e.Version)
		}
		seen[e.Version] = true
		m.migrations = append(m.migrations, e)
	}
	return nil
}

func (m *Migrator) Run() error {
	if err := m.db.Exec(`SELECT pg_advisory_lock(1)`).Error; err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer m.db.Exec(`SELECT pg_advisory_unlock(1)`)

	if err := m.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	appliedMap, err := m.appliedVersions()
	if err != nil {
		return err
	}

	for _, entry := range m.migrations {
		if appliedMap[entry.Version] {
			continue
		}

		log.Printf("[migration] applying %s_%s ...", entry.Version, entry.Name)

		err := m.db.Transaction(func(tx *gorm.DB) error {
			if err := entry.Up(tx); err != nil {
				return err
			}
			return tx.Create(&Migration{
				Version:   entry.Version,
				Name:      entry.Name,
				AppliedAt: time.Now(),
			}).Error
		})
		if err != nil {
			return fmt.Errorf("migration %s_%s failed: %w", entry.Version, entry.Name, err)
		}

		log.Printf("[migration] applied %s_%s", entry.Version, entry.Name)
	}
	return nil
}

func (m *Migrator) Rollback(steps int) error {
	if steps <= 0 {
		return fmt.Errorf("rollback steps must be > 0, got %d", steps)
	}

	if err := m.db.Exec(`SELECT pg_advisory_lock(1)`).Error; err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer m.db.Exec(`SELECT pg_advisory_unlock(1)`)

	var applied []Migration
	if err := m.db.Order("version DESC").Limit(steps).Find(&applied).Error; err != nil {
		return err
	}

	entryMap := make(map[string]MigrationEntry)
	for _, e := range m.migrations {
		entryMap[e.Version] = e
	}

	for _, a := range applied {
		entry, ok := entryMap[a.Version]
		if !ok {
			return fmt.Errorf("migration %s not found in registry", a.Version)
		}
		if entry.Down == nil {
			return fmt.Errorf("migration %s_%s has no Down function, cannot rollback", a.Version, a.Name)
		}

		log.Printf("[migration] rolling back %s_%s ...", a.Version, a.Name)

		err := m.db.Transaction(func(tx *gorm.DB) error {
			if err := entry.Down(tx); err != nil {
				return err
			}
			return tx.Where("version = ?", a.Version).Delete(&Migration{}).Error
		})
		if err != nil {
			return fmt.Errorf("rollback %s_%s failed: %w", a.Version, a.Name, err)
		}

		log.Printf("[migration] rolled back %s_%s", a.Version, a.Name)
	}
	return nil
}

func (m *Migrator) appliedVersions() (map[string]bool, error) {
	var applied []Migration
	if err := m.db.Order("version ASC").Find(&applied).Error; err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, a := range applied {
		result[a.Version] = true
	}
	return result, nil
}

func All() []MigrationEntry {
	return []MigrationEntry{}
}
