package main

import (
	"database/sql"
	"sync"
)

type App struct {
	db       *sql.DB
	dataDir  string
	alphaKey string
	jobs     map[string]*AnalysisRun
	jobsMu   sync.RWMutex
	jobSeq   int64
}

type Stock struct {
	ID        int64  `json:"id"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type Quote struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Source        string  `json:"source"`
	Freshness     string  `json:"freshness"`
	ObservedAt    string  `json:"observedAt"`
}

type ChartPoint struct {
	Time  string  `json:"time"`
	Price float64 `json:"price"`
}

type NewsItem struct {
	Title     string `json:"title"`
	Source    string `json:"source"`
	URL       string `json:"url"`
	Published string `json:"published"`
	Label     string `json:"label"`
	Catalyst  string `json:"catalyst"`
}

type Filing struct {
	Form      string `json:"form"`
	Date      string `json:"date"`
	Report    string `json:"report"`
	URL       string `json:"url"`
	Highlight bool   `json:"highlight"`
}

type WorkerEvent struct {
	ID        int64  `json:"id"`
	StockID   int64  `json:"stockId"`
	Worker    string `json:"worker"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type Analysis struct {
	ID         int64         `json:"id"`
	StockID    int64         `json:"stockId"`
	Symbol     string        `json:"symbol"`
	Signal     string        `json:"signal"`
	Confidence string        `json:"confidence"`
	Score      int           `json:"score"`
	Reasons    []string      `json:"reasons"`
	Warnings   []string      `json:"warnings"`
	Quote      Quote         `json:"quote"`
	Chart      []ChartPoint  `json:"chart"`
	News       []NewsItem    `json:"news"`
	Filings    []Filing      `json:"filings"`
	Workers    []WorkerEvent `json:"workers"`
	CreatedAt  string        `json:"createdAt"`
}

type Snapshot struct {
	ID        int64  `json:"id"`
	FileName  string `json:"fileName"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
}

type AnalysisRun struct {
	ID        string        `json:"id"`
	Symbol    string        `json:"symbol"`
	Status    string        `json:"status"`
	Workers   []WorkerEvent `json:"workers"`
	Analysis  *Analysis     `json:"analysis,omitempty"`
	Error     string        `json:"error,omitempty"`
	StartedAt string        `json:"startedAt"`
	UpdatedAt string        `json:"updatedAt"`
}

func newApp(db *sql.DB, dataDir string, alphaKey string) *App {
	return &App{
		db:       db,
		dataDir:  dataDir,
		alphaKey: alphaKey,
		jobs:     map[string]*AnalysisRun{},
	}
}
