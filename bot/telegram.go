package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/R0zark/transmission-telegram-bot/config"
	"github.com/R0zark/transmission-telegram-bot/dockerhandler"
	"github.com/R0zark/transmission-telegram-bot/transmission"
	"github.com/R0zark/transmission-telegram-bot/yamlhandler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Bot struct holds the Telegram bot
type Bot struct {
	BotAPI *tgbotapi.BotAPI
	Config config.Config
}

// NewBot initializes a new Telegram bot
func NewBot(config *config.Config) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(config.Telegram.BotToken)
	if err != nil {
		return nil, err
	}
	return &Bot{BotAPI: botAPI, Config: *config}, nil
}

// Start updates handler for the bot
func (b *Bot) Start(transmission *transmission.Client) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.BotAPI.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	log.Println("Bot ready.")

	for update := range updates {
		if update.Message == nil || !b.CheckWhiteList(strconv.Itoa(update.Message.From.ID)) {
			log.Printf("The id %d is not in the whitelist", update.Message.From.ID)
			continue
		}

		if update.Message.Document != nil {
			log.Println("Received a torrent file")
			b.HandleTorrent(updates, update, transmission)
		} else {
			log.Printf("Received the following command: %s\n", update.Message.Text)
			command := strings.Fields(update.Message.Text)[0]
			switch command {
			case "/torrent":
				b.HandleTorrentCommand(updates, update.Message.Chat.ID, transmission)
			case "/magnet":
				b.HandleMagnetLink(updates, update.Message.Chat.ID, transmission)
			case "/rss":
				b.HandleRSSAdition(updates, update.Message.Chat.ID)
			case "/help":
				b.HandleHelpCommand(update)
			default:
				log.Printf("unknown %s command\n", update.Message.Text)
				b.HandleDefault(update)
			}
		}
	}
}

// HandleTorrent handles the process once a torrent file has been uploaded
func (b *Bot) HandleTorrent(updates <-chan tgbotapi.Update, update tgbotapi.Update, transmission *transmission.Client) {
	// Get the torrent from the message
	file, err := b.BotAPI.GetFile(tgbotapi.FileConfig{FileID: update.Message.Document.FileID})
	if err != nil {
		log.Println("Error getting file link:", err)
		return
	}

	fileLink, err := getTorrent(b, update.Message.Document.FileID, file)
	if err != nil {
		log.Printf("Error getting torrent, aborting: %v", err)
		return
	}

	handleDownload(b, updates, update.Message.Chat.ID, transmission, fileLink)
}

// HandleTorrentCommand handles /torrent command which is ask for the torrent and then handle it like a direct upload
func (b *Bot) HandleTorrentCommand(updates <-chan tgbotapi.Update, chatID int64, transmission *transmission.Client) {
	requestMessage := "Please send the torrent file:"

	msg := tgbotapi.NewMessage(chatID, requestMessage)
	b.BotAPI.Send(msg)

	// Listen for the user's input for the magnet link
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Document != nil {
			b.HandleTorrent(updates, update, transmission)
		}
		break
	}
}

// HandleMagnetLink handles the /magnet command
func (b *Bot) HandleMagnetLink(updates <-chan tgbotapi.Update, chatID int64, transmission *transmission.Client) {
	var fileLink string

	requestMessage := "Please enter the magnet link:"

	msg := tgbotapi.NewMessage(chatID, requestMessage)
	b.BotAPI.Send(msg)

	// Listen for the user's input for the magnet link
	for update := range updates {
		if update.Message == nil {
			continue
		}

		fileLink = update.Message.Text
		break
	}

	handleDownload(b, updates, chatID, transmission, fileLink)
}

// handleDownload handles the common logic for getting the download path and starting the actual download
func handleDownload(b *Bot, updates <-chan tgbotapi.Update, chatID int64, transmission *transmission.Client, fileLink string) {
	var downloadPath string

	// Ask for the download path
	msg := tgbotapi.NewMessage(chatID, "Enter the download path:")
	b.BotAPI.Send(msg)

	// Listen for the user's input for the download path
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Extract the download path
		downloadPath = update.Message.Text
		break
	}

	// Start the download using the provided file/link and download path via the Transmission client
	torrentID, err := transmission.StartDownload(fileLink, downloadPath)
	if err != nil {
		log.Println("Error starting download:", err)
		return
	}

	// Notify the user that the download has started
	log.Println("Download started")
	startMsg := tgbotapi.NewMessage(chatID, "Download started!")
	b.BotAPI.Send(startMsg)

	// Poll download status every minute until it's completed
	go b.WaitForDownload(torrentID, chatID, transmission)
}

