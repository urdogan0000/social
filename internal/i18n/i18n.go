package i18n

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle
var localizers map[string]*i18n.Localizer

func Init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	_, err := bundle.LoadMessageFile("locales/en.json")
	if err != nil {
		panic(err)
	}
	_, err = bundle.LoadMessageFile("locales/tr.json")
	if err != nil {
		panic(err)
	}

	localizers = make(map[string]*i18n.Localizer)
	localizers["en"] = i18n.NewLocalizer(bundle, "en")
	localizers["tr"] = i18n.NewLocalizer(bundle, "tr")
}

func GetLocalizer(r *http.Request) *i18n.Localizer {
	locale := GetLocale(r)
	if localizer, ok := localizers[locale]; ok {
		return localizer
	}
	return localizers["en"]
}

func GetLocale(r *http.Request) string {
	if locale := r.URL.Query().Get("lang"); locale != "" {
		if locale == "tr" || locale == "en" {
			return locale
		}
	}

	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if lang == "tr" || lang == "en" {
			return lang
		}
	}

	return "en"
}

func parseAcceptLanguage(acceptLang string) string {
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(strings.Split(parts[0], ";")[0])
		if strings.HasPrefix(lang, "tr") {
			return "tr"
		}
		if strings.HasPrefix(lang, "en") {
			return "en"
		}
	}
	return "en"
}

func T(r *http.Request, messageID string, data ...interface{}) string {
	localizer := GetLocalizer(r)
	
	var templateData interface{}
	if len(data) > 0 {
		templateData = data[0]
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		return messageID
	}
	return msg
}

