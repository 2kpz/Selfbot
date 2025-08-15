package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token               string   `json:"token"`
	OwnerID             string   `json:"owner_id"`
	LastFMAPIKEY        string   `json:"lastFMAPIKEY"`
	LastFMAPIUsername   string   `json:"lastFMAPIUsername"`
	Appid               string   `json:"appid"`
	Imgurl              string   `json:"imgurl"`
	Quote               string   `json:"quote"`
	Afktext             string   `json:"afktext"`
	Prefix              string   `json:"prefix"`
	AutoReactEmoji      string   `json:"autoReactEmoji"`
	AutoReactToggle     bool     `json:"autoReactToggle"`
	Client              string   `json:"client"`
	RotateGuilds        bool     `json:"rotate_guilds"`
	GuildIDs            []string `json:"guild_ids"`
	SpotifyClientID     string   `json:spotifyclientid`
	SpotifyClientSecret string   `json:spotifyclientsecret`
}

var (
	config    Config
	startTime = time.Now()
	eightBall = []string{
		"It is certain", "Without a doubt", "Yes definitely",
		"You may rely on it", "As I see it, yes", "Most likely",
		"Outlook good", "Yes", "Signs point to yes", "Reply hazy",
	}
)

func getMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

func fetchLastFmSong() (string, string, string, error) {
	url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json&limit=1",
		config.LastFMAPIUsername, config.LastFMAPIKEY)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", fmt.Errorf("error fetching song info: %v", err)
	}
	defer resp.Body.Close()

	var lastFmResponse struct {
		RecentTracks struct {
			Track []struct {
				Name   string `json:"name"`
				Artist struct {
					Name string `json:"#text"`
				} `json:"artist"`
				Image []struct {
					Size string `json:"size"`
					Text string `json:"#text"`
				} `json:"image"`
				Date struct {
					UTS string `json:"uts"`
				} `json:"date"`
				Attr struct {
					NowPlaying string `json:"nowplaying"`
				} `json:"@attr"`
			} `json:"track"`
		} `json:"recenttracks"`
	}
	err = json.NewDecoder(resp.Body).Decode(&lastFmResponse)
	if err != nil {
		return "", "", "", fmt.Errorf("error decoding last.fm response: %v", err)
	}

	if len(lastFmResponse.RecentTracks.Track) > 0 {
		track := lastFmResponse.RecentTracks.Track[0]
		if track.Attr.NowPlaying == "true" {
			return fmt.Sprintf("  %s - %s", track.Artist.Name, track.Name), track.Artist.Name + " - " + track.Name, "", nil
		}
	}

	return "No Song Playing", "default_image_key", config.Quote, nil
}

func main() {
	loadConfig()
	dg, err := discordgo.New(config.Token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}
	defer dg.Close()
	dg.AddHandler(messageHandler)
	dg.AddHandler(mentionHandler)
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: %v", err)
	}
	log.Println("Bot is running!")
	select {}
}

func loadConfig() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
}

var cmdCount int

func countCmd() {
	cmdCount++
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID != config.OwnerID || m.Author.Bot {
		return
	}

	content := strings.ToLower(strings.TrimSpace(m.Content))
	if !strings.HasPrefix(content, config.Prefix) {
		return
	}
	content = strings.TrimSpace(content[len(config.Prefix):])

	cmdCount = 0
	switch {
	case content == "help":
		cmdCount++
		showHelp(s, m)
	case content == "cmds":
		cmdCount++
		showHelp(s, m)
	case strings.HasPrefix(content, "afktext"):
		cmdCount++
		handleChangeAfkText(s, m)
	case strings.HasPrefix(content, "help "):
		cmdCount++
		showHelp(s, m)
	case strings.HasPrefix(content, "cmds "):
		cmdCount++
		showHelp(s, m)
	case strings.HasPrefix(content, "8ball"):
		cmdCount++
		handle8Ball(s, m)
	case strings.HasPrefix(content, "boobs"):
		cmdCount++
		handleBoobs(s, m)
	case content == "cat":
		cmdCount++
		handleCat(s, m)
	case content == "ci":
		cmdCount++
		handleCredits(s, m)
	case strings.HasPrefix(content, "dick"):
		cmdCount++
		handleDick(s, m)
	case content == "ar":
		cmdCount++
		handleAR(s, m)
	case strings.HasPrefix(content, "femboy"):
		cmdCount++
		handleFemboy(s, m)
	case strings.HasPrefix(content, "ap"):
		cmdCount++
		handleAutoCurse(s, m)
	case strings.HasPrefix(content, "rizz"):
		cmdCount++
		HandleRizz(s, m)
	case strings.HasPrefix(content, "tod"):
		cmdCount++
		handleTrueOrDare(s, m)
	case content == "np":
		cmdCount++
		handleNowPlaying(s, m)
	case strings.HasPrefix(content, "hentai"):
		cmdCount++
		handleHentai(s, m)
	case strings.HasPrefix(content, "blowjob"):
		cmdCount++
		handleBlowjob(s, m)
	case strings.HasPrefix(content, "dadjoke"):
		cmdCount++
		handleJoke(s, m)
	case strings.HasPrefix(content, "define"):
		cmdCount++
		handleDictionary(s, m)
	case strings.HasPrefix(content, "fact"):
		cmdCount++
		handleFunFact(s, m)
	case strings.HasPrefix(content, "setprefix"):
		cmdCount++
		handleChangePrefix(s, m)
	case strings.HasPrefix(content, "profile"):
		cmdCount++
		handleWxrnInfo(s, m)
	case strings.HasPrefix(content, "coinflip"):
		cmdCount++
		handleCoinFlip(s, m)
	case strings.HasPrefix(content, "fakehack"):
		cmdCount++
		handleFakeHack(s, m)
	case strings.HasPrefix(content, "meme"):
		cmdCount++
		handleMeme(s, m)
	case strings.HasPrefix(content, "setemoji"):
		cmdCount++
		handleSetAutoReact(s, m)
	case strings.HasPrefix(content, "togglereact"):
		cmdCount++
		handleToggleAutoReact(s, m)
	case strings.HasPrefix(content, "ch"):
		cmdCount++
		handleChangelogs(s, m)
	case strings.HasPrefix(content, "weather"):
		cmdCount++
		handleWeather(s, m)
	case strings.HasPrefix(content, "tspam"):
		cmdCount++
		handleMultiTokenSpam(s, m)
	case strings.HasPrefix(content, "tjoiner"):
		cmdCount++
		handleMultiTokenInvite(s, m)
	case strings.HasPrefix(content, "serverinfo"):
		cmdCount++
		handleServerInfo(s, m)
	case strings.HasPrefix(content, "dog"):
		cmdCount++
		handleDogPicture(s, m)
	case strings.HasPrefix(content, "joke"):
		cmdCount++
		handleRandomJoke(s, m)
	case strings.HasPrefix(content, "quote"):
		cmdCount++
		handleMotivationQuote(s, m)
	case strings.HasPrefix(content, "fact"):
		cmdCount++
		handleRandomFact(s, m)
	case strings.HasPrefix(content, "gr"):
		cmdCount++
		guildRotator(s, m)
	case strings.HasPrefix(content, "stopgr"):
		cmdCount++
		stopGuildRotator()
	case strings.HasPrefix(content, "hwd"):
		cmdCount++
		handlehwd(s, m)
	case strings.HasPrefix(content, "purge"):
		cmdCount++
		handlePurge(s, m)
	case strings.HasPrefix(content, "baka"):
		cmdCount++
		sendReply(s, m, "BAKA", "Baka baka baka!")
	case strings.HasPrefix(content, "hug"):
		cmdCount++
		sendReply(s, m, "HUG", "Hugs you!")
	case strings.HasPrefix(content, "kiss"):
		cmdCount++
		sendReply(s, m, "KISS", "Kisses you!")
	case strings.HasPrefix(content, "pat"):
		cmdCount++
		sendReply(s, m, "PAT", "Pats you!")
	case strings.HasPrefix(content, "poke"):
		cmdCount++
		sendReply(s, m, "POKE", "Pokes you!")
	case strings.HasPrefix(content, "slap"):
		cmdCount++
		sendReply(s, m, "SLAP", "Slaps you!")
	case strings.HasPrefix(content, "tickle"):
		cmdCount++
		sendReply(s, m, "TICKLE", "Tickles you!")
	case strings.HasPrefix(content, "cuddle"):
		cmdCount++
		sendReply(s, m, "CUDDLE", "Cuddles you!")
	case strings.HasPrefix(content, "blush"):
		cmdCount++
		sendReply(s, m, "BLUSH", "Blushes!")

	default:
		sendReply(s, m, "ERROR", fmt.Sprintf("This command does not exist or is not implemented yet. To find what you are looking for, type %shelp", getCurrentPrefix()))
	}

	log.Printf("Number of command cases: %d", cmdCount)
}

