package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dancouver1/ufc-card-creator/internal/config"
	"github.com/dancouver1/ufc-card-creator/internal/db"
	"github.com/dancouver1/ufc-card-creator/internal/domain/repository"
	"github.com/dancouver1/ufc-card-creator/internal/models"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36"

var ogImageRe = regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)

// bioFullBodyRe matches UFC's "athlete_bio_full_body" image style: a
// transparent-background, torso-up fighting-stance photo used on bio /
// tale-of-the-tape pages. og:image is often a random editorial close-up
// instead, so this is preferred when present.
var bioFullBodyRe = regexp.MustCompile(`https://ufc\.com/images/styles/athlete_bio_full_body/s3/[^"?]+`)

func main() {
	names := flag.String("name", "", "Comma-separated fighter names to fetch (bypasses the DB fighter list)")
	all := flag.Bool("all", false, "Fetch images for every fighter in the database")
	outDir := flag.String("out", "./static/images/fighters", "Directory to save downloaded images")
	delay := flag.Duration("delay", 750*time.Millisecond, "Delay between requests to ufc.com")
	force := flag.Bool("force", false, "Re-download even if a local image already exists")
	updateDB := flag.Bool("update-db", true, "Update fighter_image_url in the database after downloading")
	flag.Parse()

	if *names == "" && !*all {
		log.Fatal("Provide -name \"Fighter One,Fighter Two\" or -all to fetch every fighter in the DB")
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	var fighters []models.Fighter
	var repo *repository.Repository

	if *names != "" {
		for _, n := range strings.Split(*names, ",") {
			n = strings.TrimSpace(n)
			if n != "" {
				fighters = append(fighters, models.Fighter{Name: n})
			}
		}
	}

	if *all || *updateDB {
		cfg := config.Load()
		database, err := db.NewDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer database.Close()

		repo = repository.New(database.Pool)

		if *all {
			ctx := context.Background()
			dbFighters, err := repo.Fighters.GetAllFighters(ctx)
			if err != nil {
				log.Fatalf("Failed to load fighters: %v", err)
			}
			fighters = dbFighters
		}
	}

	client := &http.Client{Timeout: 15 * time.Second}
	ctx := context.Background()

	fetched, skipped, failed := 0, 0, 0

	for i, fighter := range fighters {
		if i > 0 {
			time.Sleep(*delay)
		}

		slug := slugify(fighter.Name)
		localBase := localFilename(fighter.Name)

		if !*force {
			if existing := findExisting(*outDir, localBase); existing != "" {
				fmt.Printf("skip  %-25s (already have %s)\n", fighter.Name, existing)
				skipped++
				continue
			}
		}

		imageURL, err := fetchAthleteImage(client, slug)
		if err != nil {
			log.Printf("fail  %-25s %v\n", fighter.Name, err)
			failed++
			continue
		}

		ext := filepath.Ext(strings.SplitN(imageURL, "?", 2)[0])
		if ext == "" {
			ext = ".jpg"
		}
		destName := localBase + ext
		destPath := filepath.Join(*outDir, destName)

		if err := downloadFile(client, imageURL, destPath); err != nil {
			log.Printf("fail  %-25s %v\n", fighter.Name, err)
			failed++
			continue
		}

		fmt.Printf("saved %-25s -> %s\n", fighter.Name, destPath)
		fetched++

		if *updateDB && repo != nil {
			imagePath := "/static/images/fighters/" + destName
			if _, err := repo.Fighters.UpdateFighterImage(ctx, fighter.Name, imagePath); err != nil {
				log.Printf("warn  %-25s failed to update DB: %v\n", fighter.Name, err)
			}
		}
	}

	fmt.Printf("\n--- Done --- fetched=%d skipped=%d failed=%d\n", fetched, skipped, failed)
}

// slugify converts a fighter name into the slug ufc.com uses, e.g.
// "Justin Gaethje" -> "justin-gaethje".
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.NewReplacer("'", "", ".", "").Replace(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// localFilename mirrors the naming convention already used for images in
// static/images/fighters (see models.Fighter.SetDefaultImageURL).
func localFilename(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// findExisting returns the existing filename (with extension) for a base
// name if one is already present in dir, or "" if none is found.
func findExisting(dir, base string) string {
	for _, ext := range []string{".webp", ".png", ".jpg", ".jpeg"} {
		if _, err := os.Stat(filepath.Join(dir, base+ext)); err == nil {
			return base + ext
		}
	}
	return ""
}

// fetchAthleteImage requests the athlete page for slug and extracts a
// fighting-stance photo: the "athlete_bio_full_body" image style if present
// (transparent background, torso-up, used on tale-of-the-tape pages),
// falling back to the og:image meta tag otherwise. Returns an error if the
// slug doesn't resolve to a real athlete page (ufc.com redirects unknown
// slugs to its search page).
func fetchAthleteImage(client *http.Client, slug string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.ufc.com/athlete/"+slug, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	if strings.HasPrefix(resp.Request.URL.Path, "/search") {
		return "", fmt.Errorf("no athlete page for slug %q", slug)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if m := bioFullBodyRe.Find(body); m != nil {
		return string(m), nil
	}

	m := ogImageRe.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("no usable image found on athlete page")
	}

	imgURL := string(m[1])
	if u, err := url.Parse(imgURL); err == nil && u.Host == "" {
		imgURL = "https://www.ufc.com" + imgURL
	}
	return imgURL, nil
}

func downloadFile(client *http.Client, rawURL, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d downloading %s", resp.StatusCode, rawURL)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
