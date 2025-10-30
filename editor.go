package main

// remember: moving cursor horizontally/mouse clicking/inserting/deleting invalidates the LastCursorPosition from lines

import (
	"fmt"
	"strconv"

	pt "main/piece-table"
	"main/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CursorPosition{
// 	Position: rl.NewVector2(
// 		e.Rectangle.X,
// 		e.Rectangle.Y,
// 	),
// 	Line:          0,
// 	Column:       0,
// 	CurrentIndex: 0,
// },

const (
	UPWARD   = -1
	DOWNWARD = 1
)

// -1 indicates that the values are missing
type CursorPosition struct {
	Position                   rl.Vector2
	Line, Column, CurrentIndex int
}

// @line
type Line struct {
	Start, Length int
	Rectangle     rl.Rectangle
	AutoNewLine   bool
}

// @cursor
type Cursor struct {
	Rectangle    rl.Rectangle
	Line, Column int
	CurrentIndex int
	TickTime     float32
	TickTimer    float32
	Color        rl.Color
}

func NewCursor(rectangle rl.Rectangle, line int, column int) Cursor {
	return Cursor{
		Rectangle:    rectangle,
		Line:         line,
		Column:       column,
		CurrentIndex: 0,
		TickTime:     0.5,
		Color:        rl.White,
	}
}

func (c *Cursor) SetPosition(currentIndex int, x float32, y float32, line int, column int) {
	c.CurrentIndex = currentIndex
	c.Rectangle.X = x
	c.Rectangle.Y = y
	c.Line = line
	c.Column = column
}

func (c *Cursor) Draw() {
	c.TickTimer += rl.GetFrameTime()
	// if c.TickTimer > c.TickTime {
	rl.DrawRectangle(
		c.Rectangle.ToInt32().X,
		c.Rectangle.ToInt32().Y,
		c.Rectangle.ToInt32().Width,
		c.Rectangle.ToInt32().Height,
		c.Color,
	)
	// }
	if c.TickTimer > 1 {
		c.TickTimer = 0
	}
}

type Action = int

const (
	NONE = iota
	TYPING
	DELETE
	CURSOR_MOVE
	MOUSE_LEFT_CLICK
)

// @editor
type Editor struct {
	CharRecCache        map[rune]rl.Vector2
	WritableRec         rl.Rectangle
	BackgroundColor     rl.Color
	Cursor              Cursor
	Lines               []*Line
	LastCursorPositions map[int]CursorPosition
	PieceTable          *pt.PieceTable
	Font                *rl.Font
	FontSize            int
	FontColor           rl.Color
	PreviousCharacter   rune
	LastLineVisited     int
	Actions             []Action
	LinesXPadding       float32
	InFocus             bool
	ShowLines           bool
	EditorRec           rl.Rectangle
	linesMaxVec         rl.Vector2
}

func NewEditor(rectangle rl.Rectangle, backgroundColor rl.Color) Editor {
	pieceTable := pt.NewPieceTable(pt.Sequence{})
	fontSize := 30
	defaultFont := rl.GetFontDefault()
	return Editor{
		WritableRec:         rectangle,
		EditorRec:           rectangle,
		BackgroundColor:     backgroundColor,
		PieceTable:          &pieceTable,
		FontSize:            fontSize,
		FontColor:           rl.White,
		Actions:             []Action{},
		Cursor:              NewCursor(rl.NewRectangle(rectangle.X, rectangle.Y, 3, float32(fontSize)), 0, 0),
		Lines:               make([]*Line, 0),
		LastCursorPositions: make(map[int]CursorPosition),
		CharRecCache:        make(map[rune]rl.Vector2),
		LinesXPadding:       15,
		InFocus:             false,
		ShowLines:           true,
		Font:                &defaultFont,
	}
}

func (e *Editor) ChangeFont(font *rl.Font) {
	e.Font = font
	oneLineWidth := e.CharRectangle('1')
	if oneLineWidth.X > e.linesMaxVec.X {
		e.linesMaxVec.X = oneLineWidth.X
		e.WritableRec.Width = e.EditorRec.Width - oneLineWidth.X - e.LinesXPadding
		e.WritableRec.X = e.EditorRec.X + oneLineWidth.X + oneLineWidth.X
	}
	if oneLineWidth.Y > e.linesMaxVec.Y {
		e.linesMaxVec.Y = oneLineWidth.Y
	}
}

func (e *Editor) Index() int {
	currentLine := e.Lines[e.Cursor.Line]
	return currentLine.Start + e.Cursor.Column
}

func (e *Editor) PreviousChar() (rune, error) {
	if e.Cursor.CurrentIndex <= 0 {
		return -1, fmt.Errorf("PreviousChar: error trying to get previous char. current index <= 0")
	}
	previous, err := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex - 1))
	if err != nil {
		return -1, err
	}
	return previous, nil
}

