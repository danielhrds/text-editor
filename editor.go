package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// @line
type Row struct {
	Start, Length      int
	Rectangle          rl.Rectangle
	LastCursorPosition rl.Vector2
}

// @cursor
type Cursor struct {
	Rectangle    rl.Rectangle
	Row, Column  int
	CurrentIndex int
	LastTick     time.Time
	Color        rl.Color
}

func NewCursor(rectangle rl.Rectangle, row int, column int) Cursor {
	return Cursor{
		Rectangle:    rectangle,
		Row:          row,
		Column:       column,
		LastTick:     time.Now(),
		CurrentIndex: 0,
		Color:        rl.White,
	}
}

func (c *Cursor) SetPosition(currentIndex int, x float32, y float32, row int, column int) {
	c.CurrentIndex = currentIndex
	c.Rectangle.X = x
	c.Rectangle.Y = y
	c.Row = row
	c.Column = column
}

func (c *Cursor) Draw() {
	if time.Since(c.LastTick).Milliseconds() >= 1000 {
		c.LastTick = time.Now()
	}

	// if time.Since(c.LastTick).Milliseconds() >= 0 && time.Since(c.LastTick).Milliseconds() <= 500 {
	rl.DrawRectangle(
		c.Rectangle.ToInt32().X,
		c.Rectangle.ToInt32().Y,
		c.Rectangle.ToInt32().Width,
		c.Rectangle.ToInt32().Height,
		c.Color,
	)
	// }
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
	CharRecCache      map[rune]rl.Vector2
	Rectangle         rl.Rectangle
	BackgroundColor   rl.Color
	InFocus           bool
	Cursor            Cursor
	Rows              []*Row
	PieceTable        *PieceTable
	Font              *rl.Font
	FontSize          int
	FontColor         rl.Color
	PreviousCharacter rune
	PreviousRow       int
	Actions           []Action
}

func NewEditor(rectangle rl.Rectangle, backgroundColor rl.Color) Editor {
	pieceTable := NewPieceTable(Sequence{})
	fontSize := 30
	return Editor{
		Rectangle:       rectangle,
		BackgroundColor: backgroundColor,
		PieceTable:      &pieceTable,
		FontSize:        fontSize,
		FontColor:       rl.White,
		Actions:         []Action{},
		InFocus:         false,
		Cursor:          NewCursor(rl.NewRectangle(rectangle.X, rectangle.Y, 3, float32(fontSize)), 0, 0),
		Rows:            make([]*Row, 0),
		CharRecCache:    make(map[rune]rl.Vector2),
	}
}

func (e *Editor) Index() int {
	currentRow := e.Rows[e.Cursor.Row]
	return currentRow.Start + e.Cursor.Column
}

func (e *Editor) AddAction(action Action) {
	e.Actions = append(e.Actions, action)
}

func (e *Editor) LastAction() Action {
	action := NONE
	if len(e.Actions) > 0 {
		action = e.Actions[len(e.Actions)-1]
	}
	return action
}

func (e *Editor) CharRectangle(char rune) rl.Vector2 {
	fromCache, ok := e.CharRecCache[char]
	if ok {
		return fromCache
	}
	charRec := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0)
	if char == '\n' {
		charRec.Y = 30 // hardcoded for now
	}
	e.CharRecCache[char] = charRec
	return charRec
}

func (e *Editor) SequenceRectangle(sequence Sequence) rl.Vector2 {
	var vector2 rl.Vector2
	for _, char := range sequence {
		rec := e.CharRectangle(char)
		vector2.X += rec.X
	}
	return vector2
}

