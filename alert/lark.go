package alert

import (
	"fmt"
	logger "github.com/owlto-dao/utils-go/log"
	"github.com/owlto-dao/utils-go/network"
	"log"
	"strings"
)

// alerter := alert.NewLarkAlerter("https://open.larksuite.com/open-apis/bot/v2/hook/xxxx")

var LarkAlertBot *LarkAlerter

type LarkAlerter struct {
	*CommonAlerter
	webhook string
}

func NewLarkAlerter(webhook string) *LarkAlerter {
	alerter := &LarkAlerter{
		webhook:       strings.TrimSpace(webhook),
		CommonAlerter: NewCommonAlerter(120, 900),
	}
	LarkAlertBot = alerter
	return alerter
}

func (la *LarkAlerter) AlertTextLazyGroup(group string, msg string, err error) {
	la.DoAlertTextLazy(la, group, msg, err)
}

func (la *LarkAlerter) AlertTextLazy(msg string, err error) {
	la.DoAlertTextLazy(la, "", msg, err)
}

func (la *LarkAlerter) AlertText(msg string, err error) {
	data := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": fmt.Sprintf("%s : %v", msg, err),
		},
	}
	network.Request(la.webhook, data, nil)
	log.Println(msg, ":", err)
}

func (la *LarkAlerter) AlertMessage(msg string) {
	data := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": msg,
		},
	}
	network.Request(la.webhook, data, nil)
	logger.Infof(msg)
}
