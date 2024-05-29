# DGPT Bot

DGPT Bot is a Discord bot built in Go. It uses OpenAI for chat functionality and Redis for data storage.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go 1.18
- Redis | Free cloud DB can be easily created here - https://app.redislabs.com/
- Discord App and Created Bot | https://discord.com/developers/docs/quick-start/getting-started
- Open AI account

### Installation

1. Clone the repository:
```sh
git clone https://github.com/Mi-lex/dgpt-bot.git
```
2. Install the dependencies:
```sh
go mod download
```
3. Set up environment variables. You can find the example in [.env.example](./.env.example) file.

## Usage

To run the bot, use the following command:
```sh
go run application.go
```

## Deployment

This project includes a GitHub Actions workflow for deployment to AWS. See deploy.yml for more details.

## Built With

* [Go](https://go.dev/)
* [OpenAI](https://github.com/sashabaranov/go-openai)
* [DiscordGo](https://github.com/bwmarrin/discordgo)