func (e *Editor) CurrentChar() (rune, error) {
	currentChar, err := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex))
	if err != nil {
		return -1, err
	}
	return currentChar, nil
}

func (e *Editor) FirstLine() *Line {
	return e.Lines[0]
}

func (e *Editor) LastLine() *Line {
	return e.Lines[len(e.Lines)-1]
}

func (e *Editor) PreviousLine() (*Line, error) {
	if e.Cursor.Line-1 < 0 {
		return nil, fmt.Errorf("PreviousLine: error trying to get previous line. current line - 1 < lines length")
	}
	return e.Lines[e.Cursor.Line-1], nil
}

func (e *Editor) CurrentLine() *Line {
	return e.Lines[e.Cursor.Line]
}

func (e *Editor) NextLine() (*Line, error) {
	if e.Cursor.Line+1 >= len(e.Lines) {
		return nil, fmt.Errorf("NextLine: error trying to get next line. current line + 1 > lines length")
	}
	return e.Lines[e.Cursor.Line+1], nil
}

func (e *Editor) CharRectangle(char rune) rl.Vector2 {
	fromCache, ok := e.CharRecCache[char]
	if ok {
		return fromCache
	}
	charSize := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0)
	if char == '\n' {
		charSize.Y = 30 // hardcoded for now, maybe the correct is to assign it the line's height mean
	}
	e.CharRecCache[char] = charSize
	return charSize
}

func (e *Editor) SequenceRectangle(sequence pt.Sequence) rl.Vector2 {
	var vector2 rl.Vector2
	for _, char := range sequence.RuneForward() {
		rec := e.CharRectangle(char)
		vector2.X += rec.X
	}
	return vector2
}

func (e *Editor) SetCursorPositionByIndex(index int) {
	if index < 0 || index > int(e.PieceTable.RuneLength) {
		return
	}
	lineIndex := e.FindLineByIndex(index, false)
	line := e.Lines[lineIndex]

	// TODO: Fast path to avoid searching the whole line
	// inTheEnd := (e.Cursor.Column == line.Length-1 && !line.AutoNewLine) || (e.Cursor.Column == line.Length && line.AutoNewLine)
	// lineModified := e.Cursor.Column > line.Length
	// if e.Cursor.CurrentIndex+1 == index && (!inTheEnd && !lineModified) {
	// 	e.MoveCursorForward()
	// 	// e.MoveCursorForward()
	// 	return
	// }
	// if e.Cursor.CurrentIndex-1 == index && (!inTheEnd && !lineModified){
	// 	e.MoveCursorBackward()
	// 	return
	// }

	sequence, _, err := e.PieceTable.GetSequence(uint(line.Start), uint(line.Length))
	if err != nil {
		return
	}
	var column int
	var previousChar rune
	var currentIndex int = line.Start
	positionX := line.Rectangle.X
	for _, char := range sequence.RuneForward() {
		if currentIndex == index {
			e.LastLineVisited = e.Cursor.Line
			if previousChar != '0' {
				e.PreviousCharacter = previousChar
			}
			e.Cursor.SetPosition(
				currentIndex,
				positionX,
				line.Rectangle.Y,
				lineIndex,
				column,
			)
			return
		}
		charSize := e.CharRectangle(char)
		column++
		currentIndex++
		positionX += charSize.X
		previousChar = char
	}
}

