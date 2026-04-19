package main

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

func (a *App) startAnalysisRun(stock Stock) AnalysisRun {
	now := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("run-%d-%d", time.Now().UnixNano(), atomic.AddInt64(&a.jobSeq, 1))
	run := &AnalysisRun{
		ID:        id,
		Symbol:    stock.Symbol,
		Status:    "running",
		Workers:   initialWorkflow(stock.ID),
		StartedAt: now,
		UpdatedAt: now,
	}

	a.jobsMu.Lock()
	a.jobs[id] = run
	a.jobsMu.Unlock()

	go a.runAnalysisWorkflow(id, stock)
	return *run
}

func (a *App) getRun(id string) (AnalysisRun, bool) {
	a.jobsMu.RLock()
	defer a.jobsMu.RUnlock()
	run, ok := a.jobs[id]
	if !ok {
		return AnalysisRun{}, false
	}
	return *run, true
}

func initialWorkflow(stockID int64) []WorkerEvent {
	now := time.Now().Format(time.RFC3339)
	return []WorkerEvent{
		{StockID: stockID, Worker: "quote", Status: "pending", Message: "Waiting to fetch quote and price movement", CreatedAt: now},
		{StockID: stockID, Worker: "news", Status: "pending", Message: "Waiting to collect recent catalysts", CreatedAt: now},
		{StockID: stockID, Worker: "sec-filings", Status: "pending", Message: "Waiting to check recent SEC filings", CreatedAt: now},
		{StockID: stockID, Worker: "sentiment", Status: "pending", Message: "Waiting to score signal inputs", CreatedAt: now},
		{StockID: stockID, Worker: "summary", Status: "pending", Message: "Waiting to compose final summary", CreatedAt: now},
	}
}

func (a *App) runAnalysis(stock Stock) (Analysis, error) {
	var events []WorkerEvent
	addWorker := func(name, status, message string) {
		events = append(events, a.persistWorkerEvent(stock.ID, name, status, message))
	}

	addWorker("quote", "running", "Fetching quote and price movement")
	quote := a.getQuote(stock.Symbol)
	addWorker("quote", "complete", fmt.Sprintf("Quote loaded from %s", quote.Source))

	addWorker("news", "running", "Collecting recent catalysts")
	news := a.getNews(stock.Symbol)
	addWorker("news", "complete", fmt.Sprintf("%d news items available", len(news)))

	addWorker("sec-filings", "running", "Checking recent SEC filings")
	filings := a.getFilings(stock.Symbol)
	if len(filings) == 0 {
		addWorker("sec-filings", "stale", "No SEC filing data available in this run")
	} else {
		addWorker("sec-filings", "complete", fmt.Sprintf("%d filings available", len(filings)))
	}

	addWorker("sentiment", "running", "Scoring simple news and price signal")
	score, signal, confidence, reasons, warnings := scoreAnalysis(quote, news, filings)
	addWorker("sentiment", "complete", fmt.Sprintf("Signal scored as %s", signal))
	addWorker("summary", "complete", "Timeline entry created")

	return a.saveAnalysis(stock, quote, news, filings, events, score, signal, confidence, reasons, warnings)
}

