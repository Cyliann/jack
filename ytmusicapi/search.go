package ytmusicapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const API_KEY = "AIzaSyDyT5W0Jh49F30Pqqtyfdf7pDLFKLJoAnw"
const AlbumSearchParams = "EgWKAQIYAWoQEAkQAxAEEAoQBRAQEBUQEQ=="
const SongSearchParams = "EgWKAQIIAWoQEAMQBBAEEAoQBRAQEBUQEQ=="

type AlbumResult struct {
	Artist string
	Album  string
	Id     string
}

func ParseAlbumResults(jsonBytes []byte) ([]AlbumResult, error) {
	var root map[string]any
	if err := json.Unmarshal(jsonBytes, &root); err != nil {
		return nil, err
	}

	var results []AlbumResult

	// Navigate into:
	// contents → tabbedSearchResultsRenderer → tabs → tabRenderer → content → sectionListRenderer → contents
	contents := asMap(root["contents"])
	if contents == nil {
		return results, nil
	}

	tabs := asSlice(
		contents["tabbedSearchResultsRenderer"].(map[string]any)["tabs"],
	)

	for _, tab := range tabs {
		tabRenderer := asMap(tab)["tabRenderer"]
		if tabRenderer == nil {
			continue
		}

		sectionList := asMap(
			asMap(tabRenderer)["content"],
		)["sectionListRenderer"]

		sections := asSlice(asMap(sectionList)["contents"])

		for _, section := range sections {
			shelf := asMap(section)["musicShelfRenderer"]
			if shelf == nil {
				continue
			}

			items := asSlice(asMap(shelf)["contents"])
			for _, item := range items {
				renderer := asMap(item)["musicResponsiveListItemRenderer"]
				if renderer == nil {
					continue
				}
				flexCols := asSlice(renderer.(map[string]any)["flexColumns"])
				if len(flexCols) == 0 {
					continue
				}

				item_type := getTextRuns(
					asMap(flexCols[1])["musicResponsiveListItemFlexColumnRenderer"].(map[string]any)["text"], 0,
				)
				// Skip non albums
				if item_type != "Album" && item_type != "EP" {
					continue
				}

				// Album name
				album := getTextRuns(
					asMap(flexCols[0])["musicResponsiveListItemFlexColumnRenderer"].(map[string]any)["text"], 0,
				)

				// Artist
				artist := getTextRuns(
					asMap(flexCols[1])["musicResponsiveListItemFlexColumnRenderer"].(map[string]any)["text"], 2,
				)

				url := extractPlaylistID(renderer.(map[string]any))

				if album != "" && artist != "" {
					results = append(results, AlbumResult{
						Artist: artist,
						Album:  album,
						Id:     url,
					})
				}
			}
		}
	}

	return results, nil
}

func Search(query string) []AlbumResult {
	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "WEB_REMIX",
				"clientVersion": "1.20231220.01.00",
			},
		},
		"query":  query,
		"params": AlbumSearchParams,
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("https://music.youtube.com/youtubei/v1/search?prettyPrint=false"),
		bytes.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// err = os.WriteFile("response.json", data, 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	albums, err := ParseAlbumResults(data)
	if err != nil {
		log.Fatal(err)
	}

	return albums
}
