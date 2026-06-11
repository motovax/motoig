// Package config holds Instagram device settings and app version constants.
package config

const (
	APIDomain       = "i.instagram.com"
	DefaultAppVer   = "428.0.0.47.67"
	DefaultVersionCode = "961145276"
	DefaultBloksVersioningID = "7189b949425f9bf80ea8bd880cf5a3080b292d9b1c4b38a18d112f7c4b71e7a8"
)

var UserAgentBase = "Instagram %s Android (%s/%s; %s; %s; %s; %s; %s; %s; %s; %s)"

var DeviceSettings = map[string]string{
	"android_version":  "34",
	"android_release":  "14",
	"dpi":              "480dpi",
	"resolution":       "1344x2992",
	"manufacturer":     "Google/google",
	"device":           "husky",
	"model":            "Pixel 8 Pro",
	"cpu":              "husky",
}

type AppVersion struct {
	AppVersion       string
	VersionCode      string
	BloksVersioningID string
}

var AppSettings = map[string]AppVersion{
	DefaultAppVer: {
		AppVersion:        DefaultAppVer,
		VersionCode:       DefaultVersionCode,
		BloksVersioningID: DefaultBloksVersioningID,
	},
	"364.0.0.35.86": {
		AppVersion:        "364.0.0.35.86",
		VersionCode:       "374010953",
		BloksVersioningID: "8ccf54aad76788a6ca03ddfc33afcdcf692f2f5a3ba814ea73d5facba7fa2c2d",
	},
	"385.0.0.47.74": {
		AppVersion:        "385.0.0.47.74",
		VersionCode:       "378906843",
		BloksVersioningID: "a8973d49a9cc6a6f65a4997c10216ce2a06f65a517010e64885e92029bb19221",
	},
}

var SupportedCapabilities = []map[string]string{
	{"value": "119.0,120.0,121.0,122.0,123.0,124.0,125.0,126.0,127.0,128.0,129.0,130.0,131.0,132.0,133.0,134.0,135.0,136.0,137.0,138.0,139.0,140.0,141.0,142.0", "name": "SUPPORTED_SDK_VERSIONS"},
	{"value": "14", "name": "FACE_TRACKER_VERSION"},
	{"value": "ETC2_COMPRESSION", "name": "COMPRESSION"},
	{"value": "gyroscope_enabled", "name": "gyroscope"},
}
