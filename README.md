# Telegram Bot for Transmission Daemon

This Telegram bot was designed to interact with a Transmission daemon for initiating downloads using either torrent files or magnet links.

I reforked the original repository because the creator added unnecessary stuff that didn’t align with the bot’s original goal. goal. I will add new features that are related to the original purpose of the bot.
## Features

- Accepted commands:
  - `/torrent`: Upload a torrent file
  - `/magnet`: Input a magnet link
  - `/rss`: Adds a new feed to transmission-rss
  - `/help`: Show available commands

## Installation and Setup

1. **Clone the Repository:**

        ```bash
        git clone https://github.com/r0zark/transmission-telegram-bot.git
        ```

2. **Install Dependencies:**

        ```bash
        cd yourbot
        go mod tidy
        ```

3. **Configuration:**

        Create a `config.yaml` file in the `config` directory and fill in the required details (see [Configuring the Bot](#configuring-the-bot)).

4. **Build and Run:**

        ```bash
        go build -o transmission-telegram-bot main.go
        ./transmission-telegram-bot
        ```

## Configuring the Bot

Fill in the required details in the `config.yaml` file:

```yaml
transmission:
    url: "transmission_server_ip_address"
    port: "TRANSMISSION_PORT" #Defaults to 9091
    https: "BOOLEAN_FOR_HTTPS" #Defaults to false
    user: "YOUR_TRANSMISSION_USERNAME"
    password: "YOUR_TRANSMISSION_PASSWORD"
telegram:
    botToken: "YOUR_TELEGRAM_BOT_TOKEN"
    chatID: "YOUR_TELEGRAM_CHATID"
device:
    deviceSn: "YOUR_DEVICE_SN"
```

## Usage

- Start the bot by running the executable (`transmission-telegram-bot`).
- Interact with the bot via Telegram using the commands mentioned above.

### Docker-Compose (working on the docker image)

## Contributors
- [Samuel Ruiz](https://github.com/r0zark)
- [Manuel Mendoza](https://github.com/Coolknight) (Original creator of the bot)

## License

This project is licensed under the [MIT License](LICENSE).

## Contact or Support

For any inquiries or support, please create a new issue.