func (e *Editor) CalculateRows() {
	currentRow := &Row{
		0,
		0,
		rl.NewRectangle(e.Rectangle.X, e.Rectangle.Y, 0, 0),
		rl.NewVector2(0, 0),
	}

	e.Rows = e.Rows[:0]
	length := 0
	newLine := false
	lastSpaceIndex := -1
	var lastWidth float32 = -1
	text := e.PieceTable.ToString()
	for i := 0; i < len(text); i++ {
		// Creates a new Row
		if newLine {
			newLine = false
			lastSpaceIndex = -1
			currentRow = &Row{
				i,
				0,
				rl.NewRectangle(e.Rectangle.X, currentRow.Rectangle.Y+currentRow.Rectangle.Height, 0, 0),
				rl.NewVector2(0, 0),
			}
		}

		length++
		char := text[i]
		textSizeVector2 := e.CharRectangle(rune(char))
		currentRow.Rectangle.Width += textSizeVector2.X
		if char == ' ' {
			lastSpaceIndex = i
			lastWidth = currentRow.Rectangle.Width
		}
		if char == '\n' {
			newLine = true
			currentRow.Length = length
			length = 0
			e.Rows = append(e.Rows, currentRow)
		} else {
			charPosition := currentRow.Rectangle.X + currentRow.Rectangle.Width
			editorXBoundary := e.Rectangle.X + e.Rectangle.Width
			charOutOfEditorBounds := charPosition > editorXBoundary
			if charOutOfEditorBounds {
				newRowStart := -1
				length = 0
				width := float32(0)
				if lastSpaceIndex == -1 || lastSpaceIndex < currentRow.Start {
					// if a space isn't found within a row then we will wrap at the character
					currentRow.Length = i - currentRow.Start
					currentRow.Rectangle.Width -= textSizeVector2.X
					e.Rows = append(e.Rows, currentRow)
					newRowStart = i
					length = 1
					width = textSizeVector2.X
					logger.Println("width when splitting", currentRow.Rectangle.Width, charPosition, editorXBoundary)
				} else {
					// if a space is found within a row then we will wrap the whole word
					currentRow.Length = lastSpaceIndex - currentRow.Start + 1
					currentRow.Rectangle.Width = lastWidth
					e.Rows = append(e.Rows, currentRow)
					charBeforeSpace := lastSpaceIndex
					charAfterSpace := charBeforeSpace + 1
					newRowStart = charAfterSpace
					i = charBeforeSpace
				}
				currentRow = &Row{
					newRowStart,
					length,
					rl.NewRectangle(e.Rectangle.X, currentRow.Rectangle.Y+currentRow.Rectangle.Height, width, 0),
					rl.NewVector2(0, 0),
				}
				lastSpaceIndex = -1
			}

		}
		if currentRow.Rectangle.Height < textSizeVector2.Y {
			// this serves to adjust the row height to the higher character found
			currentRow.Rectangle.Height = textSizeVector2.Y
		}
	}

	// append the final row if it needs to, otherwise it might not be added
	if length > 0 {
		currentRow.Length = length
		e.Rows = append(e.Rows, currentRow)
	}

	// ------------ Debugging ------------

	logger.Println("------------------------------------------------")
	for i, row := range e.Rows {
		logger.Println("Row: ", i, "| Row Start: ", row.Start, "| Row Length: ", row.Length, "| Row Start+Length: ", row.Start+row.Length, "| Row X: ", row.Rectangle.X, "| Row Y: ", row.Rectangle.Y, "| Row Width: ", row.Rectangle.Width, "| Row Height: ", row.Rectangle.Height)
	}

	// -----------------------------------

}

func CheckRowsCharactersLength(rows []*Row, text string) {
	totalRowChars := 0
	for _, row := range rows {
		totalRowChars += row.Length
	}
	if totalRowChars < len(text) {
		logger.Println("Warning: more characters than space in e.Rows", totalRowChars, len(text))
	}
}

func (e *Editor) DrawText() {
	CheckRowsCharactersLength(e.Rows, e.PieceTable.ToString())
	currentRowIndex := 0
	currentRow := e.Rows[currentRowIndex]
	charXPosition := currentRow.Rectangle.X
	length := 0
	for _, char := range e.PieceTable.ToString() {
		if length >= currentRow.Length {
			currentRowIndex++
			if currentRowIndex < len(e.Rows) {
				currentRow = e.Rows[currentRowIndex]
			}
			length = 0
			charXPosition = currentRow.Rectangle.X
		}
		length++
		stringChar := string(char)
		textSizeVector2 := e.CharRectangle(char)
		rl.DrawTextEx(*e.Font, stringChar, rl.NewVector2(charXPosition, currentRow.Rectangle.Y), float32(e.FontSize), 0, e.FontColor)
		charXPosition += textSizeVector2.X
	}
}

func (e *Editor) Draw() {
	rl.DrawRectanglePro(
		e.Rectangle,
		rl.NewVector2(0, 0),
		0,
		e.BackgroundColor,
	)
	e.DrawText()
	if e.InFocus {
		e.Cursor.Draw()
	}
}

func (e *Editor) SetFontSize(fontSize int) {
	e.FontSize = fontSize
	e.Cursor.Rectangle.Height = float32(fontSize)
}

func (e *Editor) FindRowFromPosition(mouseClickPosition rl.Vector2) (*Row, int) {
	for index, row := range e.Rows {
		if mouseClickPosition.Y > row.Rectangle.Y && mouseClickPosition.Y < row.Rectangle.Y+row.Rectangle.Height {
			return row, index
		}
	}
	return nil, 1
}