// getTorrent handles processing of torrent files and returns the path on disk of the torrent file
func getTorrent(b *Bot, fileID string, file tgbotapi.File) (string, error) {
	fileLink := fmt.Sprintf("torrents/%s.torrent", fileID)

	// Create folder if does not exist
	if _, err := os.Stat("torrents"); os.IsNotExist(err) {
		err := os.Mkdir("torrents", 0755)
		if err != nil {
			log.Panic("Cannot create folder for torrents:", err)
		}
	}

	// Download torrent file
	torrentURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.BotAPI.Token, file.FilePath)

	// Create the file
	out, err := os.Create(fileLink)
	if err != nil {
		return "", fmt.Errorf("error creating torrent file: %v", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(torrentURL)
	if err != nil {
		return "", fmt.Errorf("error getting file from http: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error writing file: %v", err)
	}

	return fileLink, nil
}

// WaitForDownload is designed to be launched as a subroutine and wait for the download and inform the user
func (b *Bot) WaitForDownload(torrentID, chatID int64, transmission *transmission.Client) error {
	for {
		time.Sleep(time.Minute)

		// Check if download is complete
		isComplete, err := transmission.IsDownloadComplete(torrentID)
		if err != nil {
			log.Println("Error checking download status:", err)
			return err
		}

		if isComplete {
			// If download is complete, get the torrent name
			name, err := transmission.GetName(torrentID)
			if err != nil {
				log.Println("Error retrieving torrent info:", err)
				return err
			}

			// Notify the user about the completed download with the torrent name
			completedMsg := tgbotapi.NewMessage(chatID, "Download completed: "+name)
			b.BotAPI.Send(completedMsg)
			break // Exit the loop when download is complete
		}
	}
	return nil
}

// HandleRSSAdition handles /rss command, adding the new feed and restarting the docker
func (b *Bot) HandleRSSAdition(updates <-chan tgbotapi.Update, chatID int64) {
	// Initialize the variables for file/link and download path
	var rssUrl, downloadPath string

	// Ask for the rss url
	msg := tgbotapi.NewMessage(chatID, "Enter the RSS url:")
	b.BotAPI.Send(msg)

	// Listen for the user's input for the download path
	for update := range updates {
		if update.Message == nil {
			continue
		}
		// Extract the download path
		rssUrl = update.Message.Text
		break
	}

	// Ask for the download path
	msg = tgbotapi.NewMessage(chatID, "Enter the download path:")
	b.BotAPI.Send(msg)

	// Listen for the user's input for the download path
	for update := range updates {
		if update.Message == nil {
			continue
		}
		// Extract the download path
		downloadPath = update.Message.Text
		break
	}

	log.Printf("Adding feed to yaml...")
	// Add the new feed to the config file
	if err := yamlhandler.AddFeedToYAML(rssUrl, downloadPath); err != nil {
		log.Println("Error adding feed to yaml: ", err)
		return
	}
	log.Printf("Done.\n")

	// Restart the docker so it starts watching the new feed
	log.Printf("Restarting Transmission-rss Docker...")
	if err := dockerhandler.RestartContainer("transmission-rss"); err != nil {
		log.Println("error restarting rss docker: ", err)
		return
	}
	log.Printf("Done.\n")

	// Tell the user the new feed has been created
	msg = tgbotapi.NewMessage(chatID, "Feed created!")
	b.BotAPI.Send(msg)
}

// HandleHelpCommand handles the /help command
func (b *Bot) HandleHelpCommand(update tgbotapi.Update) {
	// Get the chat ID
	chatID := update.Message.Chat.ID

	// Create the help message with available commands
	helpMessage := "Available commands:\n" +
		"/torrent - Upload a torrent file\n" +
		"/magnet - Input a magnet link\n" +
		"/rss - Input a rss feed into transmission-rss" +
		"/help - Show available commands"

	// Send the help message to the user
	msg := tgbotapi.NewMessage(chatID, helpMessage)
	_, err := b.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error sending help message:", err)
	}
}

// HandleDefault handles any unrecognized command or input
func (b *Bot) HandleDefault(update tgbotapi.Update) {
	// Get the chat ID
	chatID := update.Message.Chat.ID

	// Send an error message for unrecognized commands
	errorMessage := "Sorry, I don't recognize that command. Please use /help to see available commands."
	msg := tgbotapi.NewMessage(chatID, errorMessage)
	_, err := b.BotAPI.Send(msg)
	if err != nil {
		log.Println("Error sending default message:", err)
	}
}
func (b *Bot) CheckWhiteList(id string) bool {
	if slices.Contains(b.Config.Telegram.Whitelist, id) {
		return true

	} else {
		return false
	}
}
