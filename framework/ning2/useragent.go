package ning2

import "strings"

// UserAgent 用户代理信息
type UserAgent struct {
	String  string
	URL     string
	Device  string
	Name    string
	Version string
	OS      string
	OSVersion string
	Bot     bool
	Mobile  bool
	Tablet  bool
	Desktop bool
}

// Parse 解析 User-Agent 字符串
func Parse(ua string) *UserAgent {
	u := &UserAgent{
		String: ua,
	}

	ua = strings.ToLower(ua)

	// 检测设备类型 - iOS 检测优先
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		if strings.Contains(ua, "iphone") {
			u.Mobile = true
			u.Desktop = false
		} else {
			u.Tablet = true
			u.Desktop = false
		}
	} else if strings.Contains(ua, "mobile") || (strings.Contains(ua, "android") && !strings.Contains(ua, "tablet")) {
		u.Mobile = true
		u.Desktop = false
	} else if strings.Contains(ua, "tablet") {
		u.Tablet = true
		u.Desktop = false
	} else {
		u.Desktop = true
	}

	// 检测机器人 - 更精确的检测
	if strings.Contains(ua, "bot") || strings.Contains(ua, "crawler") || 
	   strings.Contains(ua, "spider") || strings.Contains(ua, "curl/") || 
	   strings.Contains(ua, "wget/") || strings.Contains(ua, "python") {
		u.Bot = true
	}

	// 检测浏览器
	switch {
	case strings.Contains(ua, "edg/"):
		u.Name = "Edge"
		u.Version = extractVersion(ua, "edg/")
	case strings.Contains(ua, "chrome/"):
		u.Name = "Chrome"
		u.Version = extractVersion(ua, "chrome/")
	case strings.Contains(ua, "firefox/"):
		u.Name = "Firefox"
		u.Version = extractVersion(ua, "firefox/")
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome"):
		u.Name = "Safari"
		u.Version = extractVersion(ua, "version/")
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr/"):
		u.Name = "Opera"
		u.Version = or(extractVersion(ua, "version/"), extractVersion(ua, "opr/"))
	default:
		u.Name = "Unknown"
	}

	// 检测操作系统 - iOS 优先检测
	switch {
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ios"):
		u.OS = "iOS"
		u.OSVersion = extractVersion(ua, "os ")
	case strings.Contains(ua, "windows"):
		u.OS = "Windows"
		u.OSVersion = extractVersion(ua, "windows nt ")
	case strings.Contains(ua, "mac os x"):
		u.OS = "macOS"
		u.OSVersion = extractVersion(ua, "mac os x ")
	case strings.Contains(ua, "android"):
		u.OS = "Android"
		u.OSVersion = extractVersion(ua, "android ")
	case strings.Contains(ua, "linux"):
		u.OS = "Linux"
	default:
		u.OS = "Unknown"
	}

	// 检测设备
	if strings.Contains(ua, "iphone") {
		u.Device = "iPhone"
	} else if strings.Contains(ua, "ipad") {
		u.Device = "iPad"
	} else if strings.Contains(ua, "android") {
		u.Device = "Android"
	} else if strings.Contains(ua, "windows phone") {
		u.Device = "Windows Phone"
	} else {
		u.Device = "Desktop"
	}

	return u
}

func extractVersion(ua, name string) string {
	idx := strings.Index(ua, name)
	if idx < 0 {
		return ""
	}
	start := idx + len(name)
	end := start
	for end < len(ua) && (ua[end] >= '0' && ua[end] <= '9' || ua[end] == '.') {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return ""
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}