package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SessionStats struct {
	SessionID     string `json:"session_id"`
	TotalTurns    int    `json:"total_turns"`
	InputTokens   int    `json:"input_tokens"`
	OutputTokens  int    `json:"output_tokens"`
	CacheTokens   int    `json:"cache_tokens"`
	LastModel     string `json:"last_model"`
}

func statsPath() string {
	return filepath.Join(cacheDir(), "stats.json")
}

func enrichUsage(event *Event) {
	if event.TranscriptPath == "" {
		return
	}
	usage := parseLastUsage(event.TranscriptPath)
	if usage != nil {
		event.Usage = usage
	}
}

func parseLastUsage(transcriptPath string) *TurnUsage {
	f, err := os.Open(transcriptPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lastAssistant map[string]interface{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry["type"] == "assistant" {
			lastAssistant = entry
		}
	}

	if lastAssistant == nil {
		return nil
	}

	msg, ok := lastAssistant["message"].(map[string]interface{})
	if !ok {
		return nil
	}

	usage, ok := msg["usage"].(map[string]interface{})
	if !ok {
		return nil
	}

	u := &TurnUsage{}
	if v, ok := usage["input_tokens"].(float64); ok {
		u.InputTokens = int(v)
	}
	if v, ok := usage["output_tokens"].(float64); ok {
		u.OutputTokens = int(v)
	}
	if v, ok := usage["cache_read_input_tokens"].(float64); ok {
		u.CacheTokens = int(v)
	}
	if v, ok := msg["model"].(string); ok {
		u.Model = v
	}

	return u
}

func saveStats(event *Event) {
	if event.Usage == nil || event.SessionID == "" {
		return
	}

	stats := loadStats()
	s, found := stats[event.SessionID]
	if !found {
		s = &SessionStats{SessionID: event.SessionID}
	}

	s.TotalTurns++
	s.InputTokens += event.Usage.InputTokens
	s.OutputTokens += event.Usage.OutputTokens
	s.CacheTokens += event.Usage.CacheTokens
	if event.Usage.Model != "" {
		s.LastModel = event.Usage.Model
	}

	stats[event.SessionID] = s
	writeStats(stats)
}

func loadStats() map[string]*SessionStats {
	stats := make(map[string]*SessionStats)
	data, err := os.ReadFile(statsPath())
	if err != nil {
		return stats
	}
	json.Unmarshal(data, &stats)
	return stats
}

func writeStats(stats map[string]*SessionStats) {
	os.MkdirAll(cacheDir(), 0755)
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(statsPath(), append(data, '\n'), 0644)
}

func cmdStats() {
	stats := loadStats()
	if len(stats) == 0 {
		fmt.Println("no stats yet — waiting for first CC response")
		return
	}

	fmt.Printf("%-38s %6s %10s %10s %10s %12s\n", "SESSION", "TURNS", "INPUT", "OUTPUT", "CACHE", "MODEL")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────────────")

	var totalInput, totalOutput, totalCache int
	for _, s := range stats {
		id := s.SessionID
		if len(id) > 36 {
			id = id[:36]
		}
		fmt.Printf("%-38s %6d %10d %10d %10d %12s\n",
			id, s.TotalTurns, s.InputTokens, s.OutputTokens, s.CacheTokens, s.LastModel)
		totalInput += s.InputTokens
		totalOutput += s.OutputTokens
		totalCache += s.CacheTokens
	}

	fmt.Println("───────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%-38s %6s %10d %10d %10d\n", "", "", totalInput, totalOutput, totalCache)

	pricingLabel := "Estimated cost (Claude Opus pricing)"
	inputPrice := 15.0 / 1e6
	outputPrice := 75.0 / 1e6
	cachePrice := 1.50 / 1e6

	for _, s := range stats {
		if s.LastModel == "deepseek-v4-pro" {
			pricingLabel = "Estimated cost (DeepSeek pricing)"
			inputPrice = 0.14 / 1e6
			outputPrice = 1.10 / 1e6
			cachePrice = 0.014 / 1e6
			break
		}
	}

	cost := float64(totalInput)*inputPrice + float64(totalOutput)*outputPrice + float64(totalCache)*cachePrice
	fmt.Printf("\n%s: ≈ $%.4f\n", pricingLabel, cost)
}
