// Package rankings scrapes https://www.ufc.com/rankings and stores both the
// media-panel and META (AI) rankings tables. It's used by cmd/rankings-scraper
// for one-off/manual runs and by cmd/server to keep rankings in sync on a
// schedule.
package rankings

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
	"github.com/dancouver1/ufc-card-creator/internal/models"
)

const (
	rankingsURL = "https://www.ufc.com/rankings"
	userAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"
)

var movementRe = regexp.MustCompile(`\d+`)

// Run fetches ufc.com/rankings and replaces the stored media and meta
// rankings snapshots.
func Run(ctx context.Context, repo *repository.Repository) error {
	doc, err := fetchDoc(ctx)
	if err != nil {
		return err
	}

	mediaRankings := parseView(doc, "view-display-id-block_1")
	metaRankings := parseView(doc, "view-display-id-meta_rankings")

	if len(mediaRankings) == 0 || len(metaRankings) == 0 {
		log.Printf("rankings: warn: parsed media=%d meta=%d entries; ufc.com markup may have changed", len(mediaRankings), len(metaRankings))
	}

	if len(mediaRankings) > 0 {
		if err := repo.Rankings.ReplaceAll(ctx, models.RankTypeMedia, mediaRankings); err != nil {
			return err
		}
	}
	if len(metaRankings) > 0 {
		if err := repo.Rankings.ReplaceAll(ctx, models.RankTypeMeta, metaRankings); err != nil {
			return err
		}
	}

	log.Printf("rankings: scrape complete: media=%d meta=%d entries", len(mediaRankings), len(metaRankings))
	return nil
}

// StartScheduler runs Run immediately and then every interval until ctx is
// canceled. Scrape errors are logged, not fatal, so a transient ufc.com
// outage doesn't affect the running server.
func StartScheduler(ctx context.Context, repo *repository.Repository, interval time.Duration) {
	go func() {
		if err := Run(ctx, repo); err != nil {
			log.Printf("rankings: initial scrape failed: %v", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := Run(ctx, repo); err != nil {
					log.Printf("rankings: scheduled scrape failed: %v", err)
				}
			}
		}
	}()
}

func fetchDoc(ctx context.Context) (*goquery.Document, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rankingsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

// parseView walks every division block inside the named Drupal view display
// ("view-display-id-block_1" for media rankings, "view-display-id-meta_rankings"
// for META/AI rankings) and turns it into a flat list of Ranking rows.
func parseView(doc *goquery.Document, viewDisplayClass string) []models.Ranking {
	var rankings []models.Ranking

	doc.Find("."+viewDisplayClass+" .view-grouping").Each(func(_ int, group *goquery.Selection) {
		division := normalizeDivision(strings.TrimSpace(group.Find(".view-grouping-header").First().Text()))
		if division == "" {
			return
		}

		caption := group.Find("caption .rankings--athlete--champion").First()
		if champHref, champName, ok := parseChampion(caption); ok {
			img, _ := caption.Find("img").First().Attr("src")
			rankings = append(rankings, models.Ranking{
				Division:         division,
				Position:         0,
				FighterName:      champName,
				FighterSlug:      slugFromHref(champHref),
				IsChampion:       true,
				ChampionImageURL: img,
			})
		}

		group.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() < 2 {
				return
			}

			posText := strings.TrimSpace(cells.Eq(0).Text())
			pos, err := strconv.Atoi(posText)
			if err != nil {
				return
			}

			nameCell := cells.Eq(1)
			name := strings.TrimSpace(nameCell.Text())
			href, _ := nameCell.Find("a").First().Attr("href")

			movement := 0
			if cells.Length() > 2 {
				changeText := strings.TrimSpace(cells.Eq(2).Text())
				if changeText != "" {
					if m := movementRe.FindString(changeText); m != "" {
						n, _ := strconv.Atoi(m)
						if strings.Contains(changeText, "decreas") {
							n = -n
						}
						movement = n
					}
				}
			}

			rankings = append(rankings, models.Ranking{
				Division:    division,
				Position:    pos,
				FighterName: name,
				FighterSlug: slugFromHref(href),
				IsChampion:  false,
				Movement:    movement,
			})
		})
	})

	return rankings
}

// parseChampion extracts the champion's profile href and name from a
// division's caption block. Returns ok=false for divisions with no belt
// holder (e.g. the Pound-for-Pound lists), identified by the absence of the
// "Champion" label.
func parseChampion(caption *goquery.Selection) (href, name string, ok bool) {
	if caption.Length() == 0 {
		return "", "", false
	}
	label := strings.TrimSpace(caption.Find("h6 .text").First().Text())
	if !strings.EqualFold(label, "Champion") {
		return "", "", false
	}
	link := caption.Find("h5 a").First()
	name = strings.TrimSpace(link.Text())
	href, _ = link.Attr("href")
	if name == "" {
		return "", "", false
	}
	return href, name, true
}

func slugFromHref(href string) string {
	parts := strings.Split(strings.Trim(href, "/"), "/")
	return parts[len(parts)-1]
}

// normalizeDivision strips the "<span>Top Rank</span>" suffix ufc.com adds
// to the Pound-for-Pound headers, e.g. "Men's Pound-for-Pound Top Rank" ->
// "Men's Pound-for-Pound".
func normalizeDivision(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "Top Rank")
	return strings.TrimSpace(s)
}
