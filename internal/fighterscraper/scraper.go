// Package fighterscraper scrapes the full UFC athlete roster from
// https://www.ufc.com/athletes/all and upserts active fighters into the
// database. It's used by cmd/fighter-scraper for one-off/manual runs.
package fighterscraper

import (
	"context"
	"fmt"
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
	listingURL = "https://www.ufc.com/athletes/all"
	profileURL = "https://www.ufc.com/athlete/"
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"
)

var recordRe = regexp.MustCompile(`(\d+)-(\d+)-(\d+)`)

// Options configures a Run.
type Options struct {
	// Delay is the pause between requests to ufc.com.
	Delay time.Duration
	// MaxPages caps how many listing pages are fetched (0 = unlimited).
	MaxPages int
	// DryRun logs would-be upserts instead of writing to the database, so
	// repo may be nil when DryRun is set.
	DryRun bool
}

// Stats summarizes a completed Run.
type Stats struct {
	Listed   int
	Active   int
	Upserted int
	Failed   int
}

type listingEntry struct {
	Name        string
	Nickname    string
	WeightClass string
	Wins        int
	Losses      int
	Draws       int
	Slug        string
}

type profile struct {
	IsActive     bool
	Nationality  string
	HeightFeet   *int
	HeightInches *int
	WeightLbs    *int
	ReachCm      *int
	LegReachCm   *int
}

// Run walks every page of ufc.com/athletes/all, fetches each fighter's
// profile page to determine active/retired status, and upserts active
// fighters into the database. Retired fighters are skipped rather than
// written with is_active=false: ufc.com only exposes status via a
// per-fighter profile fetch, and roughly 2500 of the ~3150 listed athletes
// are retired/cut, so there's no reason to store them.
func Run(ctx context.Context, repo *repository.Repository, opts Options) (Stats, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	var stats Stats

	for page := 0; opts.MaxPages == 0 || page < opts.MaxPages; page++ {
		if page > 0 {
			time.Sleep(opts.Delay)
		}

		doc, err := fetchListingPage(ctx, client, page)
		if err != nil {
			return stats, fmt.Errorf("fetch listing page %d: %w", page, err)
		}

		entries := parseListingEntries(doc)
		if len(entries) == 0 {
			break
		}
		stats.Listed += len(entries)

		for _, entry := range entries {
			time.Sleep(opts.Delay)

			p, err := fetchProfile(ctx, client, entry.Slug)
			if err != nil {
				log.Printf("fighterscraper: fail  %-25s %v\n", entry.Name, err)
				stats.Failed++
				continue
			}

			if !p.IsActive {
				continue
			}
			stats.Active++

			fighter := toFighter(entry, p)

			if opts.DryRun {
				log.Printf("fighterscraper: [dry-run] active  %-25s %-20s %d-%d-%d\n", fighter.Name, fighter.WeightClass, fighter.Wins, fighter.Losses, fighter.Draws)
				stats.Upserted++
				continue
			}

			if err := repo.Fighters.UpsertFighter(ctx, &fighter); err != nil {
				log.Printf("fighterscraper: fail  %-25s upsert: %v\n", entry.Name, err)
				stats.Failed++
				continue
			}
			stats.Upserted++
		}

		log.Printf("fighterscraper: page %d done: listed=%d active=%d upserted=%d failed=%d\n", page, stats.Listed, stats.Active, stats.Upserted, stats.Failed)
	}

	log.Printf("fighterscraper: scrape complete: listed=%d active=%d upserted=%d failed=%d\n", stats.Listed, stats.Active, stats.Upserted, stats.Failed)
	return stats, nil
}

func fetchListingPage(ctx context.Context, client *http.Client, page int) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?page=%d", listingURL, page), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

