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
	padPage    = space4 // around a page's content, and the title bar's ends
	padPanel   = space3 // around the list of games
	gapSection = space4 // between the sections of a page
	gapCards   = space3 // between cards side by side

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
)

// Sizes that are not spacing but that spacing lines up with.
const (
	labelWidth = 72 // the label column of a card's fields
	heroFade   = 56 // how far the banner's darkening reaches up from its text
)