func (e *Editor) MoveCursorForward() {
	currentChar, _ := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex))
	currentRow := e.Rows[e.Cursor.Row]
	isEndOfRow := e.Cursor.Column >= currentRow.Length-1 || e.Cursor.Column >= currentRow.Length-1 && e.LastAction() == TYPING
	isNewLine := currentChar == '\n'
	charRec := e.CharRectangle(currentChar)
	isCharacter := !isNewLine && !isEndOfRow
	if isCharacter {
		e.PreviousCharacter = currentChar
		e.Cursor.SetPosition(
			e.Cursor.CurrentIndex+1,
			e.Cursor.Rectangle.X+charRec.X,
			currentRow.Rectangle.Y,
			e.Cursor.Row,
			e.Cursor.Column+1,
		)
		// e.Cursor.Column++
		// e.Cursor.Rectangle.X += charRec.X
		// e.Cursor.CurrentIndex++
	} else {
		e.Cursor.Column = 0
		e.Cursor.Rectangle.X = e.Rectangle.X
		if isNewLine {
			e.Cursor.CurrentIndex++
		}
		if e.Cursor.Row < len(e.Rows) {
			nextRow := e.Rows[e.Cursor.Row+1]
			e.Cursor.Rectangle.Y = nextRow.Rectangle.Y
			e.PreviousRow = e.Cursor.Row
			e.Cursor.Row++
		}
		// if e.LastAction() == TYPING {
		// 	e.MoveCursorForward()
		// }
	}
}

func (e *Editor) MoveCursorBackward() {
	previousCharacter, _ := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex - 1))
	previousRow := &Row{Start: -1}
	if e.Cursor.Row > 0 {
		previousRow = e.Rows[e.Cursor.Row-1]
	}
	e.Cursor.Column--
	shouldGoToPreviousRow := previousRow.Start != -1 && e.Cursor.CurrentIndex == previousRow.Start+previousRow.Length
	// shouldGoToPreviousRow := previousRow.Start != -1 && e.Cursor.Column == previousRow.Start+previousRow.Length
	if shouldGoToPreviousRow && e.Cursor.Row > 0 {
		previousRow := e.Rows[e.Cursor.Row-1]
		e.Cursor.Rectangle.Y = previousRow.Rectangle.Y
		e.Cursor.Rectangle.X = e.Rectangle.X + previousRow.Rectangle.Width
		e.Cursor.Column = previousRow.Length-1
		e.PreviousRow = e.Cursor.Row
		e.Cursor.Row--
		isFromPreviousRow := e.Cursor.CurrentIndex-1 >= previousRow.Start && e.Cursor.CurrentIndex-1 < previousRow.Start+previousRow.Length
		isNewLineFromPreviousRow := previousCharacter == '\n'
		if isFromPreviousRow && isNewLineFromPreviousRow {
			e.Cursor.CurrentIndex--
		}
	} else {
		charRec := e.CharRectangle(e.PreviousCharacter)
		e.Cursor.Rectangle.X -= charRec.X
		e.Cursor.CurrentIndex--
	}
}

func (e *Editor) SetCursorPositionFromIndex(newPosition int) {
	currentRow := e.Rows[e.Cursor.Row]
	inCurrentRowBoundaries := newPosition >= currentRow.Start && newPosition <= currentRow.Start+currentRow.Length-1
	if inCurrentRowBoundaries {
		if newPosition > e.Cursor.CurrentIndex {
			iterations := newPosition - e.Cursor.CurrentIndex
			for range iterations {
				e.MoveCursorForward()
			}
		} else if newPosition < e.Cursor.CurrentIndex {
			iterations := e.Cursor.CurrentIndex - newPosition
			for range iterations {
				e.MoveCursorBackward()
			}
		}
	} else {
		for _, row := range e.Rows {
			foundRow := newPosition >= row.Start && newPosition <= row.Start+row.Length-1 || e.Cursor.Column == 1
			if foundRow {
				if newPosition > e.Cursor.CurrentIndex {
					iterations := newPosition - e.Cursor.CurrentIndex
					for range iterations {
						e.MoveCursorForward()
					}
				}
				if newPosition < e.Cursor.CurrentIndex {
					iterations := e.Cursor.CurrentIndex - newPosition
					for range iterations {
						e.MoveCursorBackward()
					}
				}
				return
			}
		}
	}
}

