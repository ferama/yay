package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/go-shiori/go-readability"
	"github.com/gocolly/colly"
	"github.com/sashabaranov/go-openai"
)

type webSearchTool struct{}

func (t *webSearchTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "web_search",
			Description: "Performs a web search and returns the results.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query to perform.",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *webSearchTool) Handle(call openai.ToolCall) (string, error) {
	if call.Function.Name != "web_search" {
		return "", nil
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		return "", err
	}
	query, ok := args["query"].(string)
	if !ok {
		return "", nil
	}
	escapedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", escapedQuery)

	c := colly.NewCollector(
		colly.AllowedDomains("html.duckduckgo.com"),
		colly.UserAgent("Mozilla/5.0 (compatible; DuckScraper/1.0)"),
		colly.Async(true),
	)

	results := ""
	maxLinks := 5
	visited := 0

	c.OnHTML(".result", func(e *colly.HTMLElement) {
		if visited >= maxLinks {
			return
		}

		rawLink := e.ChildAttr("a.result__a", "href")
		title := e.ChildText("a.result__a")

		// DuckDuckGo wraps external links with /l/?uddg=...
		parsed, err := url.Parse(rawLink)
		finalURL := ""
		if err == nil && parsed.Host == "duckduckgo.com" && parsed.Path == "/l/" {
			uddg := parsed.Query().Get("uddg")
			unescaped, err := url.QueryUnescape(uddg)
			if err == nil {
				finalURL = unescaped
			} else {
				finalURL = uddg
			}
		} else if err == nil && parsed.Scheme != "" && parsed.Host != "" {
			finalURL = rawLink
		} else if rawLink != "" {
			finalURL = rawLink
		} else {
			finalURL = ""
		}

		if finalURL == "" {
			return
		}

		mainText := extractMainContent(finalURL)
		results += fmt.Sprintf("Title: %s\nURL: %s\nContent Preview: %s\n\n", title, finalURL, snippet(mainText, 500))
		visited++
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request error:", err)
	})

	if err := c.Visit(searchURL); err != nil {
		return "", err
	}
	c.Wait()

	return results, nil
}

// extractMainContent fetches the URL and returns the readable content
func extractMainContent(urlStr string) string {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "Failed to fetch: " + err.Error()
	}
	defer resp.Body.Close()

	u, _ := url.Parse(urlStr) // Ensure the URL is valid
	article, err := readability.FromReader(resp.Body, u)
	if err != nil {
		return "Failed to parse readable content: " + err.Error()
	}

	return article.Title + "\n\n" + article.TextContent
}

// snippet returns a short preview
func snippet(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