func (e *Editor) CalculateLines() {
	addLine := func(line *Line) {
		e.Lines = append(e.Lines, line)
		linesCountStr := strconv.Itoa(len(e.Lines))
		linesCountRec := e.SequenceRectangle(pt.Sequence(linesCountStr))
		if linesCountRec.Y > e.linesMaxVec.Y {
			e.linesMaxVec.Y = linesCountRec.Y
		}
		if linesCountRec.X > e.linesMaxVec.X {
			e.linesMaxVec.X = linesCountRec.X
			e.WritableRec.Width = e.EditorRec.Width - linesCountRec.X - e.LinesXPadding
			e.WritableRec.X = e.EditorRec.X + linesCountRec.X + linesCountRec.X
			e.CalculateLines()
			return
		}
	}

	currentLine := &Line{
		0,
		0,
		rl.NewRectangle(e.WritableRec.X, e.WritableRec.Y, 0, 0),
		false,
	}

	e.Lines = e.Lines[:0]
	var lastWidth float32 = -1
	var lastSpaceIndex int = -1
	var length int
	for i, char := range e.PieceTable.Runes() {
		charSize := e.CharRectangle(char)
		currentLine.Rectangle.Width += charSize.X
		length++
		// this serves to adjust the line height to the higher character found
		if currentLine.Rectangle.Height < charSize.Y {
			currentLine.Rectangle.Height = charSize.Y
		}

		charOutOfEditorBounds := currentLine.Rectangle.Width > e.WritableRec.Width
		if charOutOfEditorBounds {
			var innerLength int
			var width float32
			newLineStart := -1
			spaceNotFound := lastSpaceIndex == -1 || lastSpaceIndex < currentLine.Start
			currentLine.AutoNewLine = true
			if spaceNotFound {
				// wrap at the character
				currentLine.Length = i - currentLine.Start
				currentLine.Rectangle.Width -= charSize.X
				// e.Lines = append(e.Lines, currentLine)
				addLine(currentLine)
				newLineStart = i // might be wrong, perhaps newLineStart = i+1
				innerLength = 1
				width = charSize.X
			} else {
				// wrap the whole word
				innerLength = i - lastSpaceIndex
				width = currentLine.Rectangle.Width - lastWidth
				currentLine.Length = lastSpaceIndex - currentLine.Start + 1 // plus one because a line's interval is [start, length)
				currentLine.Rectangle.Width = lastWidth
				// e.Lines = append(e.Lines, currentLine)
				addLine(currentLine)
				charAfterSpace := lastSpaceIndex + 1
				newLineStart = charAfterSpace
			}
			currentLine = &Line{
				newLineStart,
				innerLength,
				rl.NewRectangle(e.WritableRec.X, currentLine.Rectangle.Y+currentLine.Rectangle.Height, width, 0),
				false,
			}
			lastSpaceIndex = -1
			length = innerLength
		} else if char == '\n' {
			currentLine.Length = length
			length = 0
			// e.Lines = append(e.Lines, currentLine)
			addLine(currentLine)
			currentLine = &Line{
				i + 1,
				0,
				rl.NewRectangle(e.WritableRec.X, currentLine.Rectangle.Y+currentLine.Rectangle.Height, 0, 0),
				false,
			}
		} else if char == ' ' {
			lastSpaceIndex = i
			lastWidth = currentLine.Rectangle.Width
		}
	}

	if length > 0 {
		currentLine.Length = length
		// e.Lines = append(e.Lines, currentLine)
		addLine(currentLine)
	}

	// ------------ Debugging ------------

	utils.Logger.Println("------------------------------------------------")
	for i, line := range e.Lines {
		utils.Logger.Println("Line: ", i, "| Line Start: ", line.Start, "| Line Length: ", line.Length, "| Line Start+Length: ", line.Start+line.Length, "| Line X: ", line.Rectangle.X, "| Line Y: ", line.Rectangle.Y, "| Line Width: ", line.Rectangle.Width, "| Line Height: ", line.Rectangle.Height)
	}
	utils.Logger.Println("Piece table PiecesAmount: ", e.PieceTable.PiecesAmount())

	// -----------------------------------

}

func CheckLinesCharactersLength(lines []*Line, text string) {
	totalLineChars := 0
	for _, line := range lines {
		totalLineChars += line.Length
	}
	if totalLineChars < len(text) {
		utils.Logger.Println("Warning: more characters than space in e.Lines", totalLineChars, len(text))
	}
}

