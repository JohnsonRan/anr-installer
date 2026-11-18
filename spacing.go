package main

// The spacing scale. Every padding, gap and margin in the window is one of these, named by the relationship it
// spaces, so that the same relationship gets the same room everywhere (UI.md, "Layout, type and loading").
// The steps are multiples of 4 DIP.
const (
	space1 = 4
	space2 = 8
	space3 = 12
	space4 = 16
	space6 = 24
)

const (
	padPage     = space4          // around a page's content, and the title bar's ends
	padPageX    = space6 + space2 // at the sides of a game's page
	padPanel    = space3          // around the list of games
	gapSection  = space4          // between the sections of a page
	gapCards    = space3          // between cards side by side
	gapCardRows = space1          // between the cards of one group, as Windows 11's Settings stacks them

	padCardY  = space3 // inside a card, a callout or a header, above and below
	padCardX  = space4 // inside a card, a callout or a header, at the sides
	gapInCard = space2 // between the lines of a card

	gapGroup    = space3 // between an icon or picture and the text beside it, and between groups in a bar
	gapControls = space2 // between controls in a row, and between the items of a panel
	gapLines    = space1 // between a name and the line under it

	padRowY = space2 // a row of the list, above and below
	padRowX = space3 // a row of the list, at the sides

	gapLabelField = space3 // between a field's label and the field
	padEmpty      = space6 // around an empty state

	gapLogEntries = space2                          // between the entries of the activity log
	padLogText    = padPage + logMark + gapControls // where a task's title, steps and output start, past its mark
)

// Sizes that are not spacing but that spacing lines up with.
const (
	labelWidth        = 72  // the label column of a card's fields
	headerIcon        = 64  // the game's icon in its page's header
	heroHeight        = 200 // the header card of a game with library art
	heroLogoW         = 280 // the box the game's logo fits in on it
	heroLogoH         = 80
	heroFade          = 64 // how far its darkening reaches up from the logo
	rowIcon           = 20 // the icon of a settings row or an info bar
	settingsRowHeight = 68 // a settings row, as Windows 11's Settings draws one
	statusDot         = 8
	progressWidth     = 160
	noticeRadius      = 6
	cardRadius        = 8
	notesWidth        = 600 // the release notes dialog
	notesHeight       = 420
	confirmWidth      = 440 // a dialog that asks before going on
	titleLogo         = 24
	titleIcon         = 16 // the settings button's icon in the title bar
	captionIcon       = 10 // the window buttons' icons, as Windows draws them
	navPillW          = 3  // the bar beside the chosen game in the list
	navPillH          = 16
	dividerLine       = 2  // the divider between the two pictures of a comparison
	dividerHandle     = 16 // the grip on it
	logMark           = 16 // a task's state in the activity log: a spinner, a check or a cross
	logTimeW          = 40 // the activity log's column of start times, "15:04"
	logTookW          = 80 // and of durations, "59 min 59 s" or "23 小时 59 分"
	// The open activity panel: about 10 lines of steps and output, or 8 with two tasks' headers. At the minimum
	// window size, 760 × 520, it and its bar leave the page about 240 DIP, its header and first card.
	activityHeight = 180
)
