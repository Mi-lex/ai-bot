package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var componentInteractionsEventStopResponse = "stop_response"

var componentsInteractionHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	componentInteractionsEventStopResponse: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Print("User clicked 'Stop' =(")

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