func (e *Editor) DrawText() {
	// CheckLinesCharactersLength(e.Lines, e.PieceTable.ToString())

	currentLineIndex := 0
	currentLine := e.Lines[currentLineIndex]
	charXPosition := currentLine.Rectangle.X
	length := 0
	DrawLineNumber := func() {
		rl.DrawRectangle(e.EditorRec.ToInt32().X, currentLine.Rectangle.ToInt32().Y, int32(e.linesMaxVec.X)+int32(e.LinesXPadding), int32(e.linesMaxVec.Y), e.BackgroundColor)
		color := rl.NewColor(90, 90, 90, 255)
		if currentLineIndex == e.Cursor.Line {
			color = rl.White
		}
		rl.DrawTextEx(*e.Font, utils.IntToString(currentLineIndex+1), rl.NewVector2(e.EditorRec.X, currentLine.Rectangle.Y), float32(e.FontSize), 0, color)
	}
	for _, char := range e.PieceTable.ToString() {
		if length >= currentLine.Length {
			currentLineIndex++
			if currentLineIndex < len(e.Lines) {
				currentLine = e.Lines[currentLineIndex]
			}
			length = 0
			charXPosition = currentLine.Rectangle.X
		}
		length++
		stringChar := string(char)
		charSize := e.CharRectangle(char)
		rl.DrawTextEx(*e.Font, stringChar, rl.NewVector2(charXPosition, currentLine.Rectangle.Y), float32(e.FontSize), 0, e.FontColor)
		DrawLineNumber()
		charXPosition += charSize.X
	}
}

func (e *Editor) Draw() {
	rl.DrawRectanglePro(
		e.EditorRec,
		rl.NewVector2(0, 0),
		0,
		e.BackgroundColor,
	)
	e.DrawText()
	// if e.InFocus {
	e.Cursor.Draw()
	// }
}

func (e *Editor) SetFontSize(fontSize int) {
	e.FontSize = fontSize
	e.Cursor.Rectangle.Height = float32(fontSize)
}

func (e *Editor) FindPositionByLineColumn(line int, column int) float32 {
	lineToSearch := e.Lines[line]
	sequence, _, err := e.PieceTable.GetSequence(uint(lineToSearch.Start), uint(lineToSearch.Length))

	if err != nil {
		return -1
	}
	var width float32 = 0
	for i, char := range sequence.RuneForward() {
		if i == column {
			break
		}
		charSize := e.CharRectangle(char)
		width += charSize.X
	}
	return e.WritableRec.X + width
}

func (e *Editor) FindLineByIndex(index int, inclusive bool) int {
	for i, line := range e.Lines {
		end := line.Start + line.Length
		if !inclusive {
			end--
		}
		autoNewLine := line.Start <= index && index <= end && line.AutoNewLine && (e.Cursor.Column > 0 || e.Cursor.Column == 0 && e.Cursor.Line == i)
		notAutoNewLine := line.Start <= index && index < line.Start+line.Length && !line.AutoNewLine
		if autoNewLine || notAutoNewLine {
			return i
		}
	}
	return -1
}

// returns: lineIndex, line, inXBounds, column, index, columnXPosition, previousChar, error
func (e *Editor) FindLineClickMetadata(mouseClick rl.Vector2) (int, *Line, bool, int, int, float32, rune, error) {
	for i, line := range e.Lines {
		inLineXBoundaries := mouseClick.X >= line.Rectangle.X && mouseClick.X <= line.Rectangle.X+line.Rectangle.Width
		inLineYBoundaries := mouseClick.Y >= line.Rectangle.Y && mouseClick.Y <= line.Rectangle.Y+line.Rectangle.Height
		if inLineXBoundaries && inLineYBoundaries {
			sequence, _, err := e.PieceTable.GetSequence(uint(line.Start), uint(line.Length))
			if err != nil {
				return -1, nil, false, -1, -1, -1, -1, err
			}
			var previousCharacter rune
			var previousCharacterX float32
			var column int
			currentIndex := line.Start
			charXPosition := line.Rectangle.X
			for _, char := range sequence.RuneForward() {
				if char == '\n' {
					break
				}
				charSize := e.CharRectangle(char)
				betweenPostPreviousCharHalfAndPreCharHalf := mouseClick.X > previousCharacterX && mouseClick.X < charXPosition+(charSize.X/2)
				if betweenPostPreviousCharHalfAndPreCharHalf || char == '\n' && i == len(sequence) {
					break
				}
				previousCharacter = char
				previousCharacterX = charXPosition
				charXPosition += charSize.X
				currentIndex++
				column++
			}
			return i, line, true, column, currentIndex, charXPosition, previousCharacter, nil
		}
		// if it isn't on line X boundaries, it makes no sense to search the metadata
		if inLineYBoundaries {
			index := line.Start + line.Length
			column := line.Length
			if !line.AutoNewLine {
				index--
				column--
			}
			previousChar, _ := e.PieceTable.GetAt(uint(index - 1))
			return i, line, false, column, index, line.Rectangle.X + line.Rectangle.Width, previousChar, nil
		}
	}
	return -1, nil, false, -1, -1, -1, -1, nil
}

