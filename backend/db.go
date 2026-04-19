package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) migrate() error {
	schema := `
	PRAGMA journal_mode = WAL;
	CREATE TABLE IF NOT EXISTS stocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS analyses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		stock_id INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		signal TEXT NOT NULL,
		confidence TEXT NOT NULL,
		score INTEGER NOT NULL,
		payload TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS worker_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		stock_id INTEGER NOT NULL,
		worker TEXT NOT NULL,
		status TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		note TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := a.db.Exec(schema)
	return err
}

func (a *App) upsertStock(symbol string) (Stock, error) {
	_, err := a.db.Exec(`INSERT INTO stocks(symbol, name) VALUES(?, ?)
		ON CONFLICT(symbol) DO UPDATE SET name = excluded.name`, symbol, companyName(symbol))
	if err != nil {
		return Stock{}, err
	}
	var stock Stock
	err = a.db.QueryRow(`SELECT id, symbol, name, created_at FROM stocks WHERE symbol = ?`, symbol).
		Scan(&stock.ID, &stock.Symbol, &stock.Name, &stock.CreatedAt)
	return stock, err
}

func (a *App) listStocks() ([]Stock, error) {
	rows, err := a.db.Query(`SELECT id, symbol, name, created_at FROM stocks ORDER BY symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []Stock
	for rows.Next() {
		var stock Stock
		if err := rows.Scan(&stock.ID, &stock.Symbol, &stock.Name, &stock.CreatedAt); err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}
	return stocks, rows.Err()
}

func (a *App) listAnalyses() ([]Analysis, error) {
	rows, err := a.db.Query(`SELECT id, payload FROM analyses ORDER BY created_at DESC LIMIT 80`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var analyses []Analysis
	for rows.Next() {
		var id int64
		var payload string
		if err := rows.Scan(&id, &payload); err != nil {
			return nil, err
		}
		var analysis Analysis
		if err := json.Unmarshal([]byte(payload), &analysis); err != nil {
			continue
		}
		analysis.ID = id
		analyses = append(analyses, analysis)
	}
	return analyses, rows.Err()
}

func (a *App) listSnapshots() ([]Snapshot, error) {
	rows, err := a.db.Query(`SELECT id, file_name, note, created_at FROM snapshots ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []Snapshot
	for rows.Next() {
		var snapshot Snapshot
		if err := rows.Scan(&snapshot.ID, &snapshot.FileName, &snapshot.Note, &snapshot.CreatedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, rows.Err()
}

func (a *App) createSnapshot(note string) (Snapshot, error) {
	fileName := "market-lens-" + time.Now().Format("20060102-150405") + ".db"
	path := filepath.Join(a.dataDir, "snapshots", fileName)
	escaped := strings.ReplaceAll(path, "'", "''")
	if _, err := a.db.Exec("VACUUM INTO '" + escaped + "'"); err != nil {
		return Snapshot{}, err
	}
	res, err := a.db.Exec(`INSERT INTO snapshots(file_name, note) VALUES(?, ?)`, fileName, note)
	if err != nil {
		return Snapshot{}, err
	}
	id, _ := res.LastInsertId()
	return Snapshot{ID: id, FileName: fileName, Note: note, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}

func (a *App) persistWorkerEvent(stockID int64, name, status, message string) WorkerEvent {
	res, _ := a.db.Exec(`INSERT INTO worker_events(stock_id, worker, status, message) VALUES(?, ?, ?, ?)`,
		stockID, name, status, message)
	id, _ := res.LastInsertId()
	return WorkerEvent{
		ID: id, StockID: stockID, Worker: name, Status: status, Message: message,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}