func logError(err error) {
	f, err := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error logging error: %v", err)
		return
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println(err)
}

func handleNowPlaying(s *discordgo.Session, m *discordgo.MessageCreate) {
	lastfmUser := config.LastFMAPIUsername
	lastfmApiKey := config.LastFMAPIKEY

	// Get Last.fm data
	lastfmURL := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json&limit=1", lastfmUser, lastfmApiKey)
	resp, err := http.Get(lastfmURL)
	if err != nil {
		log.Printf("Error fetching from Last.fm: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the currently playing song. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var lastfmResp struct {
		RecentTracks struct {
			Track []struct {
				Artist struct {
					Text string `json:"#text"`
				} `json:"artist"`
				Name string `json:"name"`
				Attr struct {
					NowPlaying string `json:"nowplaying"`
				} `json:"@attr"`
			} `json:"track"`
		} `json:"recenttracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&lastfmResp); err != nil {
		log.Printf("Error decoding Last.fm response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode song info. Please try again later.")
		return
	}

	if len(lastfmResp.RecentTracks.Track) == 0 {
		sendReply(s, m, "NOWPLAYING", "No recent tracks found.")
		return
	}

	track := lastfmResp.RecentTracks.Track[0]
	if !strings.EqualFold(track.Attr.NowPlaying, "true") {
		sendReply(s, m, "NOWPLAYING", "No song is currently playing.")
		return
	}

	artist := track.Artist.Text
	songName := track.Name

	// Get Spotify API token
	accessToken, err := getSpotifyAccessToken()
	if err != nil || accessToken == "" {
		log.Printf("Failed to get Spotify access token: %v", err)
		sendReply(s, m, "ERROR", "Failed to authenticate with Spotify.")
		return
	}

	// Search on Spotify
	searchURL := fmt.Sprintf("https://api.spotify.com/v1/search?q=track:%s%%20artist:%s&type=track&limit=1", url.QueryEscape(songName), url.QueryEscape(artist))
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Printf("Error creating Spotify search request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	searchResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error searching on Spotify: %v", err)
		return
	}
	defer searchResp.Body.Close()

	var searchResult struct {
		Tracks struct {
			Items []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(searchResp.Body).Decode(&searchResult); err != nil {
		log.Printf("Error decoding Spotify search response: %v", err)
		return
	}

	if len(searchResult.Tracks.Items) == 0 {
		sendReply(s, m, "NOWPLAYING", fmt.Sprintf("No Spotify results found for %s - %s", artist, songName))
		return
	}

	openSpotifyLink := searchResult.Tracks.Items[0].ExternalUrls.Spotify
	s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("[`%s`](%s) by **%s**\n", songName, openSpotifyLink, artist), m.Reference())
}

func handleAR(s *discordgo.Session, m *discordgo.MessageCreate) {
	mentionToggle = toggleMention()
	status := ifElse(mentionToggle, "on", "off")
	sendReply(s, m, "TOGGLE", fmt.Sprintf("Mention toggle is now %s.", status))
}

var mentionToggle = false

func toggleMention() bool {
	mentionToggle = !mentionToggle
	return mentionToggle
}

func mentionHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.Author.ID == s.State.User.ID {
		return
	}

	if m.Type == discordgo.MessageTypeReply && m.MessageReference != nil {
		refMsg, err := s.ChannelMessage(m.ChannelID, m.MessageReference.MessageID)
		if err == nil && refMsg.Author.ID == s.State.User.ID {
			return
		}
	}

	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID && mentionToggle && !m.MentionEveryone {
			sendReply(s, m, "MENTION", fmt.Sprintf(config.Afktext, m.Author.Username))
			return
		}
	}
}

func handleFakeHack(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Please provide a target username to hack.")
		return
	}

	target := args[1]

	sendReply(s, m, "FAKEHACK", fmt.Sprintf("Hacking %s's account...", target))

	go func() {
		time.Sleep(5 * time.Second)

		resp, err := http.Get("https://randomuser.me/api/")
		if err != nil {
			log.Printf("Error fetching hack info: %v", err)
			return
		}
		defer resp.Body.Close()

		var data struct {
			Results []struct {
				Login struct {
					Username string `json:"username"`
				} `json:"login"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Cell  string `json:"cell"`
				ID    struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"id"`
				Location struct {
					City   string `json:"city"`
					State  string `json:"state"`
					Street struct {
						Number int    `json:"number"`
						Name   string `json:"name"`
					} `json:"street"`
				} `json:"location"`
			} `json:"results"`
		}

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Printf("Error parsing hack info: %v", err)
			return
		}

		info := fmt.Sprintf("Hacked %s's account!\n\nUsername: %s\nEmail: %s\nPhone: %s\nCell: %s\nID: %s %s\nLocation: %s, %s %d %s",
			target, data.Results[0].Login.Username, data.Results[0].Email, data.Results[0].Phone, data.Results[0].Cell, data.Results[0].ID.Name, data.Results[0].ID.Value, data.Results[0].Location.City, data.Results[0].Location.State, data.Results[0].Location.Street.Number, data.Results[0].Location.Street.Name)

		sendReply(s, m, "FAKEHACK", info)
	}()
}

func getSpotifyAccessToken() (string, error) {
	clientID := config.SpotifyClientID
	clientSecret := config.SpotifyClientSecret

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func handleDictionary(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Please provide a word to search for.")
		return
	}

	word := args[1]

	resp, err := http.Get(fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word))
	if err != nil {
		log.Printf("Error fetching definition: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the definition. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		sendReply(s, m, "DICTIONARY", fmt.Sprintf("No definition found for %s.", word))
		return
	}

	var data []struct {
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
			} `json:"definitions"`
		} `json:"meanings"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Error decoding response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the definition. Please try again later.")
		return
	}

	if len(data) == 0 || len(data[0].Meanings) == 0 || len(data[0].Meanings[0].Definitions) == 0 {
		sendReply(s, m, "DICTIONARY", fmt.Sprintf("No definition found for %s.", word))
		return
	}

	def := data[0].Meanings[0].Definitions[0].Definition
	partOfSpeech := data[0].Meanings[0].PartOfSpeech

	sendReply(s, m, "DICTIONARY", fmt.Sprintf("%s (%s): %s", word, partOfSpeech, def))
}

func handlePurge(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID != config.OwnerID {
		return
	}

	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Please provide the amount of messages to purge.")
		return
	}

	amount, err := strconv.Atoi(args[1])
	if err != nil {
		sendReply(s, m, "ERROR", "Please provide a valid amount of messages to purge.")
		return
	}

	if amount < 1 || amount > 100 {
		sendReply(s, m, "ERROR", "Please provide a valid amount of messages to purge.")
		return
	}

	msgs, err := s.ChannelMessages(m.ChannelID, amount+1, "", "", "")
	if err != nil {
		sendReply(s, m, "ERROR", "Failed to purge messages. Please try again later.")
		return
	}

	for _, msg := range msgs {
		err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		if err != nil {
			log.Printf("Error deleting message %s: %v", msg.ID, err)
		}
	}

	sendReply(s, m, "PURGE", fmt.Sprintf("Purged %d messages.", amount))
}
func ifElse(condition bool, first string, second string) string {
	if condition {
		return first
	}
	return second
}

