package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ── Config ──────────────────────────────────────────────────────────────────

type Config struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func configPath() string {
	if p := os.Getenv("OUTLINE_CONFIG"); p != "" {
		return p
	}
	// Check shared location first, then group location
	shared := "/workspace/extra/outline-cli/config.json"
	if _, err := os.Stat(shared); err == nil {
		return shared
	}
	return "/workspace/group/outline-cli/config.json"
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil, fmt.Errorf("cannot read config (%s): %w — run 'outline setup --token TOKEN' first", configPath(), err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("config has no token — run 'outline setup --token TOKEN'")
	}
	if cfg.URL == "" {
		cfg.URL = "https://outline.nocoders.ai"
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

// ── HTTP client ──────────────────────────────────────────────────────────────

func apiPost(cfg *Config, endpoint string, body map[string]interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(cfg.URL, "/") + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response JSON: %w", err)
	}
	return result, nil
}

// ── Pretty print helpers ─────────────────────────────────────────────────────

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func prettyJSON(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

// ── setup ────────────────────────────────────────────────────────────────────

func cmdSetup(args []string) {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	token := fs.String("token", "", "API token")
	url := fs.String("url", "https://outline.nocoders.ai", "Outline base URL")
	fs.Parse(args)

	if *token == "" {
		fmt.Fprintln(os.Stderr, "ERR: --token is required")
		os.Exit(1)
	}

	cfg := &Config{
		URL:   *url,
		Token: *token,
	}
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK: Config saved → %s\n", configPath())
}

// ── collections ──────────────────────────────────────────────────────────────

func cmdCollections(args []string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		cmdCollectionsList(args[1:])
	case "create":
		cmdCollectionsCreate(args[1:])
	case "delete":
		cmdCollectionsDelete(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "ERR: unknown collections action: %s\n", args[0])
		os.Exit(1)
	}
}

func cmdCollectionsList(args []string) {
	fs := flag.NewFlagSet("collections list", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	result, err := apiPost(cfg, "/api/collections.list", map[string]interface{}{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, ok := result["data"]
	if !ok {
		fmt.Fprintln(os.Stderr, "ERR: unexpected response (no 'data')")
		os.Exit(1)
	}

	if *jsonOut {
		prettyJSON(dataRaw)
		return
	}

	items, _ := dataRaw.([]interface{})
	fmt.Printf("# Collections (%d)\n", len(items))
	for _, item := range items {
		col, _ := item.(map[string]interface{})
		id, _ := col["id"].(string)
		name, _ := col["name"].(string)
		urlId, _ := col["urlId"].(string)
		fmt.Printf("[%s] %s /%s\n", shortID(id), name, urlId)
	}
}

func cmdCollectionsCreate(args []string) {
	fs := flag.NewFlagSet("collections create", flag.ExitOnError)
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: NAME required")
		os.Exit(1)
	}
	name := strings.Join(rest, " ")

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	result, err := apiPost(cfg, "/api/collections.create", map[string]interface{}{"name": name})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, _ := result["data"].(map[string]interface{})
	id, _ := dataRaw["id"].(string)
	title, _ := dataRaw["name"].(string)
	fmt.Printf("OK: Created [%s] %s\n", shortID(id), title)
}

func cmdCollectionsDelete(args []string) {
	fs := flag.NewFlagSet("collections delete", flag.ExitOnError)
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: ID required")
		os.Exit(1)
	}
	id := rest[0]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	_, err = apiPost(cfg, "/api/collections.delete", map[string]interface{}{"id": id})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK: Deleted")
}

// ── docs ─────────────────────────────────────────────────────────────────────

func cmdDocs(args []string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		cmdDocsList(args[1:])
	case "create":
		cmdDocsCreate(args[1:])
	case "update":
		cmdDocsUpdate(args[1:])
	case "show":
		cmdDocsShow(args[1:])
	case "delete":
		cmdDocsDelete(args[1:])
	case "children":
		cmdDocsChildren(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "ERR: unknown docs action: %s\n", args[0])
		os.Exit(1)
	}
}

func cmdDocsList(args []string) {
	fs := flag.NewFlagSet("docs list", flag.ExitOnError)
	collection := fs.String("collection", "", "Collection ID filter")
	parent := fs.String("parent", "", "Parent document ID filter")
	jsonOut := fs.Bool("json", false, "JSON output")
	limit := fs.Int("limit", 0, "Max results")
	fs.Parse(args)

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	body := map[string]interface{}{}
	if *collection != "" {
		body["collectionId"] = *collection
	}
	if *parent != "" {
		body["parentDocumentId"] = *parent
	}
	if *limit > 0 {
		body["limit"] = *limit
	}

	result, err := apiPost(cfg, "/api/documents.list", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, ok := result["data"]
	if !ok {
		fmt.Fprintln(os.Stderr, "ERR: unexpected response (no 'data')")
		os.Exit(1)
	}

	if *jsonOut {
		prettyJSON(dataRaw)
		return
	}

	items, _ := dataRaw.([]interface{})
	fmt.Printf("# Documents (%d)\n", len(items))
	for _, item := range items {
		doc, _ := item.(map[string]interface{})
		id, _ := doc["id"].(string)
		title, _ := doc["title"].(string)
		fmt.Printf("[%s] %s\n", shortID(id), title)
	}
}

func cmdDocsCreate(args []string) {
	fs := flag.NewFlagSet("docs create", flag.ExitOnError)
	title := fs.String("title", "", "Document title")
	collection := fs.String("collection", "", "Collection ID")
	parent := fs.String("parent", "", "Parent document ID")
	text := fs.String("text", "", "Document text")
	fs.Parse(args)

	if *title == "" {
		fmt.Fprintln(os.Stderr, "ERR: --title is required")
		os.Exit(1)
	}
	if *collection == "" {
		fmt.Fprintln(os.Stderr, "ERR: --collection is required")
		os.Exit(1)
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	body := map[string]interface{}{
		"title":        *title,
		"collectionId": *collection,
		"publish":      true,
	}
	if *parent != "" {
		body["parentDocumentId"] = *parent
	}
	if *text != "" {
		body["text"] = *text
	}

	result, err := apiPost(cfg, "/api/documents.create", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, _ := result["data"].(map[string]interface{})
	id, _ := dataRaw["id"].(string)
	docTitle, _ := dataRaw["title"].(string)
	fmt.Printf("OK: Created [%s] %s\n", shortID(id), docTitle)
}

func cmdDocsUpdate(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: ID required")
		os.Exit(1)
	}
	id := args[0]

	fs := flag.NewFlagSet("docs update", flag.ExitOnError)
	title := fs.String("title", "", "New title")
	text := fs.String("text", "", "New text")
	fs.Parse(args[1:])

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	body := map[string]interface{}{
		"id":      id,
		"publish": true,
	}
	if *title != "" {
		body["title"] = *title
	}
	if *text != "" {
		body["text"] = *text
	}

	result, err := apiPost(cfg, "/api/documents.update", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, _ := result["data"].(map[string]interface{})
	docID, _ := dataRaw["id"].(string)
	docTitle, _ := dataRaw["title"].(string)
	fmt.Printf("OK: Updated [%s] %s\n", shortID(docID), docTitle)
}

func cmdDocsShow(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: ID required")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("docs show", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: ID required")
		os.Exit(1)
	}
	id := rest[0]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	result, err := apiPost(cfg, "/api/documents.info", map[string]interface{}{"id": id})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, ok := result["data"]
	if !ok {
		fmt.Fprintln(os.Stderr, "ERR: unexpected response (no 'data')")
		os.Exit(1)
	}

	if *jsonOut {
		prettyJSON(dataRaw)
		return
	}

	doc, _ := dataRaw.(map[string]interface{})
	docID, _ := doc["id"].(string)
	title, _ := doc["title"].(string)
	collectionId, _ := doc["collectionId"].(string)
	text, _ := doc["text"].(string)

	fmt.Printf("[%s] %s\n", shortID(docID), title)
	fmt.Printf("Collection: %s\n", shortID(collectionId))
	if text != "" {
		fmt.Printf("\n%s\n", text)
	}
}

func cmdDocsDelete(args []string) {
	fs := flag.NewFlagSet("docs delete", flag.ExitOnError)
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: ID required")
		os.Exit(1)
	}
	id := rest[0]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	_, err = apiPost(cfg, "/api/documents.delete", map[string]interface{}{"id": id})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK: Deleted")
}

func cmdDocsChildren(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: PARENT_ID required")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("docs children", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: PARENT_ID required")
		os.Exit(1)
	}
	parentID := rest[0]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	body := map[string]interface{}{
		"parentDocumentId": parentID,
	}

	result, err := apiPost(cfg, "/api/documents.list", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, ok := result["data"]
	if !ok {
		fmt.Fprintln(os.Stderr, "ERR: unexpected response (no 'data')")
		os.Exit(1)
	}

	if *jsonOut {
		prettyJSON(dataRaw)
		return
	}

	items, _ := dataRaw.([]interface{})
	fmt.Printf("# Children (%d)\n", len(items))
	for _, item := range items {
		doc, _ := item.(map[string]interface{})
		id, _ := doc["id"].(string)
		title, _ := doc["title"].(string)
		fmt.Printf("[%s] %s\n", shortID(id), title)
	}
}