func (e *Editor) FindRowFromClick(mouseClickPosition rl.Vector2) (*Row, int) {
	for i, row := range e.Rows {
		if mouseClickPosition.Y > row.Rectangle.Y && mouseClickPosition.Y < row.Rectangle.Y+row.Rectangle.Height {
			return row, i
		}
	}
	return nil, -1
}

func (e *Editor) SetCursorPositionFromClickRow(rowIndex int, mouseClickPosition rl.Vector2) {
	row := e.Rows[rowIndex]
	if row != nil {
		sequence, _, err := e.PieceTable.GetSequence(uint(row.Start), uint(row.Length))
		if err != nil {
			return
		}
		var previousCharacter rune
		var previousCharacterX float32
		column := 0
		currentIndex := row.Start
		charXPosition := row.Rectangle.X
		iterableText := string(sequence)
		logger.Println(iterableText, len(iterableText))
		for i, char := range sequence {
			if char == '\n' {
				continue
			}
			charRec := e.CharRectangle(char)
			betweenPostPreviousCharHalfAndPreCharHalf := mouseClickPosition.X > previousCharacterX && mouseClickPosition.X < charXPosition+(charRec.X/2)
			if betweenPostPreviousCharHalfAndPreCharHalf || char == '\n' && i == len(sequence) {
				break
			}
			previousCharacter = char
			previousCharacterX = charXPosition
			charXPosition += charRec.X
			currentIndex++
			column++
		}
		e.Cursor.SetPosition(currentIndex, charXPosition, row.Rectangle.Y, rowIndex, column)
		e.PreviousCharacter = previousCharacter
	}
}

func (e *Editor) SetCursorPositionFromClick(mouseClickPosition rl.Vector2) {
	_, rowIndex := e.FindRowFromClick(mouseClickPosition)
	if rowIndex >= 0 {
		e.SetCursorPositionFromClickRow(rowIndex, mouseClickPosition)
	}
}

func (e *Editor) SetCursorToPreviousRow(column int) {
	if e.Cursor.Row > 0 {

	}
}

func (e *Editor) Insert(index int, text Sequence) {
	e.AddAction(TYPING)
	if len(text) == 0 {
		return
	}
	sequenceFirstChar := text[0]
	if sequenceFirstChar == '\n' {
		e.PieceTable.Insert(uint(index), []rune{text[0]})
		text = text[1:]
	}
	if len(text) > 0 {
		e.PieceTable.Insert(uint(index), text)
	}
	currentRowBeforeCalc := e.Rows[e.Cursor.Row]
	e.CalculateRows()
	currentRowAfterCalc := e.Rows[e.Cursor.Row]
	sequenceRectangle := e.SequenceRectangle(text)
	if currentRowAfterCalc.Length < currentRowBeforeCalc.Length || sequenceFirstChar == '\n' {
		// to me, it seems reasonable to think that: 
		// if the current row decreased it's Length, then a word has been wrapped
		e.Cursor.Row++
		currentRow := e.Rows[e.Cursor.Row]
		e.Cursor.SetPosition(
			currentRow.Start+currentRow.Length-1,
			currentRow.Rectangle.Width,
			currentRow.Rectangle.Y,
			e.Cursor.Row,
			currentRow.Length-1,
		)
	} else {
		e.Cursor.SetPosition(
			e.Cursor.CurrentIndex+1,
			e.Cursor.Rectangle.X+sequenceRectangle.X,
			e.Cursor.Rectangle.Y,
			e.Cursor.Row,
			e.Cursor.Column+1,
		)
	}
}

func (e *Editor) Delete(index int, length int) {
	e.AddAction(DELETE)
	// TODO: Add suport to multichar deletion
	sequence, _, _ := e.PieceTable.GetSequence(uint(e.Cursor.CurrentIndex)-1, uint(length))
	sequenceRectangle := e.SequenceRectangle(sequence)
	e.PieceTable.Delete(uint(index), uint(length))
	e.CalculateRows()
	if e.Cursor.Row > 0 && e.Cursor.Column == 0 {
		// send cursor to last char of the previous row
		previousRow := e.Rows[e.Cursor.Row-1]
		e.PreviousRow = e.Cursor.Row
		e.Cursor.SetPosition(
			previousRow.Start+previousRow.Length-1,
			previousRow.Rectangle.Width,
			previousRow.Rectangle.Y,
			e.Cursor.Row-1,
			previousRow.Length-1,
		)
	} else {
		e.Cursor.SetPosition(
			e.Cursor.CurrentIndex-1,
			e.Cursor.Rectangle.X-sequenceRectangle.X,
			e.Cursor.Rectangle.Y,
			e.Cursor.Row,
			e.Cursor.Column-1,
		)
	}
}
