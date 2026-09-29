// Package store provides SQLite-based time-series storage for cycle data.
package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/vmo/local-ems/internal/engine/cycle"
)

// CycleRecord is a flattened cycle result for storage.
type CycleRecord struct {
	Timestamp   time.Time
	SolarW      int
	LoadW       int
	GridW       int
	ESSW        int
	DesiredW    int
	SOC         float64
	TempC       float64
	Clamped     bool
	Constraints string
}

// Store provides buffered time-series storage backed by SQLite.
type Store struct {
	db   *sql.DB
	mu   sync.Mutex
	path string

	walMode       bool
	retentionDays int
	done          chan struct{}
	wg            sync.WaitGroup
}

// Option configures a Store.
type Option func(*Store)

// WithWAL enables SQLite WAL mode.
func WithWAL() Option {
	return func(s *Store) { s.walMode = true }
}

// WithAutoRetention enables automatic pruning of records older than the given days.
func WithAutoRetention(days int) Option {
	return func(s *Store) { s.retentionDays = days }
}

// New creates and initializes a new Store at the given file path.
// Use ":memory:" for in-memory storage (testing).
func New(path string, opts ...Option) (*Store, error) {
	s := &Store{
		path: path,
		done: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if s.walMode {
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			db.Close()
			return nil, fmt.Errorf("enable wal: %w", err)
		}
	}

	s.db = db

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cycles (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  TEXT    NOT NULL,
			solar_w    INTEGER NOT NULL,
			load_w     INTEGER NOT NULL,
			grid_w     INTEGER NOT NULL,
			ess_w      INTEGER NOT NULL,
			desired_w  INTEGER NOT NULL,
			soc        REAL    NOT NULL,
			temp_c     REAL    NOT NULL,
			clamped    INTEGER NOT NULL,
			constraints TEXT   NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_cycles_ts ON cycles(timestamp);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	if s.retentionDays > 0 {
		s.wg.Add(1)
		go s.autoPruneRoutine()
	}

	return s, nil
}

func (s *Store) autoPruneRoutine() {
	defer s.wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.Prune(time.Duration(s.retentionDays) * 24 * time.Hour)
		case <-s.done:
			return
		}
	}
}

// Insert writes a cycle result to the store.
func (s *Store) Insert(r CycleRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO cycles (timestamp, solar_w, load_w, grid_w, ess_w, desired_w, soc, temp_c, clamped, constraints)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Timestamp.Format(time.RFC3339),
		r.SolarW, r.LoadW, r.GridW, r.ESSW, r.DesiredW,
		r.SOC, r.TempC,
		boolToInt(r.Clamped), r.Constraints,
	)
	return err
}

// InsertCycleResult converts a CycleResult and inserts it.
func (s *Store) InsertCycleResult(cr cycle.CycleResult) error {
	constraints := ""
	if len(cr.ResolvedPower.AppliedConstraints) > 0 {
		constraints = fmt.Sprintf("%v", cr.ResolvedPower.AppliedConstraints)
	}
	return s.Insert(CycleRecord{
		Timestamp:   cr.Timestamp,
		SolarW:      cr.State.SolarPowerW,
		LoadW:       cr.State.LoadPowerW,
		GridW:       cr.State.GridPowerW,
		ESSW:        cr.ResolvedPower.Setpoint,
		DesiredW:    cr.DesiredPower,
		SOC:         cr.State.BatterySOC,
		TempC:       cr.State.BatteryTempC,
		Clamped:     cr.ResolvedPower.Clamped,
		Constraints: constraints,
	})
}

// QueryRecent returns the most recent N records.
func (s *Store) QueryRecent(limit int) ([]CycleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT timestamp, solar_w, load_w, grid_w, ess_w, desired_w, soc, temp_c, clamped, constraints
		FROM cycles ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CycleRecord
	for rows.Next() {
		var r CycleRecord
		var ts string
		var clamped int
		if err := rows.Scan(&ts, &r.SolarW, &r.LoadW, &r.GridW, &r.ESSW, &r.DesiredW,
			&r.SOC, &r.TempC, &clamped, &r.Constraints); err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse(time.RFC3339, ts)
		r.Clamped = clamped != 0
		records = append(records, r)
	}

	// Reverse to chronological order
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	return records, nil
}

// QueryByDate returns records for a specific date (YYYY-MM-DD).
func (s *Store) QueryByDate(date string) ([]CycleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT timestamp, solar_w, load_w, grid_w, ess_w, desired_w, soc, temp_c, clamped, constraints
		FROM cycles
		WHERE timestamp >= ? AND timestamp < date(?, '+1 day')
		ORDER BY id ASC`, date, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CycleRecord
	for rows.Next() {
		var r CycleRecord
		var ts string
		var clamped int
		if err := rows.Scan(&ts, &r.SolarW, &r.LoadW, &r.GridW, &r.ESSW, &r.DesiredW,
			&r.SOC, &r.TempC, &clamped, &r.Constraints); err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse(time.RFC3339, ts)
		r.Clamped = clamped != 0
		records = append(records, r)
	}

	return records, nil
}

// DailySummary returns aggregated stats for a given date.
func (s *Store) DailySummary(date string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.db.QueryRow(`
		SELECT
			COUNT(*) as cycles,
			COALESCE(MAX(solar_w), 0) as peak_solar,
			COALESCE(MAX(load_w), 0) as peak_load,
			COALESCE(MAX(grid_w), 0) as peak_grid_import,
			COALESCE(MIN(grid_w), 0) as peak_grid_export,
			COALESCE(AVG(soc), 0) as avg_soc,
			COALESCE(MIN(soc), 0) as min_soc,
			COALESCE(MAX(soc), 0) as max_soc,
			COALESCE(AVG(temp_c), 0) as avg_temp
		FROM cycles
		WHERE timestamp >= ? AND timestamp < date(?, '+1 day')`, date, date)

	var cycles int
	var peakSolar, peakLoad, peakGridImport, peakGridExport int
	var avgSOC, minSOC, maxSOC, avgTemp float64

	if err := row.Scan(&cycles, &peakSolar, &peakLoad, &peakGridImport, &peakGridExport,
		&avgSOC, &minSOC, &maxSOC, &avgTemp); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"date":             date,
		"total_cycles":     cycles,
		"peak_solar_w":     peakSolar,
		"peak_load_w":      peakLoad,
		"peak_grid_import": peakGridImport,
		"peak_grid_export": peakGridExport,
		"avg_soc":          avgSOC,
		"min_soc":          minSOC,
		"max_soc":          maxSOC,
		"avg_temp_c":       avgTemp,
	}, nil
}

// Count returns total record count.
func (s *Store) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM cycles").Scan(&count)
	return count, err
}

// Prune deletes records older than the given duration.
func (s *Store) Prune(olderThan time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan).Format(time.RFC3339)
	result, err := s.db.Exec("DELETE FROM cycles WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Close closes the database.
func (s *Store) Close() error {
	if s.done != nil {
		close(s.done)
		s.wg.Wait()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
