package transmission

import (
	"fmt"
	"os"
	"strings"

	"github.com/R0zark/transmission-telegram-bot/config"
	"github.com/hekmon/transmissionrpc"
)

// Client struct holds the Transmission client
type Client struct {
	Client *transmissionrpc.Client
}

// NewClient initializes a new Transmission client
func NewClient(config config.Transmission) (*Client, error) {
	advconf := transmissionrpc.AdvancedConfig{
		Port:  config.Port,
		HTTPS: config.HTTPS,
	}
	client, err := transmissionrpc.New(config.URL, config.User,
		config.Password, &advconf)
	if err != nil {
		return nil, err
	}

	return &Client{Client: client}, nil
}

// StartDownload starts a download using the Transmission client
func (c *Client) StartDownload(magnetLink, downloadPath string) (int64, error) {

	var torrentId *int64

	if c.ChecksMagnetURL(magnetLink) {
		response, err := c.Client.TorrentAdd(&transmissionrpc.TorrentAddPayload{Filename: &magnetLink})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 0, err
		} else {
			// Only 3 fields will be returned/set in the Torrent struct
			fmt.Println("Added: " + *response.Name)
			torrentId = response.ID
			return *torrentId, nil
		}

	} else {
		response, err := c.Client.TorrentAddFileDownloadDir(magnetLink, downloadPath)
		if err != nil {
			return 0, err
		}
		fmt.Println("Added: " + *response.Name)
		torrentId = response.ID // Assuming only one torrent is added
		return *torrentId, nil
	}
}

// IsDownloadComplete checks if the specified torrent download is complete
func (c *Client) IsDownloadComplete(torrentID int64) (bool, error) {
	torrents, err := c.Client.TorrentGet([]string{"id", "percentDone"}, []int64{torrentID})
	if err != nil {
		return false, err
	}

	// Assuming only one torrent is retrieved
	if len(torrents) == 1 {
		percentDone := torrents[0].PercentDone
		if *percentDone == 1.0 {
			return true, nil // Download is complete
		}
		return false, nil // Download is still in progress
	}

	return false, nil // Handle the case when no torrents or multiple torrents are returned
}

// GetName returns the name of the specified torrent
func (c *Client) GetName(torrentID int64) (string, error) {
	torrents, err := c.Client.TorrentGet([]string{"id", "name"}, []int64{torrentID})
	if err != nil {
		return "", err
	}

	name := torrents[0].Name

	return *name, nil // Handle the case when no torrents or multiple torrents are returned
}

// ChecksMagnetURL checks if the specified torrent URL is a magnet
func (c *Client) ChecksMagnetURL(torrentURL string) bool {

	if strings.Split(torrentURL, ":")[0] == "magnet" {
		return true
	} else {
		return false
	}
}