// parseListingEntries parses every fighter card on a ufc.com/athletes/all
// listing page. Each card's markup contains two copies of the name/nickname
// (front face and flip-back caption), so lookups use .First().
func parseListingEntries(doc *goquery.Document) []listingEntry {
	var entries []listingEntry

	doc.Find(".c-listing-athlete-flipcard").Each(func(_ int, card *goquery.Selection) {
		name := strings.TrimSpace(card.Find(".c-listing-athlete__name").First().Text())
		if name == "" {
			return
		}

		href, _ := card.Find(".c-listing-athlete-flipcard__action a").First().Attr("href")
		slug := strings.TrimPrefix(strings.TrimSpace(href), "/athlete/")
		if slug == "" {
			return
		}

		nickname := strings.Trim(strings.TrimSpace(card.Find(".c-listing-athlete__nickname").First().Text()), `"`)
		weightClass := strings.TrimSpace(card.Find(".c-listing-athlete__title .field__item").First().Text())

		wins, losses, draws := 0, 0, 0
		if m := recordRe.FindStringSubmatch(card.Find(".c-listing-athlete__record").First().Text()); m != nil {
			wins, _ = strconv.Atoi(m[1])
			losses, _ = strconv.Atoi(m[2])
			draws, _ = strconv.Atoi(m[3])
		}

		entries = append(entries, listingEntry{
			Name:        name,
			Nickname:    nickname,
			WeightClass: weightClass,
			Wins:        wins,
			Losses:      losses,
			Draws:       draws,
			Slug:        slug,
		})
	})

	return entries
}

// fetchProfile fetches an athlete's profile page and reads its .c-bio__field
// label/text pairs (Status, Place of Birth, Weight, Height — the fields
// ufc.com's current profile redesign actually exposes; the old tale-of-the-
// tape with reach/stance/DOB is gone). ufc.com redirects unknown slugs to
// its search page, which is treated as an error.
func fetchProfile(ctx context.Context, client *http.Client, slug string) (profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL+slug, nil)
	if err != nil {
		return profile{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return profile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return profile{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	if strings.HasPrefix(resp.Request.URL.Path, "/search") {
		return profile{}, fmt.Errorf("no athlete page for slug %q", slug)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return profile{}, err
	}

	fields := map[string]string{}
	doc.Find(".c-bio__field").Each(func(_ int, field *goquery.Selection) {
		label := strings.TrimSpace(field.Find(".c-bio__label").First().Text())
		text := strings.TrimSpace(field.Find(".c-bio__text").First().Text())
		if label != "" {
			fields[label] = text
		}
	})

	status, ok := fields["Status"]
	if !ok {
		return profile{}, fmt.Errorf("no bio status found for slug %q (page structure may have changed)", slug)
	}

	var p profile
	p.IsActive = strings.EqualFold(status, "Active")
	p.Nationality = fields["Place of Birth"]

	if lbs, err := strconv.ParseFloat(fields["Weight"], 64); err == nil && lbs > 0 {
		v := int(lbs + 0.5)
		p.WeightLbs = &v
	}

	if inches, err := strconv.ParseFloat(fields["Height"], 64); err == nil && inches > 0 {
		feet := int(inches) / 12
		rem := int(inches+0.5) % 12
		p.HeightFeet = &feet
		p.HeightInches = &rem
	}

	if inches, err := strconv.ParseFloat(fields["Reach"], 64); err == nil && inches > 0 {
		cm := int(inches*2.54 + 0.5)
		p.ReachCm = &cm
	}

	if inches, err := strconv.ParseFloat(fields["Leg reach"], 64); err == nil && inches > 0 {
		cm := int(inches*2.54 + 0.5)
		p.LegReachCm = &cm
	}

	return p, nil
}

func toFighter(entry listingEntry, p profile) models.Fighter {
	f := models.Fighter{
		Name:         entry.Name,
		WeightClass:  entry.WeightClass,
		Wins:         entry.Wins,
		Losses:       entry.Losses,
		Draws:        entry.Draws,
		IsActive:     true,
		HeightFeet:   p.HeightFeet,
		HeightInches: p.HeightInches,
		WeightLbs:    p.WeightLbs,
		ReachCm:      p.ReachCm,
		LegReachCm:   p.LegReachCm,
	}
	if entry.Nickname != "" {
		f.Nickname = &entry.Nickname
	}
	if p.Nationality != "" {
		f.Nationality = &p.Nationality
	}
	return f
}
