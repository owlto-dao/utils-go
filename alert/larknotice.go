package alert

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-lark/lark"
	"github.com/owlto-dao/utils-go/convert"
)

var LarkBot *Bot

type Bot struct {
	*lark.Bot
}

func NewLarkBot(appID, appSecret string) (*Bot, error) {
	bot := lark.NewChatBot(appID, appSecret)
	err := bot.StartHeartbeat()
	if err != nil {
		return nil, err
	}
	larkBot := &Bot{
		Bot: bot,
	}
	LarkBot = larkBot
	return larkBot, nil
}

func (b *Bot) SendMessageCardByID(ctx context.Context, cardID string, chatIDOrOpenID string, template map[string]interface{}) error {
	var format = `{"type":"template","data":{"template_id":"%s","template_variable":%s}}`
	msg := lark.NewMsgBuffer(lark.MsgInteractive)
	if strings.HasPrefix(chatIDOrOpenID, "oc_") {
		om := msg.BindChatID(chatIDOrOpenID).Card(fmt.Sprintf(format, cardID, convert.ConvertToJsonString(template))).Build()
		resp, err := b.Bot.PostMessage(om)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("send message to lark fail, code: %d, msg: %v", resp.Code, resp.Msg)
		}
	} else if strings.HasPrefix(chatIDOrOpenID, "ou_") {
		om := msg.BindOpenID(chatIDOrOpenID).Card(fmt.Sprintf(format, cardID, convert.ConvertToJsonString(template))).Build()
		resp, err := b.Bot.PostMessage(om)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("send message to lark fail, code: %d, msg: %v", resp.Code, resp.Msg)
		}
	} else {
		return fmt.Errorf("unknown chatIDOrOpenID: %s", chatIDOrOpenID)
	}
	return nil
}