var jokes = []string{
	"I told my wife she was drawing her eyebrows too high. She looked surprised.",
	"Why don't scientists trust atoms? Because they make up everything.",
	"Why don't eggs tell jokes? They'd crack each other up.",
	"Why did the tomato turn red? Because it saw the salad dressing.",
	"What do you call a fake noodle? An impasta.",
	"Why did the scarecrow win an award? Because he was outstanding in his field.",
	"Why don't lobsters share? Because they're shellfish.",
	"What do you call a can opener that doesn't work? A can't opener.",
	"I'm reading a book about anti-gravity. It's impossible to put down.",
	"Why did the bicycle fall over? Because it was two-tired.",
	"What do you call a bear with no socks on? Barefoot.",
	"Why did the banana go to the doctor? He wasn't peeling well.",
	"Why did the chicken cross the playground? To get to the other slide.",
	"What do you call a group of cows playing instruments? A moo-sical band.",
	"Why did the baker go to the bank? He needed dough.",
	"Why did the mushroom go to the party? Because he was a fun-gi.",
	"Why did the pencil break up with the eraser? It was a sharp move.",
	"What do you call a fish with a sunburn? A star-fish.",
	"Why did the rabbit go to the doctor? He had hare-loss.",
	"Why did the computer go to the doctor? It had a virus.",
	"Why did the kid bring a ladder to school? He wanted to reach his full potential.",
	"What do you call a dog that does magic tricks? A labracadabrador.",
	"Why did the turkey join the band? He was a drumstick.",
	"Why did the cat join a band? Because it wanted to be the purr-cussionist.",
	"Why did the rabbit get kicked out of the bar? He was making too many hare-brained jokes.",
	"Why did the computer screen go to therapy? It was feeling a little glitchy.",
	"Why did the banana get kicked out of the fruit stand? He wasn't peeling well.",
	"Why did the egg go to therapy? It was cracking under the pressure.",
}

func handleJoke(s *discordgo.Session, m *discordgo.MessageCreate) {
	joke := jokes[rand.Intn(len(jokes))]
	sendReply(s, m, "JOKE", joke)
}
func showHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) == 1 {
		sendIni(s, m, fmt.Sprintf(" HELP -  Type %shelp <category> to view a specific help page", config.Prefix), fmt.Sprintf("[•] UTILITIES\n[•] FUN\n[•] INFORMATION\n[•] NSFW\n[•] MUSIC"))
		return
	}
	category := strings.ToUpper(args[1])
	switch category {
	case "UTILITIES":
		showUtilitiesHelp(s, m)
	case "FUN":
		showFunHelp(s, m)
	case "INFORMATION":
		showInformationHelp(s, m)
	case "NSFW":
		showNSFWHelp(s, m)
	case "MUSIC":
		showMusicHelp(s, m)
	default:
		sendReply(s, m, "ERROR", fmt.Sprintf("Unknown category `%s`. Type `%shelp` to view all categories", category, config.Prefix))
	}
}

func showUtilitiesHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "UTILITIES - HELP", fmt.Sprintf("\n[•] %sar: Toggle mention notifications\n[•] %sdefine <word>: Get the definition of a word\n[•] %safktext <text>: Change the afk text and if you want to put the user in there use <<username>> that will include the username of the person tagging you!\n[•] %ssetprefix <prefix>: Change the prefix\n[•] %sprofile <mention>: Get the profile picture of a user\n[•] %savatar <mention>: Get the avatar of a user\n[•] %ssetemoji <emoji>: Set the autoreact emoji\n[•] %sautoreact <emoji>: Set the autoreact emoji\n[•] %sserverinfo: Get information about the server", getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix()))
}

func showFunHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "FUN - HELP", fmt.Sprintf("[•] %sjoke: Sends a random joke\n[•] %sdog: Sends a random dog picture\n[•] %smotivate: Sends a random motivational quote\n[•] %squote: Sends a random quote\n[•] %srizz: Sends a random rizz\n[•] %sweather <location>: Sends the weather\n[•] %smeme: Sends a random meme\n[•] %sfact: Sends a random fact\n[•] %stod <truth or dare>: Sends a random true or dare\n[•] %scat : Sends a random cat picture\n[•] %sdick <mention>: measure someone's dick\n[•] %sdadjoke: Sends a random dad joke\n[•] %s8ball <question>: Ask the magic 8ball a question\n[•] %sap <mention>: chatpack a mention\n[•] %scoinflip: Flip a coin\n[•] %sfakehack <mention>: Fake hack a mention", getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix()))
}

func showInformationHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "INFORMATION - HELP", fmt.Sprintf("\n[•] %sci: Shows the bot's current information\n[•] %sch: Shows the bot's changelogs\n[•] %shwd: Shows the bot's hardware information", getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix()))
}

func showNSFWHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "NSFW - HELP", fmt.Sprintf("\n[•] %sboobs: Sends a random boobs picture\n[•] %shentai: Sends a random hentai picture\n[•] %sblowjob: Sends a random blowjob picture", getCurrentPrefix(), getCurrentPrefix(), getCurrentPrefix()))
}

func showMusicHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "MUSIC - HELP", fmt.Sprintf("\n[•] %snp: Shows the currently playing song\n%sas: Shows how many times you've listened to currently playing artist", getCurrentPrefix(), getCurrentPrefix()))
}

func showMultiTokenHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "MulitToken - HELP", fmt.Sprintf("\n[•] %snp: Shows the currently playing song", getCurrentPrefix()))
}

func handleTrueOrDare(s *discordgo.Session, m *discordgo.MessageCreate) {
	kind := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(m.Content, ".tod")))
	if kind == "" {
		sendReply(s, m, "ERROR", "Please specify if you want a truth or a dare")
		return
	}
	if kind != "truth" && kind != "dare" {
		sendReply(s, m, "ERROR", "Invalid kind of question. Please use either 'truth' or 'dare'")
		return
	}
	var truths []string
	switch kind {
	case "truth":
		truths = []string{
			"What is the most inappropriate thing you've ever done?",
			"What is the most dirty thing you've ever thought of?",
			"What is the most weird thing you've ever done in your life?",
			"What is the most perverted thing you've ever thought of?",
			"What is the most ridiculous thing you've ever done?",
			"What's the worst thing you've ever done to someone?",
			"What's the most disgusting thing you've ever done?",
			"What's the most disturbing thing you've ever seen?",
			"What's the most messed up thing you've ever thought of?",
			"What's the most messed up thing you've ever done?",
			"What's the most messed up thing you've ever seen?",
			"What's the most messed up thing you've ever heard?",
			"What's the most messed up thing you've ever experienced?",
		}
	case "dare":
		truths = []string{
			"I dare you to do a funny dance in front of your friends",
			"I dare you to do a crazy stunt in front of your family",
			"I dare you to do a weird facial expression in front of your friends",
			"I dare you to do a silly walk in front of your family",
			"I dare you to do a ridiculous pose in front of your friends",
		}
	default:
		return
	}
	question := truths[rand.Intn(len(truths))]
	sendReply(s, m, kind, fmt.Sprintf("%s: %s", kind, question))
}

func handleDare(s *discordgo.Session, m *discordgo.MessageCreate) {
	dares := []string{
		"I dare you to do a funny dance in front of your friends",
		"I dare you to do a crazy stunt in front of your family",
		"I dare you to do a weird facial expression in front of your friends",
		"I dare you to do a silly walk in front of your family",
		"I dare you to do a ridiculous pose in front of your friends",
	}
	dare := dares[rand.Intn(len(dares))]
	sendReply(s, m, "DARE", fmt.Sprintf("%s: %s", "dare", dare))
}
func handle8Ball(s *discordgo.Session, m *discordgo.MessageCreate) {
	question := strings.TrimSpace(strings.TrimPrefix(m.Content, ".8ball"))
	if question == "" {
		sendReply(s, m, "ERROR", "Please ask a question")
		return
	}
	response := eightBall[rand.Intn(len(eightBall))]
	sendReply(s, m, "8BALL", fmt.Sprintf("Q: %s\nA: %s", question, response))
}

func handleCat(s *discordgo.Session, m *discordgo.MessageCreate) {
	catURL := "https://thecatapi.com/api/images/get?format=src&type=png&t=" + time.Now().Format("150405")
	sendReply(s, m, "Meow!", "Here is your cat!")
	_, err := s.ChannelMessageSendReply(m.ChannelID, ""+catURL+"", m.Reference())
	if err != nil {
		log.Printf("Error sending cat picture: %v", err)
	}
}

func handleDick(s *discordgo.Session, m *discordgo.MessageCreate) {
	target := m.Author
	if len(m.Mentions) > 0 {
		target = m.Mentions[0]
	}

	today := time.Now().Format("2006-01-02")
	size := calculateDickSize(target.ID + today)
	dickArt := fmt.Sprintf("8%sD", strings.Repeat("=", size))

	sendReply(s, m, "DICK SIZE", fmt.Sprintf("%s: %d inches\n%s",
		target.Username, size, dickArt))
}

