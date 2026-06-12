package instagrapi

import "strings"

const (
	// Repo is the upstream instagrapi GitHub repository.
	Repo = "https://github.com/subzeroid/instagrapi"
	// Branch is the instagrapi branch motoig tracks.
	Branch = "master"
)

// Source returns a stable GitHub URL for an instagrapi file path.
func Source(path string) string {
	path = strings.TrimPrefix(path, "/")
	return Repo + "/blob/" + Branch + "/" + path
}

// Symbol maps a motoig API name to the instagrapi source file that defines it.
var Symbol = map[string]string{
	"Client":                         "instagrapi/__init__.py",
	"Client.Login":                   "instagrapi/mixins/auth.py",
	"Client.LoginBySessionID":        "instagrapi/mixins/auth.py",
	"Client.SetSessionID":            "instagrapi/mixins/auth.py",
	"Client.DirectThreads":           "instagrapi/mixins/direct.py",
	"Client.DirectMessages":          "instagrapi/mixins/direct.py",
	"Client.DirectSend":              "instagrapi/mixins/direct.py",
	"Client.RealtimeConnect":         "instagrapi/mixins/realtime.py",
	"Client.RealtimeDirectSubscribe": "instagrapi/realtime/client.py",
	"Client.RealtimeOn":            "instagrapi/mixins/realtime.py",
	"RealtimeClient.dispatchMessageSync": "instagrapi/realtime/client.py",
	"RealtimeClient.buildConnection":     "instagrapi/realtime/client.py",
	"State.PrivateRequest":           "instagrapi/mixins/private.py",
	"extractors.ExtractDirectThread": "instagrapi/extractors.py",
	"extractors.ExtractDirectMessage": "instagrapi/extractors.py",
}

// Ref returns the GitHub URL for a motoig symbol's instagrapi source file.
func Ref(symbol string) string {
	if path, ok := Symbol[symbol]; ok {
		return Source(path)
	}
	return Repo
}