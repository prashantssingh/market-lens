package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", a.health)
	mux.HandleFunc("/api/state", a.state)
	mux.HandleFunc("/api/stocks", a.stocks)
	mux.HandleFunc("/api/stocks/", a.stockActions)
	mux.HandleFunc("/api/runs", a.runs)
	mux.HandleFunc("/api/runs/", a.runStatus)
	mux.HandleFunc("/api/snapshots", a.snapshots)
	return mux
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (a *App) state(w http.ResponseWriter, r *http.Request) {
	stocks, err := a.listStocks()
	if err != nil {
		writeError(w, err)
		return
	}
	analyses, err := a.listAnalyses()
	if err != nil {
		writeError(w, err)
		return
	}
	snapshots, err := a.listSnapshots()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{
		"stocks":    stocks,
		"analyses":  analyses,
		"snapshots": snapshots,
	})
}

func (a *App) stocks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		stocks, err := a.listStocks()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, stocks)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Symbol string `json:"symbol"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	stock, err := a.upsertStock(strings.ToUpper(strings.TrimSpace(body.Symbol)))
	if err != nil {
		writeError(w, err)
		return
	}
	analysis, err := a.runAnalysis(stock)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, analysis)
}

func (a *App) stockActions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/stocks/"), "/")
	if len(parts) != 2 || parts[1] != "analyze" || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stock, err := a.upsertStock(strings.ToUpper(strings.TrimSpace(parts[0])))
	if err != nil {
		writeError(w, err)
		return
	}
	analysis, err := a.runAnalysis(stock)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, analysis)
}

func (a *App) runs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Symbol string `json:"symbol"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	stock, err := a.upsertStock(strings.ToUpper(strings.TrimSpace(body.Symbol)))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, a.startAnalysisRun(stock))
}

func (a *App) runStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	run, ok := a.getRun(id)
	if !ok {
		writeErrorStatus(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, run)
}

func (a *App) snapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		snapshots, err := a.listSnapshots()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, snapshots)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	snapshot, err := a.createSnapshot(body.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, snapshot)
}