func (a *App) runAnalysisWorkflow(runID string, stock Stock) {
	var events []WorkerEvent
	updateStage := func(name, status, message string) {
		event := a.persistWorkerEvent(stock.ID, name, status, message)
		events = append(events, event)
		a.updateRunStage(runID, name, status, message, event.ID, event.CreatedAt)
	}
	failRun := func(err error) {
		a.jobsMu.Lock()
		if run, ok := a.jobs[runID]; ok {
			run.Status = "failed"
			run.Error = err.Error()
			run.UpdatedAt = time.Now().Format(time.RFC3339)
		}
		a.jobsMu.Unlock()
	}

	updateStage("quote", "running", "Fetching quote and price movement")
	quote := a.getQuote(stock.Symbol)
	updateStage("quote", "complete", fmt.Sprintf("Quote loaded from %s", quote.Source))
	workflowPause()

	updateStage("news", "running", "Collecting recent catalysts")
	news := a.getNews(stock.Symbol)
	updateStage("news", "complete", fmt.Sprintf("%d news items available", len(news)))
	workflowPause()

	updateStage("sec-filings", "running", "Checking recent SEC filings")
	filings := a.getFilings(stock.Symbol)
	if len(filings) == 0 {
		updateStage("sec-filings", "stale", "No SEC filing data available in this run")
	} else {
		updateStage("sec-filings", "complete", fmt.Sprintf("%d filings available", len(filings)))
	}
	workflowPause()

	updateStage("sentiment", "running", "Scoring simple news, filing, and price signals")
	score, signal, confidence, reasons, warnings := scoreAnalysis(quote, news, filings)
	updateStage("sentiment", "complete", fmt.Sprintf("Signal scored as %s", signal))
	workflowPause()

	updateStage("summary", "running", "Composing final research summary")
	analysis, err := a.saveAnalysis(stock, quote, news, filings, events, score, signal, confidence, reasons, warnings)
	if err != nil {
		updateStage("summary", "failed", "Could not save final timeline entry")
		failRun(err)
		return
	}
	completeEvent := a.persistWorkerEvent(stock.ID, "summary", "complete", "Timeline entry created")
	analysis.Workers = append(analysis.Workers, completeEvent)
	_ = a.updateAnalysisPayload(analysis)

	a.jobsMu.Lock()
	if run, ok := a.jobs[runID]; ok {
		run.Status = "complete"
		run.Analysis = &analysis
		run.UpdatedAt = time.Now().Format(time.RFC3339)
		for i := range run.Workers {
			if run.Workers[i].Worker == "summary" {
				run.Workers[i].Status = "complete"
				run.Workers[i].Message = "Timeline entry created"
				run.Workers[i].CreatedAt = completeEvent.CreatedAt
			}
		}
	}
	a.jobsMu.Unlock()
}

func (a *App) updateRunStage(runID, name, status, message string, eventID int64, createdAt string) {
	a.jobsMu.Lock()
	defer a.jobsMu.Unlock()
	run, ok := a.jobs[runID]
	if !ok {
		return
	}
	for i := range run.Workers {
		if run.Workers[i].Worker == name {
			run.Workers[i].ID = eventID
			run.Workers[i].Status = status
			run.Workers[i].Message = message
			run.Workers[i].CreatedAt = createdAt
			break
		}
	}
	run.UpdatedAt = time.Now().Format(time.RFC3339)
}

func (a *App) saveAnalysis(stock Stock, quote Quote, news []NewsItem, filings []Filing, workers []WorkerEvent, score int, signal, confidence string, reasons, warnings []string) (Analysis, error) {
	analysis := Analysis{
		StockID: stock.ID, Symbol: stock.Symbol, Signal: signal, Confidence: confidence, Score: score,
		Reasons: reasons, Warnings: warnings, Quote: quote, Chart: makeChart(quote), News: news,
		Filings: filings, Workers: workers, CreatedAt: time.Now().Format(time.RFC3339),
	}
	payload, err := json.Marshal(analysis)
	if err != nil {
		return Analysis{}, err
	}
	res, err := a.db.Exec(`INSERT INTO analyses(stock_id, symbol, signal, confidence, score, payload)
		VALUES(?, ?, ?, ?, ?, ?)`, stock.ID, stock.Symbol, signal, confidence, score, string(payload))
	if err != nil {
		return Analysis{}, err
	}
	analysis.ID, _ = res.LastInsertId()
	return analysis, a.updateAnalysisPayload(analysis)
}

func (a *App) updateAnalysisPayload(analysis Analysis) error {
	payload, err := json.Marshal(analysis)
	if err != nil {
		return err
	}
	_, err = a.db.Exec(`UPDATE analyses SET payload = ? WHERE id = ?`, string(payload), analysis.ID)
	return err
}

func workflowPause() {
	time.Sleep(180 * time.Millisecond)
}
