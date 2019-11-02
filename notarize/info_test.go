package notarize

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func init() {
	childCommands["info-success"] = testCmdInfoSuccess
}

func TestInfo_success(t *testing.T) {
	info, err := info(context.Background(), "foo", &Options{
		Logger:  hclog.L(),
		BaseCmd: childCmd(t, "info-success"),
	})

	require := require.New(t)
	require.NoError(err)
	require.Equal(info.RequestUUID, "edc8e846-d6ce-444d-9eef-499aa444da1c")
	require.Equal(info.Hash, "644d0af906ae26c87037cd6e9073382d5b0461b39e7f23c7bb69a35debacedd4")
	require.Equal(info.LogFileURL, "https://osxapps-ssl.itunes.apple.com/itunes-assets/Enigma123/v4/29/f2/81/29f28128-e2be-158a-f421-1e19692dd935/developer_log.json?accessKey=1572864491_3132212434837665280_4XLMw7lZxMfKdHhgnlPkueVue9woI2MjQ6VEc8R0cxJrL9GGcTQSiE0C9Cu5o6o%2B3JtYGSqGWdvc3mJHbS0NBRZkHT%2BbwbdMGPT8poYk7TTkfHUIcW5aBz0aFO7RB6mSWVuZWOFT0dZ4VS%2Bep2LUP2KTDtDwiGQbTULu9VgZ1oY%3D")
	require.Equal(info.Status, "success")
	require.Equal(info.StatusMessage, "Package Approved")
}

// testCmdInfoSuccess mimicks a successful submission.
func testCmdInfoSuccess() int {
	fmt.Println(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>notarization-info</key>
	<dict>
		<key>Date</key>
		<date>2019-11-02T02:17:12Z</date>
		<key>Hash</key>
		<string>644d0af906ae26c87037cd6e9073382d5b0461b39e7f23c7bb69a35debacedd4</string>
		<key>LogFileURL</key>
		<string>https://osxapps-ssl.itunes.apple.com/itunes-assets/Enigma123/v4/29/f2/81/29f28128-e2be-158a-f421-1e19692dd935/developer_log.json?accessKey=1572864491_3132212434837665280_4XLMw7lZxMfKdHhgnlPkueVue9woI2MjQ6VEc8R0cxJrL9GGcTQSiE0C9Cu5o6o%2B3JtYGSqGWdvc3mJHbS0NBRZkHT%2BbwbdMGPT8poYk7TTkfHUIcW5aBz0aFO7RB6mSWVuZWOFT0dZ4VS%2Bep2LUP2KTDtDwiGQbTULu9VgZ1oY%3D</string>
		<key>RequestUUID</key>
		<string>edc8e846-d6ce-444d-9eef-499aa444da1c</string>
		<key>Status</key>
		<string>success</string>
		<key>Status Code</key>
		<integer>0</integer>
		<key>Status Message</key>
		<string>Package Approved</string>
	</dict>
	<key>os-version</key>
	<string>10.15.1</string>
	<key>success-message</key>
	<string>No errors getting notarization info.</string>
	<key>tool-path</key>
	<string>/Applications/Xcode.app/Contents/SharedFrameworks/ContentDeliveryServices.framework/Versions/A/Frameworks/AppStoreService.framework</string>
	<key>tool-version</key>
	<string>4.00.1181</string>
</dict>
</plist>
`))
	return 0
}
