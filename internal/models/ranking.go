package models

import "time"

// RankType distinguishes UFC's two ranking systems: media-panel voted vs the
// AI-generated "META" rankings.
type RankType string

const (
	RankTypeMedia RankType = "media"
	RankTypeMeta  RankType = "meta"
)

// Ranking is a single division/position entry in either the media or meta
// UFC rankings, mirroring the tables at ufc.com/rankings.
type Ranking struct {
	ID                int       `json:"id"`
	RankType          RankType  `json:"rank_type"`
	Division          string    `json:"division"`
	Position          int       `json:"position"`
	FighterName       string    `json:"fighter_name"`
	FighterSlug       string    `json:"fighter_slug,omitempty"`
	IsChampion        bool      `json:"is_champion"`
	Movement          int       `json:"movement"`
	ChampionImageURL  string    `json:"champion_image_url,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// DivisionRankings groups a division's champion and ordered contender list
// for template rendering.
type DivisionRankings struct {
	Division   string
	Champion   *Ranking
	Contenders []Ranking
}
