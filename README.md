# Kitty Bot

**Kitty Bot** is a powerful, multi-functional Discord bot developed by **Kawa**, designed for utility, fun, and automation. Built with the `discordgo` library, this bot offers a wide range of commands from moderation and music to NSFW content, server analytics, and advanced multi-token operations.

---

## 🌟 Features

### 🔧 Utilities
- **Auto React**: Automatically react to messages with a custom emoji.
- **Auto AFK**: Set a custom AFK message with dynamic username tagging.
- **Purge Messages**: Bulk delete messages in a channel.
- **Server Info**: View detailed server information including owner, creation date, and member count.
- **User Profile Lookup**: Fetch user data via external APIs (e.g., `wxrn.lol`).
- **Custom Prefix**: Fully customizable command prefix (`!`, `.`, `?`, etc.).

### 🎉 Fun & Entertainment
- **8-Ball**: Ask the magic 8-ball a question.
- **Coin Flip**: Flip a virtual coin.
- **Jokes & Dad Jokes**: Get random jokes or dad jokes.
- **Truth or Dare**: Play truth or dare with dynamic prompts.
- **Rizz Generator**: Send smooth pickup lines.
- **Memes**: Pull random memes from Reddit.
- **Random Facts & Quotes**: Motivational quotes and fun facts.
- **Pet Commands**: `hug`, `kiss`, `pat`, `slap`, `cuddle`, and more.

### 🎵 Music Integration
- **Now Playing (NP)**: Fetch the currently playing song from Last.fm.
- **Spotify Linking**: Auto-search and link the current track on Spotify.
- **Artist Stats**: View play counts and listening stats.

### 🐾 Image & Media
- **Random Cat/Dog Pics**: Get cute animal images.
- **NSFW Content** (18+):
  - `boobs`
  - `hentai`
  - `blowjob`
  *(Note: Use responsibly and in appropriate channels.)*

### 🌦️ Weather
- Get real-time weather for any city with temperature, humidity, wind speed, and more.

### ⚙️ Advanced Tools
- **Multi-Token Spam**: Send messages across accounts using multiple tokens.
- **Multi-Token Invite Joiner**: Mass-join servers using token lists and proxies.
- **Guild Rotator**: Rotate identity guilds for account customization.
- **Fake Hack Command**: Fun, fake hacking simulation with random user data.

### 📊 Bot Info & Monitoring
- **Uptime & Memory Usage**: Monitor bot performance.
- **Hardware Info**: View server specs (customizable).
- **Changelogs & Credits**: Track updates and give credit.

---

## 🛠️ Setup & Configuration

### Prerequisites
- Go 1.19+
- Discord Bot Token
- API Keys:
  - Last.fm API
  - Spotify Client ID & Secret (for Spotify linking)
  - OpenWeatherMap API key (for weather)
- Configured `config.json`

### Installation

1. Clone the repository:
```bash
git clone [https://github.com/Kawa/kitty-bot.git](https://github.com/2kpz/Selfbot.git)
cd selfbot
```

2. Install dependencies:
```bash
go mod tidy
```

3. Configure `config.json`:
```json
{
  "token": "YOUR_BOT_TOKEN",
  "owner_id": "YOUR_USER_ID",
  "lastFMAPIKEY": "LASTFM_API_KEY",
  "lastFMAPIUsername": "LASTFM_USERNAME",
  "SpotifyClientID": "SPOTIFY_CLIENT_ID",
  "SpotifyClientSecret": "SPOTIFY_CLIENT_SECRET",
  "prefix": ".",
  "autoReactToggle": false,
  "autoReactEmoji": "❤️",
  "guild_ids": ["GUILD_ID_1", "GUILD_ID_2"],
  "rotate_guilds": true
}
```

4. Run the bot:
```bash
go run main.go
```

---

---

## 🚀 Commands

| Category       | Commands |
|----------------|--------|
| **Utilities**  | `ar`, `afktext`, `setprefix`, `purge`, `serverinfo`, `profile`, `togglereact`, `setemoji` |
| **Fun**        | `8ball`, `joke`, `dadjoke`, `meme`, `rizz`, `tod`, `coinflip`, `cat`, `dog` |
| **NSFW**       | `boobs`, `hentai`, `blowjob` |
| **Music**      | `np` |
| **Info**       | `ci`, `ch`, `hwd`, `weather`, `fact`, `quote` |
| **Multi-Token**| `tspam`, `tjoiner` |

> Use `.help` in Discord to see full command list.

---

## ⚠️ Disclaimer

- This bot is for **educational and entertainment purposes only**.
- Misuse of multi-token features may violate Discord’s ToS.
- NSFW commands should only be used in appropriate, age-restricted channels.
- The developer is not responsible for misuse of this software.

---

## 🙌 Credits

- **Developer**: [Kawa](https://github.com/2kpz)
- **Special Thanks**: 94kx (design inspiration), DiscordGo community
- **APIs Used**: Last.fm, Spotify, OpenWeatherMap, Waifu.pics, NekosAPI, meme-api.com

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.

---

> ✨ **Made with love by Kawa** | `@94kx <3` | `v2H5`