func handleRandomFact(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://uselessfacts.jsph.pl/random.json?language=en"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching random fact: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch a random fact. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Text string `json:"text"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding random fact response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the random fact. Please try again later.")
		return
	}

	sendReply(s, m, "RANDOM FACT", result.Text)
}

func handleMotivationQuote(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://api.quotable.io/random"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching motivational quote: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch a motivational quote. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding motivational quote response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the motivational quote. Please try again later.")
		return
	}

	sendReply(s, m, "MOTIVATIONAL QUOTE", fmt.Sprintf("%s - %s", result.Content, result.Author))
}

var (
	intervalTimer *time.Timer
	mu            = &sync.Mutex{}
	doneChan      = make(chan struct{})
)

func guildRotator(s *discordgo.Session, m *discordgo.MessageCreate) {
	mu.Lock()
	defer mu.Unlock()

	if len(config.GuildIDs) == 0 {
		sendReply(s, m, "ERROR", "No guilds found to rotate.")
		return
	}

	if intervalTimer != nil {
		intervalTimer.Stop()
	}

	ticker := time.NewTicker(15 * time.Second)
	index := 0
	go func() {
		for {
			select {
			case <-ticker.C:
				if config.RotateGuilds {
					rotateGuild(s.Token, []string{config.GuildIDs[index]}, 15*time.Minute)
				}
				index = (index + 1) % len(config.GuildIDs)
			case <-doneChan:
				ticker.Stop()
				return
			}
		}
	}()
	sendReply(s, m, "GUILD ROTATOR", "Guild rotator has been started.")
}

func stopGuildRotator() {
	mu.Lock()
	defer mu.Unlock()
	sendReply(nil, nil, "GUILD ROTATOR", "Guild rotator has been stopped.")
	close(doneChan)
	doneChan = make(chan struct{})
}
func rotateGuild(token string, guildIds []string, delay time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	index := 0
	for range ticker.C {
		if !config.RotateGuilds {
			return
		}

		guildId := guildIds[index]
		body, err := json.Marshal(map[string]interface{}{
			"identity_guild_id": guildId,
			"identity_enabled":  true,
		})
		if err != nil {
			log.Printf("Error marshaling body: %v", err)
			return
		}

		req, err := http.NewRequest("PUT", "https://discord.com/api/v9/users/@me/clan", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error making request: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				return
			}
			log.Printf("Error changing to guild ID: %s: %s", guildId, string(body))
			return
		}

		index = (index + 1) % len(guildIds)
	}
}

func handleDogPicture(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://dog.ceo/api/breeds/image/random"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching dog picture: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch a dog picture. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Message string `json:"message"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding dog picture response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the dog picture. Please try again later.")
		return
	}

	sendReply(s, m, "DOG PICTURE", "Here is your dog!")
	_, err = s.ChannelMessageSend(m.ChannelID, result.Message)
	if err != nil {
		log.Printf("Error sending dog picture: %v", err)
	}
}

func handleRandomJoke(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://official-joke-api.appspot.com/jokes/random"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching random joke: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch a random joke. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Setup     string `json:"setup"`
		Punchline string `json:"punchline"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding random joke response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the random joke. Please try again later.")
		return
	}

	sendReply(s, m, "RANDOM JOKE", fmt.Sprintf("%s - %s", result.Setup, result.Punchline))
}

func handleWxrnInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Please provide a Discord ID.")
		return
	}

	discordID := args[1]
	resp, err := http.Get(fmt.Sprintf("https://api.wxrn.lol/discord/%s", discordID))
	if err != nil {
		log.Printf("Error fetching discord info: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the information. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		sendReply(s, m, "discord INFO", "No information found for the given Discord ID.")
		return
	}

	var data struct {
		Username     string `json:"username"`
		DisplayName  string `json:"displayName"`
		AvatarURL    string `json:"avatarUrl"`
		BannerURL    string `json:"bannerUrl"`
		ProfileDecor string `json:"profileDecorationUrl"`
		ResponseTime string `json:"response_time"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Error decoding discord response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the information. Please try again later.")
		return
	}

	info := fmt.Sprintf("[•] Username: %s\n[•] Display Name: %s\n[•] Avatar: %s\n[•] Banner: %s\n[•] Profile Decoration: %s\n[•] Response Time: %s",
		data.Username, data.DisplayName, data.AvatarURL, data.BannerURL, data.ProfileDecor, data.ResponseTime)

	sendIni(s, m, "DISCORD INFO", info)
}

func handleCredits(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "CLIENT INFO", fmt.Sprintf("\n[•] Developer: anthem\n[•] Version: 2H5\n[•] .go Files: 34\n[•] Commands: 43\n[•] Code Lines: 2676\n[•] Uptime: %s\n[•] Memory: %.2f MB\n[•] Current Prefix: %s ", time.Since(startTime).Round(time.Second), getMemoryUsage(), config.Prefix))
}

func handleChangelogs(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "Changelogs v2H5", fmt.Sprintf("\n[•] Added Weather Support"))
}

func handlehwd(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendIni(s, m, "Hardware info", fmt.Sprintf("\n[•] CPU: AMD Ryzen Threadripper PRO 5995WX \n[•] RAM: 512GB DDR4\n[•] Storage: 7TB\n[•] OS: ARCHB\n[•] Uplink: 10GB\n[•] Downlink: 10GB"))
}
func handleFemboy(s *discordgo.Session, m *discordgo.MessageCreate) {
	target := m.Author
	if len(m.Mentions) > 0 {
		target = m.Mentions[0]
	}

	percentage := rand.Intn(101)

	sendReply(s, m, "FEMBOY PROCENTAGE", fmt.Sprintf("%s is: %d%% a femboy :3", target.Username, percentage))
}

func getCurrentPrefix() string {
	return config.Prefix
}

