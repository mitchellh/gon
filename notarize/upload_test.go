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
	childCommands["upload-success"] = testCmdUploadSuccess
	childCommands["upload-errors"] = testCmdUploadErrors
	childCommands["upload-exit-status"] = testCmdUploadExitStatus
}

func TestUpload_success(t *testing.T) {
	uuid, err := upload(context.Background(), &Options{
		Logger:  hclog.L(),
		BaseCmd: childCmd(t, "upload-success"),
	})

	require.NoError(t, err)
	require.Equal(t, uuid, "edc8e846-d6ce-444d-9eef-499aa444da1c")
}

func TestUpload_errors(t *testing.T) {
	uuid, err := upload(context.Background(), &Options{
		Logger:  hclog.L(),
		BaseCmd: childCmd(t, "upload-errors"),
	})

	require.Error(t, err)
	require.Empty(t, uuid)
}

func TestUpload_exitStatus(t *testing.T) {
	uuid, err := upload(context.Background(), &Options{
		Logger:  hclog.L(),
		BaseCmd: childCmd(t, "upload-exit-status"),
	})

	require.Error(t, err)
	require.Empty(t, uuid)
}

// testCmdUploadSuccess mimicks a successful submission.
func testCmdUploadSuccess() int {
	fmt.Println(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>notarization-upload</key>
	<dict>
		<key>RequestUUID</key>
		<string>edc8e846-d6ce-444d-9eef-499aa444da1c</string>
	</dict>
	<key>os-version</key>
	<string>10.15.1</string>
	<key>success-message</key>
	<string>No errors uploading './terraform.zip'.</string>
	<key>tool-path</key>
	<string>/Applications/Xcode.app/Contents/SharedFrameworks/ContentDeliveryServices.framework/Versions/A/Frameworks/AppStoreService.framework</string>
	<key>tool-version</key>
	<string>4.00.1181</string>
</dict>
</plist>
`))
	return 0
}

// testCmdUploadErrors mimicks a successful submission.
func testCmdUploadErrors() int {
	fmt.Println(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>os-version</key>
	<string>10.15.1</string>
	<key>product-errors</key>
	<array>
		<dict>
			<key>code</key>
			<integer>-18000</integer>
			<key>message</key>
			<string>ERROR ITMS-90732: "The software asset has already been uploaded. The upload ID is 671a0eb9-01b0-4966-b485-78b9f370abff" at SoftwareAssets/EnigmaSoftwareAsset</string>
			<key>userInfo</key>
			<dict>
				<key>NSLocalizedDescription</key>
				<string>ERROR ITMS-90732: "The software asset has already been uploaded. The upload ID is 671a0eb9-01b0-4966-b485-78b9f370abff" at SoftwareAssets/EnigmaSoftwareAsset</string>
				<key>NSLocalizedFailureReason</key>
				<string>ERROR ITMS-90732: "The software asset has already been uploaded. The upload ID is 671a0eb9-01b0-4966-b485-78b9f370abff" at SoftwareAssets/EnigmaSoftwareAsset</string>
				<key>NSLocalizedRecoverySuggestion</key>
				<string>ERROR ITMS-90732: "The software asset has already been uploaded. The upload ID is 671a0eb9-01b0-4966-b485-78b9f370abff" at SoftwareAssets/EnigmaSoftwareAsset</string>
			</dict>
		</dict>
	</array>
	<key>tool-path</key>
	<string>/Applications/Xcode.app/Contents/SharedFrameworks/ContentDeliveryServices.framework/Versions/A/Frameworks/AppStoreService.framework</string>
	<key>tool-version</key>
	<string>4.00.1181</string>
</dict>
</plist>
`))

	// Despite an error we return exit code 0 so we can test that we error
	// in the presence of errors in the output.
	return 0
}

// testCmdUploadExitStatus
func testCmdUploadExitStatus() int {
	return 1
}
