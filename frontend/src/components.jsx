import React from 'react';
import { formatDate, money, signalClass, signed } from './format.js';

export function Watchlist({ stocks, analyses, selectedSymbol, onSelect }) {
  return (
    <aside className="watchlist panel">
      <div className="panel-header">
        <h2>Watchlist</h2>
        <span>{stocks.length}</span>
      </div>
      <div className="stock-list">
        {stocks.map((stock) => {
          const latest = analyses.find((item) => item.symbol === stock.symbol);
          return (
            <button
              key={stock.id}
              className={`stock-row ${selectedSymbol === stock.symbol ? 'active' : ''}`}
              onClick={() => onSelect(stock.symbol)}
            >
              <span>
                <strong>{stock.symbol}</strong>
                <small>{stock.name}</small>
              </span>
              <span className={`mini-signal ${signalClass(latest?.signal)}`}>
                {latest?.signal || 'new'}
              </span>
            </button>
          );
        })}
        {stocks.length === 0 && <p className="empty">No stocks yet. Run an analysis to begin.</p>}
      </div>
    </aside>
  );
}

export function AnalysisCommand({ value, loading, onChange, onSubmit, onSnapshot }) {
  return (
    <form onSubmit={onSubmit} className="command-bar">
      <input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="Enter ticker, e.g. AAPL"
        aria-label="Ticker"
      />
      <button disabled={loading}>{loading ? 'Analyzing...' : 'Run Analysis'}</button>
      <button type="button" onClick={onSnapshot} disabled={loading}>Snapshot</button>
    </form>
  );
}

export function StockOverview({ analysis, stock, loading }) {
  if (!analysis) {
    return (
      <section className="overview empty-overview">
        <div>
          <p className="eyebrow">ready</p>
          <h2>Run your first analysis</h2>
          <p>Market Lens will process quote, news, SEC filings, sentiment, and a final summary.</p>
        </div>
      </section>
    );
  }

  return (
    <section className="overview">
      <div>
        <p className="eyebrow">selected stock</p>
        <h2>{analysis.symbol}</h2>
        <p>{stock?.name || analysis.symbol}</p>
      </div>
      <div className={`signal-score ${signalClass(analysis.signal)}`}>
        <span>{analysis.signal}</span>
        <strong>{analysis.score}/100</strong>
      </div>
      <div className="quote-strip compact">
        <Metric label="Price" value={`$${money(analysis.quote.price)}`} />
        <Metric label="Move" value={`${signed(analysis.quote.changePercent)}%`} tone={analysis.quote.changePercent >= 0 ? 'positive' : 'negative'} />
        <Metric label="Source" value={analysis.quote.source} />
        <Metric label="Freshness" value={analysis.quote.freshness} />
      </div>
      {loading && <div className="processing-ribbon"><Spinner /> Processing workflow</div>}
    </section>
  );
}

function Metric({ label, value, tone = '' }) {
  return (
    <div className={tone}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

export function Chart({ points }) {
  if (!points?.length) return <div className="chart empty">No chart data yet.</div>;
  const width = 760;
  const height = 220;
  const prices = points.map((point) => point.price);
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  const range = max - min || 1;
  const path = points.map((point, index) => {
    const x = (index / (points.length - 1)) * width;
    const y = height - ((point.price - min) / range) * (height - 28) - 14;
    return `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
  }).join(' ');

  return (
    <div className="chart">
      <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Price chart">
        <defs>
          <linearGradient id="chartFill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="#1ed59f" stopOpacity="0.34" />
            <stop offset="100%" stopColor="#1ed59f" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={`${path} L ${width} ${height} L 0 ${height} Z`} fill="url(#chartFill)" />
        <path d={path} fill="none" stroke="#1ed59f" strokeWidth="3" strokeLinecap="round" />
      </svg>
      <div className="chart-labels">
        <span>{points[0].time}</span>
        <span>{points[points.length - 1].time}</span>
      </div>
    </div>
  );
}

export function SummaryPanel({ analysis }) {
  if (!analysis) return <Panel title="Summary"><p className="empty">No summary yet.</p></Panel>;
  return (
    <Panel title="Final Summary" meta={`Confidence: ${analysis.confidence}`}>
      <ul className="reason-list">
        {analysis.reasons.map((reason) => <li key={reason}>{reason}</li>)}
      </ul>
      <div className="warning-block">
        {analysis.warnings.map((warning) => <span key={warning}>{warning}</span>)}
      </div>
    </Panel>
  );
}

export function FeedPanel({ title, items, type }) {
  return (
    <Panel title={title} meta={`${items.length}`}>
      <div className="feed-list condensed">
        {items.slice(0, 4).map((item, index) => (
          <article key={`${type}-${index}`} className="feed-item">
            <div>
              <span className={`mini-signal ${type === 'filing' ? 'info' : item.label}`}>
                {type === 'filing' ? item.form : item.label}
              </span>
              {type === 'news' && <span className="tag">{item.catalyst}</span>}
            </div>
            {item.url ? (
              <a href={item.url} target="_blank" rel="noreferrer">{item.title || `${item.form} filing`}</a>
            ) : (
              <strong>{item.title || `${item.form} filing`}</strong>
            )}
            <small>{item.source || item.date} {item.published ? `· ${formatDate(item.published)}` : ''}</small>
          </article>
        ))}
        {items.length === 0 && <p className="empty">No items available.</p>}
      </div>
    </Panel>
  );
}

export function WorkflowRail({ run, fallbackWorkers, timeline, snapshots }) {
  const workers = run?.workers?.length ? run.workers : fallbackWorkers || [];
  const running = run?.status === 'running';
  return (
    <aside className="workflow-rail panel">
      <div className="panel-header">
        <h2>Workflow</h2>
        <span className={`mini-signal ${signalClass(run?.status || 'idle')}`}>{run?.status || 'idle'}</span>
      </div>
      {running && <div className="workflow-loader"><Spinner /> Running threaded analysis</div>}
      <div className="worker-list">
        {workers.map((worker, index) => (
          <div key={`${worker.worker}-${worker.id || index}`} className="worker-row">
            <span className={`dot ${worker.status}`}></span>
            <div>
              <strong>{worker.worker}</strong>
              <small>{worker.message}</small>
            </div>
            <span>{worker.status}</span>
          </div>
        ))}
      </div>
      <div className="timeline-compact">
        <div className="panel-header">
          <h2>Timeline</h2>
          <span>{timeline.length}</span>
        </div>
        {timeline.slice(0, 5).map((entry) => (
          <article key={entry.id} className="timeline-entry compact-entry">
            <div>
              <strong>{entry.symbol}</strong>
              <span className={`mini-signal ${signalClass(entry.signal)}`}>{entry.signal}</span>
            </div>
            <small>{formatDate(entry.createdAt)} · score {entry.score}</small>
          </article>
        ))}
        {timeline.length === 0 && <p className="empty">No timeline entries yet.</p>}
      </div>
      <div className="snapshot-row compact-snapshots">
        <strong>Snapshots</strong>
        {snapshots.length ? snapshots.slice(0, 2).map((snapshot) => (
          <span key={snapshot.id}>{snapshot.fileName}</span>
        )) : <span>No snapshots.</span>}
      </div>
    </aside>
  );
}

function Panel({ title, meta, children }) {
  return (
    <section className="panel content-panel">
      <div className="panel-header">
        <h2>{title}</h2>
        {meta && <span>{meta}</span>}
      </div>
      {children}
    </section>
  );
}

export function Spinner() {
  return <span className="spinner" aria-label="Processing"></span>;
}