// ── search ────────────────────────────────────────────────────────────────────

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	limit := fs.Int("limit", 0, "Max results")
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "ERR: QUERY required")
		os.Exit(1)
	}
	query := strings.Join(rest, " ")

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	body := map[string]interface{}{"query": query}
	if *limit > 0 {
		body["limit"] = *limit
	}

	result, err := apiPost(cfg, "/api/documents.search", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}

	dataRaw, ok := result["data"]
	if !ok {
		fmt.Fprintln(os.Stderr, "ERR: unexpected response (no 'data')")
		os.Exit(1)
	}

	if *jsonOut {
		prettyJSON(dataRaw)
		return
	}

	items, _ := dataRaw.([]interface{})
	fmt.Printf("# Results (%d)\n", len(items))
	for _, item := range items {
		entry, _ := item.(map[string]interface{})
		doc, _ := entry["document"].(map[string]interface{})
		context, _ := entry["context"].(string)
		id, _ := doc["id"].(string)
		title, _ := doc["title"].(string)
		fmt.Printf("[%s] %s\n", shortID(id), title)
		if context != "" {
			// Print first line of context trimmed
			lines := strings.SplitN(strings.TrimSpace(context), "\n", 2)
			fmt.Printf("  %s\n", strings.TrimSpace(lines[0]))
		}
	}
}

// ── usage ────────────────────────────────────────────────────────────────────

func printUsage() {
	fmt.Print(`outline — Outline wiki CLI

Usage:
  outline setup --token TOKEN [--url URL]

  outline collections list [--json]
  outline collections create NAME
  outline collections delete ID

  outline docs list [--collection ID] [--parent ID] [--json] [--limit N]
  outline docs create --title TITLE --collection ID [--parent ID] [--text TEXT]
  outline docs update ID [--title TITLE] [--text TEXT]
  outline docs show ID [--json]
  outline docs delete ID
  outline docs children PARENT_ID [--json]

  outline search QUERY [--json] [--limit N]

Config: OUTLINE_CONFIG env var or /workspace/extra/outline-cli/config.json
`)
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup(os.Args[2:])
	case "collections":
		cmdCollections(os.Args[2:])
	case "docs":
		cmdDocs(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "ERR: unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