func handleFunFact(s *discordgo.Session, m *discordgo.MessageCreate) {
	funFacts := []string{
		"Honey never spoils.",
		"Bananas are berries, but strawberries aren't.",
		"Octopuses have three hearts.",
		"A day on Venus is longer than a year on Venus.",
		"A group of flamingos is called a 'flamboyance'.",
		"Octopuses have more than 300 suckers on their arms.",
		"The world's largest living organism is a fungus.",
		"Antarctica has the highest average elevation of any continent.",
		"The world's largest waterfall is underwater.",
		"The world's largest snowflake was 15 inches wide.",
		"The world's largest recorded snowfall was 75.8 feet.",
		"The world's largest living species of lizard is the Komodo dragon.",
		"The world's largest living species of snake is the green anaconda.",
		"The world's largest living species of spider is the Goliath birdeater.",
		"The world's largest living species of insect is the Atlas beetle.",
	}
	randomFact := funFacts[rand.Intn(len(funFacts))]
	sendReply(s, m, "FUN FACT", randomFact)
}
func handleCoinFlip(s *discordgo.Session, m *discordgo.MessageCreate) {
	outcome := "Heads"
	if rand.Intn(2) == 0 {
		outcome = "Tails"
	}
	sendReply(s, m, "COINFLIP", fmt.Sprintf("The coin landed on: %s", outcome))
}
func handleMeme(s *discordgo.Session, m *discordgo.MessageCreate) {
	memeURL := "https://meme-api.com/gimme"
	resp, err := http.Get(memeURL)
	if err != nil {
		log.Printf("Error fetching meme: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the meme. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var meme struct {
		PostLink  string   `json:"postLink"`
		Subreddit string   `json:"subreddit"`
		Title     string   `json:"title"`
		Url       string   `json:"url"`
		Nsfw      bool     `json:"nsfw"`
		Spoiler   bool     `json:"spoiler"`
		Author    string   `json:"author"`
		Ups       int      `json:"ups"`
		Preview   []string `json:"preview"`
	}

	err = json.NewDecoder(resp.Body).Decode(&meme)
	if err != nil {
		log.Printf("Error parsing meme response: %v", err)
		sendReply(s, m, "ERROR", "Failed to parse the meme. Please try again later.")
		return
	}

	sendIni(s, m, "MEME", fmt.Sprintf("[•] Title:%s\n[•] Link: %s\n[•] Subreddit: %s\n[•] Author: %s\n[•] Ups: %d", meme.Title, meme.PostLink, meme.Subreddit, meme.Author, meme.Ups))
	_, err = s.ChannelMessageSend(m.ChannelID, meme.Preview[0])
	if err != nil {
		log.Printf("Error sending meme: %v", err)
	}
}
func HandleRizz(s *discordgo.Session, m *discordgo.MessageCreate) {
	target := m.Author
	if len(m.Mentions) > 0 {
		target = m.Mentions[0]
	}

	url := "https://rizzapi.vercel.app/random"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching rizz: %v", err)
		return
	}
	defer resp.Body.Close()

	var rizz struct {
		Text string `json:"text"`
	}
	err = json.NewDecoder(resp.Body).Decode(&rizz)
	if err != nil {
		log.Printf("Error parsing rizz response: %v", err)
		return
	}

	sendReply(s, m, "RIZZ", fmt.Sprintf("%s, %s", target.Username, rizz.Text))
}

func calculateDickSize(input string) int {
	hash := fnv.New32a()
	hash.Write([]byte(input))
	return int(hash.Sum32()%12) + 3
}

func sendReply(s *discordgo.Session, m *discordgo.MessageCreate, title, content string) {
	msg := fmt.Sprintf("```ansi\n[34m[KITTY][37m %s\n```\n```\n%s```\n-# by @94kx <3 ", title, content)
	s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
}

func sendIni(s *discordgo.Session, m *discordgo.MessageCreate, title, content string) {
	msg := fmt.Sprintf("```ansi\n[34m[KITTY][37m %s\n```\n```ini\n%s\n```\n-# by @94kx <3", title, content)
	s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
}

var autoCurseActive = false

var autoCurseTarget *discordgo.User

func handleAutoCurse(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Mentions) == 0 {
		sendReply(s, m, "ERROR", "Please mention a user to start auto curse.")
		return
	}

	autoCurseTarget = m.Mentions[0]
	autoCurseActive = true
	sendReply(s, m, "AUTO CURSE", fmt.Sprintf("Auto curse has been started on %s for 30 seconds.", autoCurseTarget.Username))

	go func() {
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()

		for autoCurseActive {
			select {
			case <-timer.C:
				autoCurseActive = false
				autoCurseTarget = nil
				sendReply(s, m, "AUTO CURSE", "Auto curse has been stopped.")
				return
			default:
				curse := fmt.Sprintf("# %s <@%s>", getCurseWord(), autoCurseTarget.ID)
				_, err := s.ChannelMessageSend(m.ChannelID, curse)
				if err != nil {
					log.Printf("Error sending curse word: %v", err)
				}
				time.Sleep(150 * time.Millisecond)
			}
		}
	}()
}

func getCurseWord() string {
	data, err := ioutil.ReadFile("./files/words.txt")
	if err != nil {
		log.Fatalf("Error reading curse words file: %v", err)
	}
	curseWords := strings.Split(string(data), "\n")
	return curseWords[rand.Intn(len(curseWords))]
}

func handleChangeAfkText(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Mentions) > 0 {
		sendReply(s, m, "ERROR", "Please don't mention anyone when changing the afk text.")
		return
	}

	if len(m.Content) < 9 {
		sendReply(s, m, "ERROR", "Please provide a valid afk text.")
		return
	}

	afkText := m.Content[8:]
	if strings.Contains(afkText, "<<username>>") {
		afkText = strings.ReplaceAll(afkText, "<<username>>", "%s")
	} else {
		afkText = fmt.Sprintf("%s", afkText)
	}
	config.Afktext = afkText
	saveConfig()
	sendReply(s, m, "AFK TEXT", fmt.Sprintf("The afk text has been changed to: %s", afkText))
}

