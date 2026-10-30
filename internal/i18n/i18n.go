// Package i18n provides UI strings for en / zh-Hans / zh-Hant.
package i18n

import "github.com/ilaziness/apidoc/internal/config"

// Catalog holds translated UI strings.
type Catalog struct {
	AppTitle           string
	Home               string
	Collections        string
	NewCollection      string
	Name               string
	Rename             string
	Delete             string
	Save               string
	OK                 string
	Discard            string
	Send               string
	Cancel             string
	Close              string
	Environments       string
	NewEnvironment     string
	Settings           string
	Language           string
	Theme              string
	Proxy              string
	DataDir            string
	Browse             string
	Trash              string
	Restore            string
	EmptyTrash         string
	Folder             string
	Markdown           string
	HTTPRequest        string
	Method             string
	URL                string
	PathParams         string
	Query              string
	Headers            string
	Body               string
	Auth               string
	Options            string
	Params             string
	Response           string
	ImportOpenAPI      string
	ImportNew          string
	ImportCurrent      string
	FilePath           string
	ConfirmDelete      string
	ConfirmCopyDataDir string
	RestartRequired    string
	UnsavedPrompt      string
	NoSelection        string
	EditEnv            string
	Variables          string
	DeleteCollection   string
	Preview            string
	EditDoc            string
	Debug              string
	Description        string
	Sending            string
	Saved              string
	FollowRedirects    string
	SkipTLS            string
	Timeout            string
	ProxySystem        string
	ProxyOff           string
	ProxyManual        string
	ThemeLight         string
	ThemeDark          string
	ThemeSystem        string
}

// T returns the catalog for lang (falls back to English).
func T(lang string) Catalog {
	switch lang {
	case config.LangZHHans:
		return zhHans
	case config.LangZHHant:
		return zhHant
	default:
		return en
	}
}

var en = Catalog{
	AppTitle: "ApiDoc", Home: "Home", Collections: "Projects", NewCollection: "New project",
	Name: "Name", Rename: "Rename", Delete: "Delete", Save: "Save", OK: "OK", Discard: "Discard",
	Send: "Send", Cancel: "Cancel", Close: "Close",
	Environments: "Environments", NewEnvironment: "New environment", Settings: "Settings",
	Language: "Language", Theme: "Theme", Proxy: "Proxy", DataDir: "Data directory", Browse: "Choose…",
	Trash: "Trash", Restore: "Restore", EmptyTrash: "Empty trash",
	Folder: "Folder", Markdown: "Markdown", HTTPRequest: "HTTP request",
	Method: "Method", URL: "URL", PathParams: "Path params", Query: "Query",
	Headers: "Headers", Body: "Body", Auth: "Auth", Options: "Options", Params: "Params",
	Response:      "Response",
	ImportOpenAPI: "Import OpenAPI", ImportNew: "Import as new project", ImportCurrent: "Import into current project",
	FilePath: "File path", ConfirmDelete: "Move to trash?",
	ConfirmCopyDataDir: "Copy data to the new directory (same-name files overwritten). Restart the app afterwards.",
	RestartRequired:    "Data copied. Restart the app to use the new directory.",
	UnsavedPrompt:      "Save or discard changes first.", NoSelection: "Select an item",
	EditEnv: "Edit environment", Variables: "Variables", DeleteCollection: "Delete project",
	Preview: "Preview", EditDoc: "Edit", Debug: "Debug", Description: "Description",
	Sending: "Sending…", Saved: "Saved",
	FollowRedirects: "Follow redirects", SkipTLS: "Skip TLS verify", Timeout: "Timeout (sec)",
	ProxySystem: "System proxy", ProxyOff: "No proxy", ProxyManual: "Manual proxy",
	ThemeLight: "Light", ThemeDark: "Dark", ThemeSystem: "System",
}

