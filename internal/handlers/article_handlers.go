package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/models"
)

// HandleArticles returns articles with filtering and pagination.
func (h *Handler) HandleArticles(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	feedIDStr := r.URL.Query().Get("feed_id")
	category := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var feedID int64
	if feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	articles, err := h.DB.GetArticles(filter, feedID, category, showHidden, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(articles)
}

// HandleProgress returns the current fetch progress.
func (h *Handler) HandleProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.Fetcher.GetProgress()
	json.NewEncoder(w).Encode(progress)
}

// HandleMarkRead marks an article as read or unread.
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	readStr := r.URL.Query().Get("read")
	read := true
	if readStr == "false" || readStr == "0" {
		read = false
	}

	if err := h.DB.MarkArticleRead(id, read); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleToggleFavorite toggles the favorite status of an article.
func (h *Handler) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.DB.ToggleFavorite(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleGetUnreadCounts returns unread counts for all feeds.
func (h *Handler) HandleGetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	// Get total unread count
	totalCount, err := h.DB.GetTotalUnreadCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unread counts per feed
	feedCounts, err := h.DB.GetUnreadCountsForAllFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total":       totalCount,
		"feed_counts": feedCounts,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleMarkAllAsRead marks all articles as read.
func (h *Handler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	feedIDStr := r.URL.Query().Get("feed_id")

	var err error
	if feedIDStr != "" {
		// Mark all as read for a specific feed
		feedID, parseErr := strconv.ParseInt(feedIDStr, 10, 64)
		if parseErr != nil {
			http.Error(w, "Invalid feed_id parameter", http.StatusBadRequest)
			return
		}
		err = h.DB.MarkAllAsReadForFeed(feedID)
	} else {
		// Mark all as read globally
		err = h.DB.MarkAllAsRead()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleRefresh triggers a refresh of all feeds.
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	go h.Fetcher.FetchAll(context.Background())
	w.WriteHeader(http.StatusOK)
}

// HandleCleanupArticles triggers manual cleanup of articles.
func (h *Handler) HandleCleanupArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.DB.CleanupUnimportantArticles()
	if err != nil {
		log.Printf("Error cleaning up articles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Cleaned up %d articles", count)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": count,
	})
}

// HandleToggleHideArticle toggles the hidden status of an article.
func (h *Handler) HandleToggleHideArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.ToggleArticleHidden(id); err != nil {
		log.Printf("Error toggling article hidden status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleGetArticleContent fetches the article content from RSS feed dynamically.
func (h *Handler) HandleGetArticleContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Get all articles to find the one we need
	allArticles, err := h.DB.GetArticles("", 0, "", false, 1000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var article *models.Article
	for i := range allArticles {
		if allArticles[i].ID == articleID {
			article = &allArticles[i]
			break
		}
	}

	if article == nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Get the feed to fetch fresh content
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var feedURL string
	for i := range feeds {
		if feeds[i].ID == article.FeedID {
			feedURL = feeds[i].URL
			break
		}
	}

	if feedURL == "" {
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Parse the feed to get fresh content
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsedFeed, err := h.Fetcher.ParseFeed(ctx, feedURL)
	if err != nil {
		log.Printf("Error parsing feed for article content: %v", err)
		http.Error(w, "Failed to fetch article content", http.StatusInternalServerError)
		return
	}

	// Find the article in the feed by URL
	var content string
	for _, item := range parsedFeed.Items {
		if item.Link == article.URL {
			content = item.Content
			if content == "" {
				content = item.Description
			}
			break
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"content": content,
	})
}
