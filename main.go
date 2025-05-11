package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

var (
	accessToken  string
	clientID     string
	clientSecret string
	redirectURI  string
)

func main() {
	_ = godotenv.Load()
	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	redirectURI = os.Getenv("REDIRECT_URI")

	r := gin.Default()
	r.GET("/", authRedirect)
	r.GET("/callback", handleCallback)
	r.GET("/artists", getArtists)
	r.GET("/play", play)
	r.GET("/pause", pause)
	r.GET("/top", topTracks)
	r.GET("/play-track/:track_id", playTrack)
	r.Run() // defaults to ":8080"
}

func authRedirect(c *gin.Context) {
	scope := "user-follow-read user-modify-playback-state user-read-playback-state user-read-currently-playing user-top-read"
	authURL := fmt.Sprintf("https://accounts.spotify.com/authorize?response_type=code&client_id=%s&scope=%s&redirect_uri=%s",
		url.QueryEscape(clientID), url.QueryEscape(scope), url.QueryEscape(redirectURI))
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func handleCallback(c *gin.Context) {
	code := c.Query("code")

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &tokenResp)
	accessToken = tokenResp.AccessToken

	c.String(200, "Visit /artists, /play, /pause or /top, /play-track/:track_id (ex: /play-track/spotify:artist:2J7iEpAkPqe41okCsm4Ja5)")
}

func basicAuth() string {
	return base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
}

func getArtists(c *gin.Context) {
	url := "https://api.spotify.com/v1/me/following?type=artist"
	method := "GET"
	spotifyAPI(c, method, url, nil)
}

func play(c *gin.Context) {
	url := "https://api.spotify.com/v1/me/player/play"
	method := "PUT"
	spotifyAPI(c, method, url, nil)
}

func pause(c *gin.Context) {
	url := "https://api.spotify.com/v1/me/player/pause"
	method := "PUT"
	spotifyAPI(c, method, url, nil)
}

func topTracks(c *gin.Context) {
	url := "https://api.spotify.com/v1/me/top/tracks?limit=10"
	method := "GET"
	spotifyAPI(c, method, url, nil)
}

func playTrack(c *gin.Context) {
	trackID := c.Param("track_id")
	url := "https://api.spotify.com/v1/me/player/play"
	method := "PUT"

	data := map[string]interface{}{
		"uris": []string{"spotify:track:" + trackID},
	}
	spotifyAPI(c, method, url, data)
}

func spotifyAPI(c *gin.Context, method, url string, data map[string]interface{}) {
	var jsonData []byte
	if data != nil {
		jsonData, _ = json.Marshal(data)
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Request creation failed"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Request failed"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}
