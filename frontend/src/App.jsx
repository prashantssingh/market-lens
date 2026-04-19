import React, { useEffect, useMemo, useState } from 'react';
import { createSnapshot, fetchRun, fetchState, startRun } from './api.js';
import {
  AnalysisCommand,
  Chart,
  FeedPanel,
  StockOverview,
  SummaryPanel,
  Watchlist,
  WorkflowRail,
} from './components.jsx';

export default function App() {
  const [stocks, setStocks] = useState([]);
  const [analyses, setAnalyses] = useState([]);
  const [snapshots, setSnapshots] = useState([]);
  const [selectedSymbol, setSelectedSymbol] = useState('');
  const [symbolInput, setSymbolInput] = useState('');
  const [activeRun, setActiveRun] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function loadState(nextSymbol = selectedSymbol) {
    const state = await fetchState();
    setStocks(state.stocks || []);
    setAnalyses(state.analyses || []);
    setSnapshots(state.snapshots || []);
    if (!nextSymbol && state.stocks?.length) {
      setSelectedSymbol(state.stocks[0].symbol);
    }
  }

  useEffect(() => {
    loadState().catch((err) => setError(err.message));
  }, []);

  useEffect(() => {
    if (!activeRun || activeRun.status !== 'running') return undefined;
    const timer = window.setInterval(async () => {
      try {
        const run = await fetchRun(activeRun.id);
        setActiveRun(run);
        if (run.status === 'complete') {
          setSelectedSymbol(run.symbol);
          await loadState(run.symbol);
          setLoading(false);
        }
        if (run.status === 'failed') {
          setError(run.error || 'Analysis failed');
          setLoading(false);
        }
      } catch (err) {
        setError(err.message);
        setLoading(false);
      }
    }, 550);
    return () => window.clearInterval(timer);
  }, [activeRun]);

  const selectedAnalysis = useMemo(() => {
    if (activeRun?.analysis) return activeRun.analysis;
    return analyses.find((item) => item.symbol === selectedSymbol) || analyses[0];
  }, [activeRun, analyses, selectedSymbol]);

  const selectedStock = useMemo(() => {
    return stocks.find((stock) => stock.symbol === (selectedSymbol || selectedAnalysis?.symbol));
  }, [stocks, selectedSymbol, selectedAnalysis]);

  const timeline = useMemo(() => {
    const symbol = selectedSymbol || selectedAnalysis?.symbol;
    return analyses.filter((item) => !symbol || item.symbol === symbol);
  }, [analyses, selectedSymbol, selectedAnalysis]);

  async function runAnalysis(event) {
    event.preventDefault();
    const symbol = symbolInput.trim() || selectedSymbol || selectedAnalysis?.symbol;
    if (!symbol) return;
    setLoading(true);
    setError('');
    const run = await startRun(symbol);
    setActiveRun(run);
    setSelectedSymbol(run.symbol);
    setSymbolInput('');
  }

  async function snapshot() {
    setLoading(true);
    setError('');
    try {
      await createSnapshot();
      await loadState();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="app-shell">
      <header className="top-bar">
        <div>
          <p className="eyebrow">one-view research workflow</p>
          <h1>Market Lens</h1>
        </div>
        <span className="status-pill">SQLite persisted · Docker volume</span>
      </header>

      {error && <div className="error-banner">{error}</div>}

      <section className="single-view">
        <Watchlist
          stocks={stocks}
          analyses={analyses}
          selectedSymbol={selectedSymbol}
          onSelect={setSelectedSymbol}
        />

        <section className="analysis-board panel">
          <AnalysisCommand
            value={symbolInput}
            loading={loading}
            onChange={setSymbolInput}
            onSubmit={runAnalysis}
            onSnapshot={snapshot}
          />
          <StockOverview analysis={selectedAnalysis} stock={selectedStock} loading={loading} />
          <Chart points={selectedAnalysis?.chart || []} />
          <div className="intel-grid">
            <SummaryPanel analysis={selectedAnalysis} />
            <FeedPanel title="News / Catalysts" items={selectedAnalysis?.news || []} type="news" />
            <FeedPanel title="SEC Filings" items={selectedAnalysis?.filings || []} type="filing" />
          </div>
        </section>

        <WorkflowRail
          run={activeRun}
          fallbackWorkers={selectedAnalysis?.workers || []}
          timeline={timeline}
          snapshots={snapshots}
        />
      </section>
    </main>
  );
}
