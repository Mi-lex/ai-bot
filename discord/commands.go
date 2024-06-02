package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const STOP_RESPONSE_EVENT = "stop_response"

var chatResponseComponents = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: STOP_RESPONSE_EVENT,
				Label:    "Stop responding",
			},
		},
	},
}

var componentsInteractionHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	STOP_RESPONSE_EVENT: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

const SET_MODEL_COMMAND = "set_model"

func createSetModelCommand(models []string) *discordgo.ApplicationCommand {
	var choices []*discordgo.ApplicationCommandOptionChoice

	for _, modelName := range models {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  modelName,
			Value: modelName,
		})
	}

	return &discordgo.ApplicationCommand{
		Name:        SET_MODEL_COMMAND,
		Description: "Command to set the model for following usage",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "string-option",
				Description: "String option",
				Required:    true,
				Choices:     choices,
			},
		},
	}
}

var applicationCommandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	/**
	 * TODO! add actual handler
	 * saving model into the store
	 **/
	SET_MODEL_COMMAND: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		modelOption := i.ApplicationCommandData().Options[0]

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			// Ignore type for now, they will be discussed in "responses"
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf(
					modelOption.StringValue(),
				),
			},
		})
	},
}
