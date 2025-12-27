package ytmusicapi

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func asSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func getTextRuns(v any, pos int) string {
	m := asMap(v)
	if m == nil {
		return ""
	}

	runs := asSlice(m["runs"])
	if len(runs) == 0 {
		return ""
	}

	run := asMap(runs[pos])
	if run == nil {
		return ""
	}

	text, _ := run["text"].(string)
	return text
}

// Navigate into:
// rendered -> musicItemThumbnailOverlayRenderer -> content -> musicPlayButtonRenderer -> playNavigationEndpoint ->
// -> watchPlaylistEndpoint -> playlistId
func extractPlaylistID(renderer map[string]any) string {
	overlay := asMap(renderer["overlay"])
	if overlay == nil {
		return ""
	}

	thumbOverlay := asMap(overlay["musicItemThumbnailOverlayRenderer"])
	if thumbOverlay == nil {
		return ""
	}

	content := asMap(thumbOverlay["content"])
	if content == nil {
		return ""
	}

	playButton := asMap(content["musicPlayButtonRenderer"])
	if playButton == nil {
		return ""
	}

	endpoint := asMap(playButton["playNavigationEndpoint"])
	if endpoint == nil {
		return ""
	}

	watchPlaylist := asMap(endpoint["watchPlaylistEndpoint"])
	if watchPlaylist == nil {
		return ""
	}

	playlistID, _ := watchPlaylist["playlistId"].(string)
	return playlistID
}
