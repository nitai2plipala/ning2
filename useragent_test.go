package ning2

import (
	"testing"
)

// ==================== UserAgent 基础测试 ====================

func TestParse_Basic(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36"
	u := Parse(ua)

	if u == nil {
		t.Fatal("Parse returned nil")
	}
	if u.String != ua {
		t.Error("String field not set correctly")
	}
}

func TestParse_Empty(t *testing.T) {
	u := Parse("")
	if u == nil {
		t.Fatal("Parse returned nil for empty string")
	}
	if u.String != "" {
		t.Error("String field should be empty")
	}
}

// ==================== 浏览器检测测试 ====================

func TestParse_Chrome(t *testing.T) {
	tests := []struct {
		ua      string
		wantName string
	}{
		{"Mozilla/5.0 Chrome/120.0.0.0", "Chrome"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", "Chrome"},
		{"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", "Chrome"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.Name != tt.wantName {
			t.Errorf("UA %q: want name %s, got %s", tt.ua, tt.wantName, u.Name)
		}
	}
}

func TestParse_Firefox(t *testing.T) {
	tests := []struct {
		ua      string
		wantName string
	}{
		{"Mozilla/5.0 Firefox/120.0", "Firefox"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0", "Firefox"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0", "Firefox"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.Name != tt.wantName {
			t.Errorf("UA %q: want name %s, got %s", tt.ua, tt.wantName, u.Name)
		}
	}
}

func TestParse_Safari(t *testing.T) {
	tests := []struct {
		ua      string
		wantName string
	}{
		{"Mozilla/5.0 Safari/605.1.15", "Safari"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15", "Safari"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1", "Safari"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.Name != tt.wantName {
			t.Errorf("UA %q: want name %s, got %s", tt.ua, tt.wantName, u.Name)
		}
	}
}

func TestParse_Edge(t *testing.T) {
	tests := []struct {
		ua      string
		wantName string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0", "Edge"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0", "Edge"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.Name != tt.wantName {
			t.Errorf("UA %q: want name %s, got %s", tt.ua, tt.wantName, u.Name)
		}
	}
}

func TestParse_Opera(t *testing.T) {
	tests := []struct {
		ua      string
		wantName string
	}{
		{"Opera/9.80", "Opera"},
		// Opera 在现代浏览器中会包含 Chrome，需要先检测 OPR
		{"OPR/120.0.0.0", "Opera"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.Name != tt.wantName {
			t.Errorf("UA %q: want name %s, got %s", tt.ua, tt.wantName, u.Name)
		}
	}
}

func TestParse_Unknown(t *testing.T) {
	tests := []string{
		"Unknown",
		"",
		"Mozilla/4.0",
		"CustomClient/1.0",
	}

	for _, ua := range tests {
		u := Parse(ua)
		if u.Name != "Unknown" {
			t.Errorf("UA %q: want name Unknown, got %s", ua, u.Name)
		}
	}
}

// ==================== 操作系统检测测试 ====================

func TestParse_OS_Windows(t *testing.T) {
	tests := []struct {
		ua    string
		wantOS string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0", "Windows"},
		{"Mozilla/5.0 (Windows NT 6.1; Win64; x64) Firefox/120.0", "Windows"},
		{"Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36", "Windows"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.OS != tt.wantOS {
			t.Errorf("UA %q: want OS %s, got %s", tt.ua, tt.wantOS, u.OS)
		}
	}
}

func TestParse_OS_MacOS(t *testing.T) {
	tests := []struct {
		ua    string
		wantOS string
	}{
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/120.0.0.0", "macOS"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0) Safari/605.1.15", "macOS"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0", "macOS"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.OS != tt.wantOS {
			t.Errorf("UA %q: want OS %s, got %s", tt.ua, tt.wantOS, u.OS)
		}
	}
}

func TestParse_OS_iOS(t *testing.T) {
	tests := []struct {
		ua    string
		wantOS string
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) Chrome/120.0.0.0", "iOS"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) Safari/604.1", "iOS"},
		{"Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) Safari/604.1", "iOS"},
		{"Mozilla/5.0 (iPod touch; CPU iPhone OS 16_0 like Mac OS X) Safari/604.1", "iOS"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.OS != tt.wantOS {
			t.Errorf("UA %q: want OS %s, got %s", tt.ua, tt.wantOS, u.OS)
		}
	}
}

func TestParse_OS_Android(t *testing.T) {
	tests := []struct {
		ua    string
		wantOS string
	}{
		{"Mozilla/5.0 (Linux; Android 13) Chrome/120.0.0.0", "Android"},
		{"Mozilla/5.0 (Linux; Android 12) Safari/537.36", "Android"},
		{"Mozilla/5.0 (Linux; Android 13; SM-S918B) AppleWebKit/537.36", "Android"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.OS != tt.wantOS {
			t.Errorf("UA %q: want OS %s, got %s", tt.ua, tt.wantOS, u.OS)
		}
	}
}

func TestParse_OS_Linux(t *testing.T) {
	tests := []struct {
		ua    string
		wantOS string
	}{
		{"Mozilla/5.0 (X11; Linux x86_64) Chrome/120.0.0.0", "Linux"},
		{"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) Firefox/120.0", "Linux"},
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36", "Linux"},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.OS != tt.wantOS {
			t.Errorf("UA %q: want OS %s, got %s", tt.ua, tt.wantOS, u.OS)
		}
	}
}

// ==================== 设备检测测试 ====================

func TestParse_Device_iPhone(t *testing.T) {
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) Chrome/120.0.0.0"
	u := Parse(ua)

	if u.Device != "iPhone" {
		t.Errorf("want device iPhone, got %s", u.Device)
	}
}

func TestParse_Device_iPad(t *testing.T) {
	ua := "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) Safari/604.1"
	u := Parse(ua)

	if u.Device != "iPad" {
		t.Errorf("want device iPad, got %s", u.Device)
	}
}

func TestParse_Device_Android(t *testing.T) {
	ua := "Mozilla/5.0 (Linux; Android 13) Chrome/120.0.0.0 Mobile"
	u := Parse(ua)

	if u.Device != "Android" {
		t.Errorf("want device Android, got %s", u.Device)
	}
}

func TestParse_Device_Desktop(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0"
	u := Parse(ua)

	if u.Device != "Desktop" {
		t.Errorf("want device Desktop, got %s", u.Device)
	}
}

// ==================== 设备类型测试 ====================

func TestParse_Mobile(t *testing.T) {
	tests := []struct {
		ua       string
		wantMob  bool
		wantDesk bool
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0) Chrome/120.0.0.0", true, false},
		{"Mozilla/5.0 (Linux; Android 13) Chrome/120.0.0.0 Mobile", true, false},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0", false, true},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1.15", false, true},
	}

	for _, tt := range tests {
		u := Parse(tt.ua)
		if u.Mobile != tt.wantMob {
			t.Errorf("UA %q: want mobile=%v, got %v", tt.ua, tt.wantMob, u.Mobile)
		}
		if u.Desktop != tt.wantDesk {
			t.Errorf("UA %q: want desktop=%v, got %v", tt.ua, tt.wantDesk, u.Desktop)
		}
	}
}

func TestParse_Tablet(t *testing.T) {
	ua := "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) Safari/604.1"
	u := Parse(ua)

	if !u.Tablet {
		t.Error("iPad should be detected as tablet")
	}
	if u.Desktop {
		t.Error("iPad should not be detected as desktop")
	}
}

// ==================== 机器人检测测试 ====================

func TestParse_Bot_Google(t *testing.T) {
	ua := "Googlebot/2.1 (+http://www.google.com/bot.html)"
	u := Parse(ua)

	if !u.Bot {
		t.Error("Googlebot should be detected as bot")
	}
}

func TestParse_Bot_Bing(t *testing.T) {
	ua := "Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)"
	u := Parse(ua)

	if !u.Bot {
		t.Error("Bingbot should be detected as bot")
	}
}

func TestParse_Bot_Yandex(t *testing.T) {
	ua := "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)"
	u := Parse(ua)

	if !u.Bot {
		t.Error("YandexBot should be detected as bot")
	}
}

func TestParse_Bot_Curl(t *testing.T) {
	ua := "curl/7.68.0"
	u := Parse(ua)

	if !u.Bot {
		t.Error("curl should be detected as bot")
	}
}

func TestParse_Bot_Wget(t *testing.T) {
	ua := "wget/1.21"
	u := Parse(ua)

	if !u.Bot {
		t.Error("wget should be detected as bot")
	}
}

func TestParse_Bot_Python(t *testing.T) {
	ua := "python-requests/2.28.0"
	u := Parse(ua)

	if !u.Bot {
		t.Error("python-requests should be detected as bot")
	}
}

func TestParse_NotBot(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0"
	u := Parse(ua)

	if u.Bot {
		t.Error("Normal browser should not be detected as bot")
	}
}

// ==================== 边界情况测试 ====================

func TestParse_LongUA(t *testing.T) {
	ua := string(make([]byte, 10000))
	for i := range ua {
		ua = ua[:i] + "a"
	}
	u := Parse(ua)
	if u == nil {
		t.Error("Parse should handle long UA string")
	}
}

func TestParse_SpecialChars(t *testing.T) {
	tests := []string{
		"()<>",
		"<script>alert(1)</script>",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	for _, ua := range tests {
		u := Parse(ua)
		if u == nil {
			t.Errorf("Parse failed for UA: %s", ua)
		}
	}
}

// ==================== 性能基准测试 ====================

func BenchmarkParse_Chrome(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(ua)
	}
}

func BenchmarkParse_iPhone(b *testing.B) {
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(ua)
	}
}

func BenchmarkParse_Bot(b *testing.B) {
	ua := "Googlebot/2.1 (+http://www.google.com/bot.html)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(ua)
	}
}