func (e *Editor) MoveCursorForward() {
	currentLine := e.CurrentLine()
	if e.Cursor.CurrentIndex == int(e.PieceTable.RuneLength)-1 {
		// if e.Cursor.Line == len(e.Lines)-1 && e.Cursor.Column == currentLine.Length-1 {
		return
	}
	clear(e.LastCursorPositions)
	currentChar, _ := e.CurrentChar()
	charSize := e.CharRectangle(currentChar)
	nextCharIsSpace := currentChar == ' '
	isEndOfLineSpace := e.Cursor.Column >= currentLine.Length-1 && nextCharIsSpace
	isEndOfLine := isEndOfLineSpace || e.Cursor.Column >= currentLine.Length
	isNewLine := currentChar == '\n'
	isCharacter := !isEndOfLine && !isNewLine
	if isCharacter {
		e.PreviousCharacter = currentChar
		e.Cursor.SetPosition(
			e.Cursor.CurrentIndex+1,
			e.Cursor.Rectangle.X+charSize.X,
			currentLine.Rectangle.Y,
			e.Cursor.Line,
			e.Cursor.Column+1,
		)
	} else {
		e.LastLineVisited = e.Cursor.Line
		e.Cursor.Column = 0
		e.Cursor.Rectangle.X = e.WritableRec.X
		if isNewLine || isEndOfLineSpace {
			e.Cursor.CurrentIndex++
		}
		nextLine, _ := e.NextLine()
		e.Cursor.Rectangle.Y = nextLine.Rectangle.Y
		e.Cursor.Line++
	}
}

func (e *Editor) MoveCursorBackward() {
	if e.Cursor.CurrentIndex == 0 {
		return
	}
	clear(e.LastCursorPositions)
	currentChar, _ := e.CurrentChar()
	shouldGoToPreviousLine := e.Cursor.Column == 0 && e.Cursor.Line > 0
	if shouldGoToPreviousLine {
		e.LastLineVisited = e.Cursor.Line
		previousLine, _ := e.PreviousLine()
		newColumn := previousLine.Length
		newCurrentIndex := previousLine.Start + previousLine.Length
		if !previousLine.AutoNewLine {
			newCurrentIndex--
			newColumn--
		}

		newPosition := e.WritableRec.X + previousLine.Rectangle.Width
		previousChar, _ := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex - 1))
		if previousLine.AutoNewLine && previousChar == ' ' {
			previousCharSize := e.CharRectangle(previousChar)
			newPosition -= previousCharSize.X
			newCurrentIndex--
			newColumn--
		}
		e.Cursor.SetPosition(
			newCurrentIndex,
			newPosition,
			previousLine.Rectangle.Y,
			e.Cursor.Line-1,
			newColumn,
		)
	} else {
		previousCharacter, _ := e.PreviousChar()
		charSize := e.CharRectangle(previousCharacter)
		e.Cursor.Rectangle.X -= charSize.X
		e.Cursor.CurrentIndex--
		e.Cursor.Column--
	}
	e.PreviousCharacter = currentChar
}

