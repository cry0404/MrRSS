package feed

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"time"

	"MrRSS/backend"
	"MrRSS/backend/history"
)

func GetFeedContent(db *sql.DB) []backend.FeedContentsInfo {
	feedList := GetFeedList(db)

	result := []backend.FeedContentsInfo{}

	rssLinks := []string{}

	for _, feed := range feedList {
		if feed.Category == "RSS/Atom" {
			rssLinks = append(rssLinks, feed.Link)
		} else {
			fmt.Println("Unknown category: ", feed.Category)
		}
	}

	rssContent := getRssContent(rssLinks)
	result = append(result, rssContent...)

	// sort
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time > result[j].Time
	})

	for i := range result {
		if history.CheckInHistory(db, result[i]) {
			result[i].Readed = history.GetHistoryReaded(db, result[i])
		}
	}

	return result
}

func filterImage(content string) *string {
	imgRegex := regexp.MustCompile(`img[^>]*src="([^"]*)`)

	var firstImageURL *string
	imgMatches := imgRegex.FindStringSubmatch(content)
	if len(imgMatches) > 1 {
		firstImageURL = &imgMatches[1]
	}

	return firstImageURL
}

func getTimeSince(t *time.Time) string {
	timeSince := time.Since(*t)
	timeStr := "now"

	if timeSince > 0 {
		if timeSince < time.Hour {
			minutes := int(timeSince.Minutes())
			if minutes == 1 {
				timeStr = fmt.Sprintf("%d minute ago", minutes)
			} else {
				timeStr = fmt.Sprintf("%d minutes ago", minutes)
			}
		} else if timeSince < 24*time.Hour {
			hours := int(timeSince.Hours())
			if hours == 1 {
				timeStr = fmt.Sprintf("%d hour ago", hours)
			} else {
				timeStr = fmt.Sprintf("%d hours ago", hours)
			}
		} else if timeSince < 365*24*time.Hour {
			days := int(timeSince.Hours() / 24)
			if days == 1 {
				timeStr = fmt.Sprintf("%d day ago", days)
			} else {
				timeStr = fmt.Sprintf("%d days ago", days)
			}
		} else {
			years := int(timeSince.Hours() / (365 * 24))
			if years == 1 {
				timeStr = fmt.Sprintf("%d year ago", years)
			} else {
				timeStr = fmt.Sprintf("%d years ago", years)
			}
		}
	}

	return timeStr
}
