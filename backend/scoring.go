package main

import (
	"math"
	"strings"
	"time"
)

func scoreAnalysis(quote Quote, news []NewsItem, filings []Filing) (int, string, string, []string, []string) {
	score := 50
	var reasons []string
	var warnings []string

	if quote.ChangePercent > 1 {
		score += 12
		reasons = append(reasons, "Price momentum is positive in the latest quote.")
	} else if quote.ChangePercent < -1 {
		score -= 12
		reasons = append(reasons, "Price momentum is negative in the latest quote.")
	} else {
		reasons = append(reasons, "Price movement is relatively muted.")
	}

	positive, negative := 0, 0
	for _, item := range news {
		switch item.Label {
		case "positive":
			positive++
		case "negative":
			negative++
		}
	}
	score += positive*5 - negative*6
	if positive > negative {
		reasons = append(reasons, "Recent news mix leans positive.")
	} else if negative > positive {
		reasons = append(reasons, "Recent news mix contains more negative items.")
	} else {
		reasons = append(reasons, "Recent news does not show a clear directional bias.")
	}

	for _, filing := range filings {
		if filing.Form == "4" {
			reasons = append(reasons, "Recent Form 4 activity is present and worth reviewing.")
			break
		}
	}
	if len(filings) == 0 {
		warnings = append(warnings, "SEC filing data was unavailable or unsupported for this symbol.")
	}
	if quote.Source == "Mock provider" {
		warnings = append(warnings, "Quote data is mock data because no market-data API key is configured.")
	}
	if len(news) > 0 && news[0].Source == "Mock provider" {
		warnings = append(warnings, "News data is sample data because no news API key is configured.")
	}
	warnings = append(warnings, "Options and institutional analysis are deferred from v1.")

	score = int(math.Max(0, math.Min(100, float64(score))))
	signal := "Neutral"
	if score >= 61 {
		signal = "Bullish"
	} else if score <= 39 {
		signal = "Bearish"
	}
	confidence := "Medium"
	if len(warnings) > 2 {
		confidence = "Low"
	} else if quote.Source != "Mock provider" && len(news) >= 4 && len(filings) > 0 {
		confidence = "High"
	}
	return score, signal, confidence, trimStrings(reasons, 5), warnings
}

func makeChart(quote Quote) []ChartPoint {
	points := []ChartPoint{}
	now := time.Now()
	for i := 29; i >= 0; i-- {
		wave := math.Sin(float64(i)/3) * quote.Price * 0.006
		drift := float64(29-i) * quote.Change / 35
		points = append(points, ChartPoint{
			Time:  now.Add(time.Duration(-i*10) * time.Minute).Format("15:04"),
			Price: round(quote.Price - drift + wave),
		})
	}
	return points
}

func normalizeLabel(label string) string {
	label = strings.ToLower(label)
	if strings.Contains(label, "bullish") || strings.Contains(label, "positive") {
		return "positive"
	}
	if strings.Contains(label, "bearish") || strings.Contains(label, "negative") {
		return "negative"
	}
	return "neutral"
}

func catalystFor(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "earnings"):
		return "earnings"
	case strings.Contains(lower, "product"):
		return "product"
	case strings.Contains(lower, "analyst"):
		return "analyst"
	case strings.Contains(lower, "lawsuit") || strings.Contains(lower, "sec") || strings.Contains(lower, "regulatory"):
		return "legal/regulatory"
	case strings.Contains(lower, "macro") || strings.Contains(lower, "rates"):
		return "macro"
	case strings.Contains(lower, "partner"):
		return "partnership"
	case strings.Contains(lower, "offering") || strings.Contains(lower, "dilution"):
		return "financing/dilution"
	default:
		return "market"
	}
}
