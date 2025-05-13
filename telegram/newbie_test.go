package telegram

import (
	"testing"

	"github.com/forPelevin/gomoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"
)

func TestIsBanNewbieForEntities(t *testing.T) {
	hdl := &InstanceObj{}

	testCases := []struct {
		name           string
		ignore         map[string]struct{}
		text           string
		entityType     string
		url            string
		ban            bool
		noEntities     bool
		offset, length int
	}{
		{
			name:       "text_link",
			ignore:     nil,
			text:       "123",
			entityType: "text_link",
			url:        "https://github.com/",
			ban:        true,
		},
		{
			name:       "no urls",
			ignore:     nil,
			ban:        false,
			noEntities: true,
		},
		{
			name:       "ignore text_link",
			ignore:     map[string]struct{}{"github.com": {}},
			text:       "123",
			entityType: "text_link",
			url:        "https://github.com/",
			ban:        false,
		},
		{
			name:       "ignore url",
			ignore:     map[string]struct{}{"github.com": {}},
			text:       "🤦‍♂️ https://github.com/ 🤦‍♂️",
			entityType: "url",
			ban:        false,
			offset:     6,
			length:     19,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var entities []tgbotapi.MessageEntity

			if !testCase.noEntities {
				entity := tgbotapi.MessageEntity{
					Type:   testCase.entityType,
					URL:    testCase.url,
					Offset: testCase.offset,
					Length: testCase.length,
				}

				entities = []tgbotapi.MessageEntity{entity}
			}

			urlsList := hdl.getURLsFromEntities(testCase.text, entities...)

			ban := hdl.isBanNewbieForURLs(testCase.ignore, urlsList)
			if testCase.ban {
				require.True(t, ban, testCase.name)
			} else {
				require.False(t, ban, testCase.name)
			}
		})
	}
}

func TestEmojii(t *testing.T) {
	require.Equal(t, 3, len(gomoji.CollectAll("🤦‍♂️ 1 🤦‍♂️ 2 🤦‍♂️ 3")))
}

func TestLen(t *testing.T) {
	res := string([]rune("D32KAE_31_P04NO_PQX088B0T")[0:6])
	require.Equal(t, "D32KAE", res)
}