func handleChangePrefix(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := strings.TrimPrefix(m.Content, config.Prefix)
	args := strings.Fields(content)

	if len(args) < 2 || args[0] != "setprefix" {
		sendReply(s, m, "ERROR", "Usage: setprefix [new_prefix]\nValid prefixes: !, &, ;, :")
		return
	}

	newPrefix := args[1]
	if newPrefix != "!" && newPrefix != "&" && newPrefix != ";" && newPrefix != ":" && newPrefix != "." && newPrefix != "," && newPrefix != "/" && newPrefix != "?" && newPrefix != "off" {
		sendReply(s, m, "ERROR", fmt.Sprintf("The prefix %s is not valid. The prefix can be !, &, ;, or : or off to turn off the Prefix", newPrefix))
		return
	}

	if newPrefix == "off" {
		config.Prefix = ""
	} else {
		config.Prefix = newPrefix
	}
	saveConfig()
	sendReply(s, m, "PREFIX", fmt.Sprintf("The command prefix has been changed to: %s", config.Prefix))
}

func saveConfig() {
	data, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		log.Fatalf("Error encoding config: %v", err)
	}

	err = ioutil.WriteFile("config.json", data, 0644)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
	}
}

func handleBoobs(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://api.nekosapi.com/v4/images/random?&tags=large_breasts&limit=1"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching boobs image: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the boobs image. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result []struct {
		Url string `json:"url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding boobs image response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the boobs image. Please try again later.")
		return
	}

	sendReply(s, m, "BOOBS", "Here is your boobs ya gooner!")
	_, err = s.ChannelMessageSend(m.ChannelID, result[0].Url)
	if err != nil {
		log.Printf("Error sending boobs image: %v", err)
	}
}

func handleHentai(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://api.waifu.pics/nsfw/waifu"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching hentai image: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the hentai image. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Url string `json:"url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding hentai image response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the hentai image. Please try again later.")
		return
	}
	sendReply(s, m, "HENTAI", "Here is your hentai ya gooner!")
	_, err = s.ChannelMessageSend(m.ChannelID, result.Url)
	if err != nil {
		log.Printf("Error sending hentai image: %v", err)
	}
}

func handleBlowjob(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://api.waifu.pics/nsfw/blowjob"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching blowjob image: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the blowjob image. Please try again later.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Url string `json:"url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding blowjob image response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the blowjob image. Please try again later.")
		return
	}
	sendReply(s, m, "BLOWJOB", "Here is your blowjob ya gooner!")
	_, err = s.ChannelMessageSend(m.ChannelID, result.Url)
	if err != nil {
		log.Printf("Error sending blowjob image: %v", err)
	}
}
func handleServerInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error fetching guild info: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the guild info. Please try again later.")
		return
	}

	owner, err := s.GuildMember(guild.ID, guild.OwnerID)
	if err != nil {
		log.Printf("Error fetching owner info: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the owner info. Please try again later.")
		return
	}
	guildID, err := strconv.ParseInt(guild.ID, 10, 64)
	if err != nil {
		log.Printf("Error converting guild ID: %v", err)
		sendReply(s, m, "ERROR", "Failed to parse guild ID. Please try again later.")
		return
	}
	createdAt := time.Unix((guildID>>22)+1420070400000, 0).Format(time.UnixDate)
	info := fmt.Sprintf("[•] Name: %s\n[•] ID: %s\n[•] Owner: %s\n[•] Creation Time: %s\n[•] Region: %s\n[•] Verification Level: %d\n[•] Total Members: %d\n[•] Total Roles: %d\n[•] Total Channels: %d\n[•] Total Emotes: %d",
		guild.Name, guild.ID, owner.User.Username, createdAt, guild.Region, guild.VerificationLevel, len(guild.Members), len(guild.Roles), len(guild.Channels), len(guild.Emojis))
	sendIni(s, m, "SERVER INFO", info)
}

func addReactionHandler(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID != s.State.User.ID && config.AutoReactToggle && config.AutoReactEmoji != "" {
		err := s.MessageReactionAdd(m.ChannelID, m.MessageID, config.AutoReactEmoji)
		if err != nil {
			log.Printf("Error sending reaction: %v", err)
		}
	}
}

func handleSetAutoReact(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Please provide an emoji for auto-react.")
		return
	}

	emoji := args[1]
	config.AutoReactEmoji = getEmoji(s, emoji)
	sendReply(s, m, "AUTOREACT", fmt.Sprintf("Auto-react emoji set to: %s", config.AutoReactEmoji))
}

func handleToggleAutoReact(s *discordgo.Session, m *discordgo.MessageCreate) {
	config.AutoReactToggle = !config.AutoReactToggle
	status := ifElse(config.AutoReactToggle, "on", "off")
	sendReply(s, m, "AUTOREACT", fmt.Sprintf("Auto-react toggle is now %s", status))
}

var autoReactEmoji string
var autoReactToggle bool

func getEmoji(s *discordgo.Session, emoji string) string {
	emojis, err := s.GuildEmojis(s.State.Guilds[0].ID)
	if err != nil {
		return ""
	}

	for _, e := range emojis {
		if e.Name == emoji {
			return fmt.Sprintf("<:%s:%s>", e.Name, e.ID)
		}
	}

	return fmt.Sprintf("%s", emoji)
}

