package discord

import (
	"github.com/bwmarrin/discordgo"
)

const componentInteractionsEventStopResponse = "stop_response"

var componentsInteractionHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	componentInteractionsEventStopResponse: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		stopResponse, exist := currentChatResponses[i.ChannelID]

		if exist {
			stopResponse()
		}

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    i.Message.Content,
				Components: []discordgo.MessageComponent{},
			},
		})

		if err != nil {
			panic(err)
		}
	},
}
