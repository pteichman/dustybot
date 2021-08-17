package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("DISCORD_TOKEN environment variable not set")
		os.Exit(1)
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

var (
	matchTiktok   = regexp.MustCompile(`https://vm.tiktok.com/[^/]+/`)
	matchTiktokID = regexp.MustCompile(`https://m.tiktok.com/v/(.*)\.html`)
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if url := matchTiktok.FindString(m.Content); url != "" {
		content := ""

		embed, err := getEmbed(url)
		if err != nil {
			content = fmt.Sprintf("error: %s", err)
		}

		msg := &discordgo.MessageSend{
			Content: content,
			Embed:   embed,
			Reference: &discordgo.MessageReference{
				MessageID: m.ID,
				ChannelID: m.ChannelID,
				GuildID:   m.GuildID,
			},
		}

		s.ChannelMessageSendComplex(m.ChannelID, msg)
	}
}

func getEmbed(url string) (*discordgo.MessageEmbed, error) {
	real, err := getEmbedURL(url)
	if err != nil {
		return nil, fmt.Errorf("getRealUrl: %w", err)
	}

	req, err := http.NewRequest("GET", real, nil)
	if err != nil {
		return nil, fmt.Errorf("bad url: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	var tok embedResp
	err = json.NewDecoder(resp.Body).Decode(&tok)
	if err != nil {
		return nil, fmt.Errorf("bad json: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		URL:   url,
		Title: tok.Title,
		Author: &discordgo.MessageEmbedAuthor{
			URL:  tok.AuthorURL,
			Name: tok.AuthorName,
		},
		Image: &discordgo.MessageEmbedImage{
			URL:    tok.ThumbnailURL,
			Width:  tok.ThumbnailWidth,
			Height: tok.ThumbnailHeight,
		},
	}

	return embed, nil
}

func getEmbedURL(url string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("bad url: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}

	if resp.StatusCode != 301 {
		return "", fmt.Errorf("unexpected response: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	m := matchTiktokID.FindStringSubmatch(loc)

	url = "https://www.tiktok.com/oembed?url=" + m[0]
	return url, nil
}

type embedResp struct {
	Title string `json:"title"`

	AuthorName string `json:"author_name"`
	AuthorURL  string `json:"author_url"`

	ThumbnailURL    string `json:"thumbnail_url"`
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
}
