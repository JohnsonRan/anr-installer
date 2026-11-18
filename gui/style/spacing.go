package style

// The spacing scale. Every padding, gap and margin in the window is one of these, named by the relationship it
// spaces, so that the same relationship gets the same room everywhere (gui/AGENTS.md, "Layout, type and loading").
// The steps are multiples of 4 DIP.
const (
	Space1 = 4
	Space2 = 8
	space3 = 12
	Space4 = 16
	space6 = 24
)

const (
	PadPage     = Space4          // around a page's content and the list of games, and the title bar's ends
	PadPageX    = space6 + Space2 // at the sides of a page, and of the activity log under it
	GapSection  = Space4          // between the sections of a page
	GapCards    = space3          // between cards side by side
	GapCardRows = Space1          // between the cards of one group, as Windows 11's Settings stacks them
	// Above and below a group's heading: with it, a heading is taller than a button beside it, so that headings
	// with and without one take the same room.
	padHeadingY = Space2

	PadCardY  = space3 // inside a card, a callout or a header, above and below
	PadCardX  = Space4 // inside a card, a callout or a header, at the sides
	GapInCard = Space2 // between the lines of a card

	GapGroup    = space3 // between an icon or picture and the text beside it, and between groups in a bar
	GapControls = Space2 // between controls in a row, and between the items of a panel
	GapLines    = Space1 // between a name and the line under it

	PadRowY = Space2 // a row of the list, above and below
	PadRowX = space3 // a row of the list, at the sides

	GapLabelField = space3 // between a field's label and the field
	PadEmpty      = space6 // around an empty state

	GapLogEntries = Space2                           // between the entries of the activity log
	PadLogText    = PadPageX + LogMark + GapControls // where a task's title, steps and output start, past its mark
)

// Sizes that are not spacing but that spacing lines up with.
const (
	LabelWidth        = 72  // the label column of a card's fields
	HeaderIcon        = 64  // the game's icon in its page's header
	HeroHeight        = 200 // the header card of a game with library art
	HeroLogoW         = 280 // the box the game's logo fits in on it
	HeroLogoH         = 80
	HeroFade          = 64 // how far its darkening reaches up from the logo
	RowIcon           = 20 // the icon of a settings row or an info bar
	SettingsRowHeight = 68 // a settings row, as Windows 11's Settings draws one
	StatusDot         = 8
	ProgressWidth     = 160
	NoticeRadius      = 6
	CardRadius        = 8
	NotesWidth        = 600 // the release notes dialog
	NotesHeight       = 420
	ConfirmWidth      = 440 // a dialog that asks before going on
	TitleLogo         = 24
	TitleIcon         = 16 // the settings button's icon in the title bar
	CaptionIcon       = 10 // the window buttons' icons, as Windows draws them
	NavPillW          = 3  // the bar beside the chosen game in the list
	NavPillH          = 16
	DividerLine       = 2  // the divider between the two pictures of a comparison
	DividerHandle     = 16 // the grip on it
	LogMark           = 16 // a task's state in the activity log: a spinner, a check or a cross
	LogTimeW          = 40 // the activity log's column of start times, "15:04"
	LogTookW          = 80 // and of durations, "59 min 59 s" or "23 小时 59 分"
	// The open activity panel: about 10 lines of steps and output, or 8 with two tasks' headers. At the minimum
	// window size, 760 × 520, it and its bar leave the page about 240 DIP, its header and first card.
	ActivityHeight = 180
)