func (e *Editor) _internalMoveCursorBackwardOrDownward(direction int) {
	e.LastLineVisited = e.Cursor.Line
	var line *Line
	var lastCursorPosition CursorPosition
	var ok bool
	var newCurrentIndex int
	var newLine int
	var shouldDecreaseColumnAndIndex bool = false

	if direction == UPWARD {
		line, _ = e.PreviousLine()
		lastCursorPosition, ok = e.LastCursorPositions[e.Cursor.Line-1]
		newLine = e.Cursor.Line - 1
		newCurrentIndex = e.Cursor.CurrentIndex - (e.Cursor.Column + line.Length - e.Cursor.Column)
		shouldDecreaseColumnAndIndex = e.Cursor.Line-1 == 0 || !line.AutoNewLine
	}
	if direction == DOWNWARD {
		line, _ = e.NextLine()
		lastCursorPosition, ok = e.LastCursorPositions[e.Cursor.Line+1]
		newLine = e.Cursor.Line + 1
		newCurrentIndex = e.Cursor.CurrentIndex + (e.CurrentLine().Length - e.Cursor.Column + e.Cursor.Column)
		shouldDecreaseColumnAndIndex = !line.AutoNewLine
	}
	if ok {
		e.Cursor.SetPosition(
			lastCursorPosition.CurrentIndex,
			lastCursorPosition.Position.X,
			lastCursorPosition.Position.Y,
			newLine,
			lastCursorPosition.Column,
		)
	} else if e.Cursor.Column >= line.Length {
		e.LastCursorPositions[e.Cursor.Line] = CursorPosition{
			Position:     rl.Vector2{X: e.Cursor.Rectangle.X, Y: e.Cursor.Rectangle.Y},
			Line:         e.Cursor.Line,
			Column:       e.Cursor.Column,
			CurrentIndex: e.Cursor.CurrentIndex,
		}
		newColumn := line.Length
		newCurrentIndex := line.Start + line.Length
		if shouldDecreaseColumnAndIndex {
			newColumn--
			newCurrentIndex -= 1
		}
		e.Cursor.SetPosition(
			newCurrentIndex,
			line.Rectangle.X+line.Rectangle.Width,
			line.Rectangle.Y,
			newLine,
			newColumn,
		)
	} else {
		newX := e.FindPositionByLineColumn(newLine, e.Cursor.Column)
		e.Cursor.SetPosition(
			newCurrentIndex,
			newX,
			line.Rectangle.Y,
			newLine,
			e.Cursor.Column,
		)
	}
}

func (e *Editor) MoveCursorUpward() {
	if e.Cursor.Line == 0 {
		return
	}
	e._internalMoveCursorBackwardOrDownward(UPWARD)
}

func (e *Editor) MoveCursorDownward() {
	if e.Cursor.Line >= len(e.Lines)-1 {
		return
	}
	e._internalMoveCursorBackwardOrDownward(DOWNWARD)
}

func (e *Editor) SetCursorPositionByClick(mouseClick rl.Vector2) error {
	lineIndex, line, _, column, index, xPosition, previousChar, err := e.FindLineClickMetadata(mouseClick)
	if err != nil {
		return err
	}
	e.LastLineVisited = e.Cursor.Line
	clear(e.LastCursorPositions)
	if line != nil {
		e.Cursor.SetPosition(
			index,
			xPosition,
			line.Rectangle.Y,
			lineIndex,
			column,
		)
		e.PreviousCharacter = previousChar
	}
	if line == nil {
		// TODO: it needs to be within editor boundaries, if not, return
		lastLine := e.LastLine()
		index := lastLine.Start + lastLine.Length
		column := lastLine.Length
		if !lastLine.AutoNewLine {
			index--
			column--
		}
		previousChar, _ := e.PieceTable.GetAt(uint(index - 1))
		e.PreviousCharacter = previousChar
		xPosition := lastLine.Rectangle.X + lastLine.Rectangle.Width
		e.Cursor.SetPosition(
			index,
			xPosition,
			lastLine.Rectangle.Y,
			len(e.Lines)-1,
			column,
		)
	}
	return nil
}

func (e *Editor) Insert(index int, sequence pt.Sequence) {
	size, _ := e.PieceTable.Insert(uint(index), sequence)
	e.CalculateLines()
	e.SetCursorPositionByIndex(index + int(size))
}

func (e *Editor) Delete(index int, length int) {
	_ = e.PieceTable.Delete(uint(index-length), uint(length))
	e.CalculateLines()
	e.SetCursorPositionByIndex(index - length)
}