func handleWeather(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Please provide a city name.")
		return
	}

	city := strings.Join(args[1:], " ")
	resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=9de243494c0b295cca9337e1e96b00e2&units=metric", url.QueryEscape(city)))
	if err != nil {
		log.Printf("Error fetching weather info: %v", err)
		sendReply(s, m, "ERROR", "Failed to fetch the weather information. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		sendReply(s, m, "WEATHER", "No weather information found for the given city.")
		return
	}

	var data struct {
		Weather []struct {
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Main struct {
			Temp     float64 `json:"temp"`
			Pressure int     `json:"pressure"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Error decoding weather response: %v", err)
		sendReply(s, m, "ERROR", "Failed to decode the weather information. Please try again later.")
		return
	}

	weather := fmt.Sprintf("[•] City: %s\n[•] Temperature: %.2f \u00B0C\n[•] Pressure: %d mbar\n[•] Humidity: %d%%\n[•] Wind Speed: %.2f m/s\n[•] Weather: %s\n[•] Icon: http://openweathermap.org/img/w/%s.png",
		city, data.Main.Temp, data.Main.Pressure, data.Main.Humidity, data.Wind.Speed, data.Weather[0].Description, data.Weather[0].Icon)

	sendIni(s, m, "WEATHER", weather)
}

func handleMultiTokenSpam(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 4 {
		sendReply(s, m, "ERROR", "Usage: multispam <channelID> <message> <duration>")
		return
	}

	channelID := args[1]
	message := strings.Join(args[2:len(args)-1], " ")
	durationStr := args[len(args)-1]

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		sendReply(s, m, "ERROR", "Invalid duration format. Please use a valid time format (e.g., 10s, 1m).")
		return
	}

	b, err := ioutil.ReadFile("files/token.txt")
	if err != nil {
		log.Printf("Error reading token file: %v", err)
		sendReply(s, m, "ERROR", "Failed to read the token file. Please make sure the file is in the correct location and has the correct permissions.")
		return
	}

	tokens := strings.Split(string(b), "\n")

	for _, token := range tokens {
		if token == "" {
			continue
		}

		go func(t string) {
			dg, err := discordgo.New(t)
			if err != nil {
				log.Printf("Error logging in with token %s: %v", t, err)
				return
			}

			err = dg.Open()
			if err != nil {
				log.Printf("Error opening WebSocket for token %s: %v", t, err)
				return
			}
			defer dg.Close()

			timer := time.NewTimer(duration)
			defer timer.Stop()

			for {
				select {
				case <-timer.C:
					log.Printf("Ending spam with token %s", t)
					return
				default:
					_, err = dg.ChannelMessageSend(channelID, message)
					if err != nil {
						log.Printf("Error sending message with token %s: %v", t, err)
						return
					}
					time.Sleep(200 * time.Millisecond) // Adjust the sleep duration as needed
				}
			}
		}(token)
	}
	sendReply(s, m, "MULTISPAM", "Started spamming with multiple tokens.")
}

var globalLimiter = rate.NewLimiter(rate.Every(700*time.Millisecond), 1)

func handleMultiTokenInvite(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		sendReply(s, m, "ERROR", "Usage: multinvite <inviteCode>")
		return
	}

	inviteCode := args[1]

	tokenBytes, err := os.ReadFile("files/token.txt")
	if err != nil {
		log.Printf("Error reading token file: %v", err)
		sendReply(s, m, "ERROR", "Failed to read token file.")
		return
	}
	tokens := strings.Split(string(tokenBytes), "\n")

	proxyStrings := readLines("files/http.txt")
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i, token := range tokens {
		if token == "" {
			continue
		}

		var proxy string
		if i < len(proxyStrings) {
			proxy = proxyStrings[i]
		}

		sem <- struct{}{}
		wg.Add(1)

		go func(token, proxy string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			if err := globalLimiter.Wait(context.Background()); err != nil {
				log.Printf("Rate limiter error for token %s: %v", token, err)
				return
			}

			err := joinWithToken(inviteCode, token, proxy)
			if err != nil {
				log.Printf("Failed to join with token %s: %v", token, err)
			}
		}(token, proxy)
	}

	wg.Wait()
	sendReply(s, m, "MULTIINVITE", "Finished joining with multiple tokens.")
}

// Helper to read lines from a file
func readLines(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func joinWithToken(inviteCode, token, proxy string) error {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	if proxy != "" {
		if !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
			proxy = "http://" + proxy
		}

		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.Printf("Invalid proxy %s: %v", proxy, err)
			return err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	apiURL := fmt.Sprintf("https://canary.discord.com/api/v9/invites/%s", inviteCode)
	body := bytes.NewBufferString("{}")

	putReq, _ := http.NewRequest("PUT", apiURL, body)
	putReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0")
	putReq.Header.Set("Accept", "*/*")
	putReq.Header.Set("Accept-Language", "fr,fr-FR;q=0.8,en-US;q=0.5,en;q=0.3")
	putReq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("X-Context-Properties", "eyJsb2NhdGlvbiI6IkpvaW4gR3VpbGQiLCJsb2NhdGlvbl9ndWlsZF9pZCI6Ijk4OTkxOTY0NTY4MTE4ODk1NCIsImxvY2F0aW9uX2NoYW5uZWxfaWQiOiI5OTAzMTc0ODgxNzg4NjgyMjQiLCJsb2NhdGlvbl9jaGFubmVsX3R5cGUiOjB9")
	putReq.Header.Set("Authorization", token)
	putReq.Header.Set("X-Super-Properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiRmlyZWZveCIsImRldmljZSI6IiIsInN5c3RlbV9sb2NhbGUiOiJmciIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQ7IHJ2OjEwMi4wKSBHZWNrby8yMDEwMDEwMSBGaXJlZm94LzEwMi4wIiwiYnJvd3Nlcl92ZXJzaW9uIjoiMTAyLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MTM2MjQwLCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")
	putReq.Header.Set("X-Discord-Locale", "en-US")
	putReq.Header.Set("X-Debug-Options", "bugReporterEnabled")
	putReq.Header.Set("Origin", "https://discord.com")
	putReq.Header.Set("DNT", "1")
	putReq.Header.Set("Connection", "keep-alive")
	putReq.Header.Set("Referer", "https://discord.com")
	putReq.Header.Set("Cookie", "__dcfduid=21183630021f11edb7e89582009dfd5e; __sdcfduid=21183631021f11edb7e89582009dfd5ee4936758ec8c8a248427f80a1732a58e4e71502891b76ca0584dc6fafa653638; locale=en-US")
	putReq.Header.Set("Sec-Fetch-Dest", "empty")
	putReq.Header.Set("Sec-Fetch-Mode", "cors")
	putReq.Header.Set("Sec-Fetch-Site", "same-origin")
	putReq.Header.Set("TE", "trailers")

	resp, err := client.Do(putReq)
	if err != nil {
		return fmt.Errorf("PUT request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✅ Successfully joined with token %s via proxy %s", token, proxy)
		return nil
	} else if resp.StatusCode == 429 {
		var retryAfter float64
		if err := json.Unmarshal(bodyBytes, &map[string]interface{}{"retry_after": &retryAfter}); err == nil {
			log.Printf("⏳ Ratelimited. Retrying after %.1f seconds", retryAfter)
			time.Sleep(time.Duration(retryAfter*1000) * time.Millisecond)
			return joinWithToken(inviteCode, token, proxy)
		}
	} else if resp.StatusCode == 403 {
		log.Printf("🔒 Token locked or CAPTCHA required: %s", token)
	} else {
		log.Printf("❌ Failed with token %s via proxy %s. Status: %d, Body: %s", token, proxy, resp.StatusCode, bodyBytes)
	}

	return nil
}
