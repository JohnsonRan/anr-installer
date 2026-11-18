package style

import (
	"github.com/egoist/mygo/ui"

	"anr-installer/gui/i18n"
)

// Section is a group of cards under its heading, as Windows 11's Settings has them, with the group's actions at
// the heading's end.
func Section(c *ui.Context, t *ui.Theme, title string, actions func(), cards func()) {
	ui.Column(c).Gap(GapCardRows).Children(func() {
		ui.Row(c).Gap(GapControls).Children(func() {
			RoleBodyStrong.On(t, ui.Text(c, title)).PaddingY(padHeadingY).Grow(1).Shrink(1)
			if actions != nil {
				actions()
			}
		})
		cards()
	})
}

// PageHeader is the title row of a page shown in place of a game's, as the settings or a comparison: the way back,
// the page's title over a line about it, if any, and the page's actions at its end. The title lines up with the
// button by its text, so that it sits in one place on every page; where the window is too narrow for the actions
// beside the title, they go under it.
func PageHeader(c *ui.Context, t *ui.Theme, back func(), title, line string, actions func()) {
	ui.Row(c).Wrap().Justify(ui.SpaceBetween).AlignItems(ui.Baseline).GapX(GapGroup).GapY(GapInCard).Children(func() {
		ui.Row(c).Gap(GapGroup).AlignItems(ui.Baseline).Shrink(1).Children(func() {
			ui.Button(c, i18n.T("back")).OnClick(back)
			ui.Column(c).Shrink(1).Gap(GapLines).Children(func() {
				RoleHeading.On(t, ui.Text(c, title)).MaxLines(1).Ellipsis("…").Tooltip(title)
				if line != "" {
					RoleSmall.On(t, ui.Text(c, line)).TextColor(t.TextMuted).MaxLines(1).Ellipsis("…").Tooltip(line)
				}
			})
		})
		if actions != nil {
			ui.Row(c).Gap(GapControls).Children(actions)
		}
	})
}

// InfoBar is a tinted bar with an icon in its color, a message, and controls at its end.
func InfoBar(c *ui.Context, t *ui.Theme, color ui.Color, icon *ui.SVG, s string, controls func()) {
	box := ui.Row(c).Padding(PadCardY, PadCardX).Radius(CardRadius).Background(color.Alpha(0.10)).Border(1, color.Alpha(0.30))
	Callout(c, box, func() {
		ui.Icon(c, icon).Size(RowIcon, RowIcon).TextColor(color)
		ui.Text(c, s)
	}, controls)
}

// Callout lays out box, an info bar, as Windows 11's InfoBar: its icon, message and actions on one line, centered
// on it, with the actions at its end, where they fit; else the actions on a line of their own under the message.
// message draws the icon, rowIcon wide, and the message, whose lines start beside the icon.
func Callout(c *ui.Context, box ui.Element, message func(), actions func()) {
	// A wrapped line has one child, which Justify(SpaceBetween) puts at its start.
	box.Wrap().Justify(ui.SpaceBetween).AlignItems(ui.Center).GapX(GapGroup).GapY(GapInCard).Children(func() {
		ui.Row(c).Gap(GapGroup).AlignItems(ui.Start).Shrink(1).Children(message)
		if actions != nil {
			ui.Row(c).Gap(GapControls).PaddingStart(RowIcon + GapGroup).Children(actions)
		}
	})
}

// SettingsRow is a card of Windows 11's Settings: an icon, a title over its description, and what goes at its
// end, a value or controls. The title and the description wrap.
func SettingsRow(c *ui.Context, t *ui.Theme, icon *ui.SVG, title, description string, end func()) {
	settingsCard(c, t, icon, end, func() {
		ui.Text(c, title)
		if description != "" {
			RoleSmall.On(t, ui.Text(c, description)).TextColor(t.TextMuted)
		}
	})
}

// ValueRow is a SettingsRow whose description is a value, as a path, or whose title is, as a date: each stays on
// one line, a value cut in its middle where the row is too narrow, and shows in full on hover.
func ValueRow(c *ui.Context, t *ui.Theme, icon *ui.SVG, title, value string, end func()) {
	settingsCard(c, t, icon, end, func() {
		ui.Text(c, title).MaxLines(1).Ellipsis("…").Tooltip(title)
		RoleSmall.On(t, ui.Text(c, value)).TextColor(t.TextMuted).MaxLines(1).TruncateMiddle().Tooltip(value)
	})
}

// settingsCard lays a card out as Windows 11's settings cards: what goes at its end goes under its text, past the
// icon, where the card is too narrow for both on one line.
func settingsCard(c *ui.Context, t *ui.Theme, icon *ui.SVG, end func(), text func()) {
	ui.Row(c).Gap(GapGroup).Padding(PadCardY, PadCardX).MinHeight(SettingsRowHeight).Radius(CardRadius).
		Background(ApexSurface).Border(1, ApexLine).Children(func() {
		ui.Icon(c, icon).Size(RowIcon, RowIcon).TextColor(t.TextMuted)
		// A wrapped line has one child, which Justify(SpaceBetween) puts at its start.
		ui.Row(c).Grow(1).Shrink(1).Wrap().Justify(ui.SpaceBetween).AlignItems(ui.Center).GapX(GapGroup).GapY(GapInCard).Children(func() {
			ui.Column(c).Shrink(1).Gap(GapLines).Children(text)
			if end != nil {
				ui.Row(c).Gap(GapControls).Children(end)
			}
		})
	})
}
