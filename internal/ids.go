// Package internal holds helpers for motoig.
package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateAndroidDeviceID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("android-%x", hash[:8])
}

func GenerateMutationToken() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(3199988888888888888))
	return fmt.Sprintf("%d", n.Int64()+6800011111111111111)
}

func GenerateJazoest(symbols string) string {
	sum := 0
	for _, c := range symbols {
		sum += int(c)
	}
	return fmt.Sprintf("2%d", sum)
}

func GenerateSignature(data string) string {
	return fmt.Sprintf("signed_body=SIGNATURE.%s", url.QueryEscape(data))
}

func GenToken(size int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func PrefixURL(rawURL, host string) string {
	if len(rawURL) > 0 && rawURL[0] == '/' {
		return "https://" + host + rawURL
	}
	return rawURL
}

func FormatUserAgent(appVer, androidVer, androidRelease, dpi, resolution, manufacturer, model, cpu, locale, versionCode string) string {
	return fmt.Sprintf(UserAgentBase,
		appVer, androidVer, androidRelease, dpi, resolution,
		manufacturer, model, cpu, locale, versionCode,
	)
}

var UserAgentBase = "Instagram %s Android (%s/%s; %s; %s; %s; %s; %s; %s; %s)"

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func JoinMapValues(m map[string]string, sep string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(m[k])
	}
	return b.String()
}
