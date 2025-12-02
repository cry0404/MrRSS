package article

import (
	"encoding/json"
	"net/http"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

// HandleProgress returns the current fetch progress.
func HandleProgress(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	progress := h.Fetcher.GetProgress()
	json.NewEncoder(w).Encode(progress)
}

// HandleFilteredArticles returns articles filtered by advanced conditions from the database.
func HandleFilteredArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 50
	}

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	// Get all articles from database
	// Note: Using a high limit to fetch all articles for filtering
	// For very large datasets, consider implementing database-level filtering
	articles, err := h.DB.GetArticles("", 0, "", showHidden, 50000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get feeds for category lookup
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a map of feed ID to category
	feedCategories := make(map[int64]string)
	for _, feed := range feeds {
		feedCategories[feed.ID] = feed.Category
	}

	// Apply filter conditions
	if len(req.Conditions) > 0 {
		var filteredArticles []models.Article
		for _, article := range articles {
			if evaluateArticleConditions(article, req.Conditions, feedCategories) {
				filteredArticles = append(filteredArticles, article)
			}
		}
		articles = filteredArticles
	}

	// Apply pagination
	total := len(articles)
	offset := (page - 1) * limit
	end := offset + limit

	// Handle edge cases for pagination
	var paginatedArticles []models.Article
	if offset >= total {
		// No more articles to show
		paginatedArticles = []models.Article{}
	} else {
		if end > total {
			end = total
		}
		paginatedArticles = articles[offset:end]
	}

	hasMore := end < total

	response := FilterResponse{
		Articles: paginatedArticles,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
	}

	json.NewEncoder(w).Encode(response)
}
