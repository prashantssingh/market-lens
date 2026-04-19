package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (a *App) getQuote(symbol string) Quote {
	if a.alphaKey != "" {
		url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, a.alphaKey)
		res, err := http.Get(url)
		if err == nil && res.StatusCode == http.StatusOK {
			defer res.Body.Close()
			var payload map[string]map[string]string
			if json.NewDecoder(res.Body).Decode(&payload) == nil {
				if q := payload["Global Quote"]; q != nil && q["05. price"] != "" {
					price := parseFloat(q["05. price"])
					change := parseFloat(q["09. change"])
					return Quote{
						Symbol: symbol, Price: price, Change: change, ChangePercent: parsePercent(q["10. change percent"]),
						Source: "Alpha Vantage", Freshness: "fresh", ObservedAt: time.Now().Format(time.RFC3339),
					}
				}
			}
		}
	}
	return mockQuote(symbol)
}

func (a *App) getNews(symbol string) []NewsItem {
	if a.alphaKey != "" {
		url := fmt.Sprintf("https://www.alphavantage.co/query?function=NEWS_SENTIMENT&tickers=%s&limit=8&apikey=%s", symbol, a.alphaKey)
		res, err := http.Get(url)
		if err == nil && res.StatusCode == http.StatusOK {
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			var payload struct {
				Feed []struct {
					Title            string `json:"title"`
					URL              string `json:"url"`
					Source           string `json:"source"`
					TimePublished    string `json:"time_published"`
					OverallSentiment string `json:"overall_sentiment_label"`
				} `json:"feed"`
			}
			if json.Unmarshal(body, &payload) == nil && len(payload.Feed) > 0 {
				items := []NewsItem{}
				for _, item := range payload.Feed {
					items = append(items, NewsItem{
						Title: item.Title, Source: item.Source, URL: item.URL, Published: item.TimePublished,
						Label: normalizeLabel(item.OverallSentiment), Catalyst: catalystFor(item.Title),
					})
				}
				return items
			}
		}
	}
	return mockNews(symbol)
}

func (a *App) getFilings(symbol string) []Filing {
	ciks := map[string]string{
		"AAPL": "0000320193", "MSFT": "0000789019", "NVDA": "0001045810",
		"TSLA": "0001318605", "AMZN": "0001018724", "GOOGL": "0001652044",
		"META": "0001326801", "QQQ": "", "SPY": "",
	}
	cik := ciks[symbol]
	if cik == "" {
		return nil
	}
	url := fmt.Sprintf("https://data.sec.gov/submissions/CIK%s.json", cik)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Market Lens local research app contact@example.com")
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return mockFilings(symbol)
	}
	defer res.Body.Close()
	var payload struct {
		Filings struct {
			Recent struct {
				AccessionNumber []string `json:"accessionNumber"`
				Form            []string `json:"form"`
				FilingDate      []string `json:"filingDate"`
				ReportDate      []string `json:"reportDate"`
			} `json:"recent"`
		} `json:"filings"`
	}
	if json.NewDecoder(res.Body).Decode(&payload) != nil {
		return mockFilings(symbol)
	}
	important := map[string]bool{"8-K": true, "10-Q": true, "10-K": true, "4": true, "S-3": true, "424B": true}
	var filings []Filing
	for i := 0; i < len(payload.Filings.Recent.Form) && len(filings) < 8; i++ {
		form := payload.Filings.Recent.Form[i]
		if !important[form] && !strings.HasPrefix(form, "424B") {
			continue
		}
		acc := strings.ReplaceAll(payload.Filings.Recent.AccessionNumber[i], "-", "")
		report := ""
		if i < len(payload.Filings.Recent.ReportDate) {
			report = payload.Filings.Recent.ReportDate[i]
		}
		filings = append(filings, Filing{
			Form: form, Date: payload.Filings.Recent.FilingDate[i], Report: report, Highlight: true,
			URL: fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%s/%s/", strings.TrimLeft(cik, "0"), acc),
		})
	}
	return filings
}

func mockQuote(symbol string) Quote {
	seed := 0
	for _, r := range symbol {
		seed += int(r)
	}
	price := 70 + float64(seed%420) + float64(seed%17)/10
	changePct := float64((seed%900)-420) / 100
	change := price * changePct / 100
	return Quote{
		Symbol: symbol, Price: round(price), Change: round(change), ChangePercent: round(changePct),
		Source: "Mock provider", Freshness: "sample", ObservedAt: time.Now().Format(time.RFC3339),
	}
}

func mockNews(symbol string) []NewsItem {
	return []NewsItem{
		{Title: symbol + " shows renewed investor attention after latest market session", Source: "Mock provider", Published: time.Now().Add(-2 * time.Hour).Format(time.RFC3339), Label: "positive", Catalyst: "market"},
		{Title: "Analysts watch " + symbol + " for near-term catalyst confirmation", Source: "Mock provider", Published: time.Now().Add(-5 * time.Hour).Format(time.RFC3339), Label: "neutral", Catalyst: "analyst"},
		{Title: symbol + " risk profile remains tied to broader macro conditions", Source: "Mock provider", Published: time.Now().Add(-9 * time.Hour).Format(time.RFC3339), Label: "neutral", Catalyst: "macro"},
	}
}

func mockFilings(symbol string) []Filing {
	if symbol == "QQQ" || symbol == "SPY" {
		return nil
	}
	return []Filing{{Form: "8-K", Date: time.Now().AddDate(0, 0, -6).Format("2006-01-02"), Report: "", URL: "https://www.sec.gov/edgar/search/", Highlight: true}}
}

func companyName(symbol string) string {
	names := map[string]string{
		"AAPL": "Apple Inc.", "MSFT": "Microsoft Corporation", "NVDA": "NVIDIA Corporation",
		"TSLA": "Tesla, Inc.", "AMZN": "Amazon.com, Inc.", "GOOGL": "Alphabet Inc.",
		"META": "Meta Platforms, Inc.", "SPY": "SPDR S&P 500 ETF Trust", "QQQ": "Invesco QQQ Trust",
	}
	if name := names[symbol]; name != "" {
		return name
	}
	return symbol
}
