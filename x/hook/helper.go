package hook

import "strings"

// getPackageFile gets package/file.go style return
func getPackageFile(s string) string {
	fileIndex := strings.LastIndex(s, "/")
	packageIndex := strings.LastIndex(s[:fileIndex], "/")
	atIndex := strings.LastIndex(s[packageIndex+1:fileIndex], "@")
	if atIndex == -1 {
		return s[packageIndex+1:]
	}
	return s[packageIndex+1:packageIndex+atIndex+1] + "" + s[fileIndex:]
}