var zhHans = Catalog{
	AppTitle: "ApiDoc", Home: "首页", Collections: "项目", NewCollection: "新建项目",
	Name: "名称", Rename: "重命名", Delete: "删除", Save: "保存", OK: "确定", Discard: "丢弃",
	Send: "发送", Cancel: "取消", Close: "关闭",
	Environments: "环境", NewEnvironment: "新建环境", Settings: "设置",
	Language: "界面语言", Theme: "主题", Proxy: "代理", DataDir: "数据目录", Browse: "选择…",
	Trash: "回收站", Restore: "还原", EmptyTrash: "清空回收站",
	Folder: "文件夹", Markdown: "Markdown", HTTPRequest: "HTTP 接口",
	Method: "方法", URL: "URL", PathParams: "路径参数", Query: "Query",
	Headers: "Header", Body: "Body", Auth: "鉴权", Options: "选项", Params: "参数",
	Response:      "响应",
	ImportOpenAPI: "导入 OpenAPI", ImportNew: "导入为新项目", ImportCurrent: "导入到当前项目",
	FilePath: "文件路径", ConfirmDelete: "移入回收站？",
	ConfirmCopyDataDir: "将把数据复制到新目录，同名文件会覆盖。完成后必须重启应用。",
	RestartRequired:    "数据已复制。请重启应用以使用新目录。",
	UnsavedPrompt:      "请先保存或丢弃更改。", NoSelection: "请选择一项",
	EditEnv: "编辑环境", Variables: "变量", DeleteCollection: "删除项目",
	Preview: "预览文档", EditDoc: "编辑文档", Debug: "调试", Description: "说明",
	Sending: "发送中…", Saved: "已保存",
	FollowRedirects: "跟随重定向", SkipTLS: "跳过 TLS 校验", Timeout: "超时（秒）",
	ProxySystem: "系统代理", ProxyOff: "不使用代理", ProxyManual: "手动代理",
	ThemeLight: "浅色", ThemeDark: "深色", ThemeSystem: "跟随系统",
}

var zhHant = Catalog{
	AppTitle: "ApiDoc", Home: "首頁", Collections: "專案", NewCollection: "新建專案",
	Name: "名稱", Rename: "重新命名", Delete: "刪除", Save: "儲存", OK: "確定", Discard: "捨棄",
	Send: "傳送", Cancel: "取消", Close: "關閉",
	Environments: "環境", NewEnvironment: "新建環境", Settings: "設定",
	Language: "介面語言", Theme: "主題", Proxy: "代理", DataDir: "資料目錄", Browse: "選擇…",
	Trash: "資源回收筒", Restore: "還原", EmptyTrash: "清空回收筒",
	Folder: "資料夾", Markdown: "Markdown", HTTPRequest: "HTTP 介面",
	Method: "方法", URL: "URL", PathParams: "路徑參數", Query: "Query",
	Headers: "Header", Body: "Body", Auth: "驗證", Options: "選項", Params: "參數",
	Response:      "回應",
	ImportOpenAPI: "匯入 OpenAPI", ImportNew: "匯入為新專案", ImportCurrent: "匯入到目前專案",
	FilePath: "檔案路徑", ConfirmDelete: "移入資源回收筒？",
	ConfirmCopyDataDir: "將把資料複製到新目錄，同名檔案會覆蓋。完成後必須重新啟動應用。",
	RestartRequired:    "資料已複製。請重新啟動應用以使用新目錄。",
	UnsavedPrompt:      "請先儲存或捨棄變更。", NoSelection: "請選擇一項",
	EditEnv: "編輯環境", Variables: "變數", DeleteCollection: "刪除專案",
	Preview: "預覽文件", EditDoc: "編輯文件", Debug: "偵錯", Description: "說明",
	Sending: "傳送中…", Saved: "已儲存",
	FollowRedirects: "跟隨重新導向", SkipTLS: "略過 TLS 驗證", Timeout: "逾時（秒）",
	ProxySystem: "系統代理", ProxyOff: "不使用代理", ProxyManual: "手動代理",
	ThemeLight: "淺色", ThemeDark: "深色", ThemeSystem: "跟隨系統",
}
