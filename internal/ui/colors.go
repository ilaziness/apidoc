package ui

import "github.com/go-gui-org/go-gui/gui"

// Dialog chrome colors (go-gui theme roles do not cover title/body/rule split).
var (
	colorDialogTitleLight = gui.RGB(236, 239, 244)
	colorDialogTitleDark  = gui.RGB(45, 48, 56)
	colorDialogBodyLight  = gui.RGB(255, 255, 255)
	colorDialogBodyDark   = gui.RGB(32, 34, 40)
	colorDialogRuleLight  = gui.RGB(220, 224, 230)
	colorDialogRuleDark   = gui.RGB(58, 62, 70)
)

func themeIsDark() bool {
	c := gui.CurrentTheme().ColorBackground
	return int(c.R)+int(c.G)+int(c.B) < 384
}

func colorDialogTitle() gui.Color {
	if themeIsDark() {
		return colorDialogTitleDark
	}
	return colorDialogTitleLight
}

func colorDialogBody() gui.Color {
	if themeIsDark() {
		return colorDialogBodyDark
	}
	return colorDialogBodyLight
}

func colorDialogRule() gui.Color {
	if themeIsDark() {
		return colorDialogRuleDark
	}
	return colorDialogRuleLight
}
