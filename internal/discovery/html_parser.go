package discovery

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// findFriendLinks searches for friend link pages and extracts links
func (s *Service) findFriendLinks(ctx context.Context, homepage string) ([]string, error) {
	return s.findFriendLinksWithProgress(ctx, homepage, nil)
}

// findFriendLinksWithProgress searches for friend link pages with progress updates
func (s *Service) findFriendLinksWithProgress(ctx context.Context, homepage string, progressCb ProgressCallback) ([]string, error) {
	// Try to find friend link page
	friendPageURL, err := s.findFriendLinkPage(ctx, homepage)
	if err != nil {
		log.Printf("Could not find friend link page, trying homepage: %v", err)
		friendPageURL = homepage
	}

	if progressCb != nil {
		progressCb(Progress{
			Stage:   "fetching_friend_page",
			Message: "Fetching friend links page",
			Detail:  friendPageURL,
		})
	}

	// Fetch and parse the friend link page
	doc, err := s.fetchHTML(ctx, friendPageURL)
	if err != nil {
		return nil, err
	}

	// Extract all external links
	links := s.extractExternalLinks(doc, friendPageURL)

	if progressCb != nil {
		progressCb(Progress{
			Stage:   "found_links",
			Message: fmt.Sprintf("Found %d potential blog links", len(links)),
			Total:   len(links),
		})
	}

	return links, nil
}

// findFriendLinkPage searches for a friend link page
func (s *Service) findFriendLinkPage(ctx context.Context, homepage string) (string, error) {
	doc, err := s.fetchHTML(ctx, homepage)
	if err != nil {
		return "", err
	}

	// Expanded patterns for friend link pages (multiple languages and variations)
	patterns := []string{
		// Chinese patterns
		"友链", "友情链接", "博客友链", "友情", "朋友们", "小伙伴", "友邻", "链接",
		// English patterns
		"blogroll", "friends", "links", "friend links", "blog links",
		"link", "buddy", "buddies", "partner", "partners", "bloggers",
		"recommended", "blog roll", "favorite blogs", "other blogs",
		// Common URL paths
		"about/links", "friends.html", "links.html", "blogroll.html",
		"friend", "flink", "link-exchange",
	}

	var foundURL string
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		if foundURL != "" {
			return
		}

		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		text := strings.ToLower(strings.TrimSpace(sel.Text()))
		hrefLower := strings.ToLower(href)

		// Check if link text or href contains friend link patterns
		for _, pattern := range patterns {
			if strings.Contains(text, pattern) || strings.Contains(hrefLower, pattern) {
				// Resolve relative URLs
				if absURL := s.resolveURL(homepage, href); absURL != "" {
					foundURL = absURL
					return
				}
			}
		}
	})

	if foundURL != "" {
		return foundURL, nil
	}

	return "", errFriendLinkPageNotFound
}

// extractExternalLinks extracts all external links from a page
func (s *Service) extractExternalLinks(doc *goquery.Document, baseURL string) []string {
	seen := make(map[string]bool)
	var links []string

	baseU, err := url.Parse(baseURL)
	if err != nil {
		return links
	}

	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		absURL := s.resolveURL(baseURL, href)
		if absURL == "" {
			return
		}

		u, err := url.Parse(absURL)
		if err != nil {
			return
		}

		// Only include external links (different domain)
		if u.Host != baseU.Host && u.Host != "" {
			// Skip common non-blog domains
			if s.isValidBlogDomain(u.Host) && !seen[absURL] {
				seen[absURL] = true
				links = append(links, absURL)
			}
		}
	})

	return links
}

// isValidBlogDomain checks if a domain is likely a blog
func (s *Service) isValidBlogDomain(host string) bool {
	// Skip common non-blog domains
	skipDomains := []string{
		"facebook.com", "twitter.com", "instagram.com", "linkedin.com",
		"youtube.com", "github.com", "stackoverflow.com", "reddit.com",
		"weibo.com", "zhihu.com", "bilibili.com", "douban.com",
		"google.com", "baidu.com", "bing.com", "yahoo.com",
	}

	hostLower := strings.ToLower(host)
	for _, skip := range skipDomains {
		if strings.Contains(hostLower, skip) {
			return false
		}
	}

	return true
}

// resolveURL resolves a relative URL to an absolute URL
func (s *Service) resolveURL(base, href string) string {
	if href == "" {
		return ""
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return ""
	}

	return baseURL.ResolveReference(hrefURL).String()
}
