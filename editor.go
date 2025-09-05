package main

// remember: moving cursor horizontally/mouse clicking/inserting/deleting invalidates the LastCursorPosition from rows

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CursorPosition{
// 	Position: rl.NewVector2(
// 		e.Rectangle.X,
// 		e.Rectangle.Y,
// 	),
// 	Row:          0,
// 	Column:       0,
// 	CurrentIndex: 0,
// },

const (
	UPWARD   = -1
	DOWNWARD = 1
)

// -1 indicates that the values are missing
type CursorPosition struct {
	Position                  rl.Vector2
	Row, Column, CurrentIndex int
}

// @line
type Row struct {
	Start, Length int
	Rectangle     rl.Rectangle
	AutoNewLine   bool
}

// @cursor
type Cursor struct {
	Rectangle    rl.Rectangle
	Row, Column  int
	CurrentIndex int
	TickTime     float32
	TickTimer    float32
	Color        rl.Color
}

func NewCursor(rectangle rl.Rectangle, row int, column int) Cursor {
	return Cursor{
		Rectangle:    rectangle,
		Row:          row,
		Column:       column,
		CurrentIndex: 0,
		TickTime:     0.5,
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
	Rectangle           rl.Rectangle
	BackgroundColor     rl.Color
	InFocus             bool
	Cursor              Cursor
	Rows                []*Row
	LastCursorPositions map[int]CursorPosition
	PieceTable          *PieceTable
	Font                *rl.Font
	FontSize            int
	FontColor           rl.Color
	PreviousCharacter   rune
	LastRowVisited      int
	Actions             []Action
}

func NewEditor(rectangle rl.Rectangle, backgroundColor rl.Color) Editor {
	pieceTable := NewPieceTable(Sequence{})
	fontSize := 30
	return Editor{
		Rectangle:           rectangle,
		BackgroundColor:     backgroundColor,
		PieceTable:          &pieceTable,
		FontSize:            fontSize,
		FontColor:           rl.White,
		Actions:             []Action{},
		InFocus:             false,
		Cursor:              NewCursor(rl.NewRectangle(rectangle.X, rectangle.Y, 3, float32(fontSize)), 0, 0),
		Rows:                make([]*Row, 0),
		LastCursorPositions: make(map[int]CursorPosition),
		CharRecCache:        make(map[rune]rl.Vector2),
	}
}

func (e *Editor) Index() int {
	currentRow := e.Rows[e.Cursor.Row]
	return currentRow.Start + e.Cursor.Column
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

func (e *Editor) FirstRow() *Row {
	return e.Rows[0]
}

func (e *Editor) LastRow() *Row {
	return e.Rows[len(e.Rows)-1]
}

func (e *Editor) PreviousRow() (*Row, error) {
	if e.Cursor.Row-1 < 0 {
		return nil, fmt.Errorf("PreviousRow: error trying to get previous row. current row - 1 < rows length")
	}
	return e.Rows[e.Cursor.Row-1], nil
}

func (e *Editor) CurrentRow() *Row {
	return e.Rows[e.Cursor.Row]
}

func (e *Editor) NextRow() (*Row, error) {
	if e.Cursor.Row+1 >= len(e.Rows) {
		return nil, fmt.Errorf("NextRow: error trying to get next row. current row + 1 > rows length")
	}
	return e.Rows[e.Cursor.Row+1], nil
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
	charSize := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0)
	if char == '\n' {
		charSize.Y = 30 // hardcoded for now, maybe the correct is to assign it the row's height mean
	}
	e.CharRecCache[char] = charSize
	return charSize
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
		false,
	}

	var lastWidth float32 = -1
	lastSpaceIndex := -1
	e.Rows = e.Rows[:0]
	length := 0
	text := e.PieceTable.ToString()
	var i = 0
	for i < len(text) {
		char := text[i]
		charSize := e.CharRectangle(rune(char))
		currentRow.Rectangle.Width += charSize.X
		length++
		// this serves to adjust the row height to the higher character found
		if currentRow.Rectangle.Height < charSize.Y {
			currentRow.Rectangle.Height = charSize.Y
		}

		charPosition := currentRow.Rectangle.X + currentRow.Rectangle.Width
		editorXBoundary := e.Rectangle.X + e.Rectangle.Width
		charOutOfEditorBounds := charPosition+charSize.X > editorXBoundary

		if char == ' ' {
			lastSpaceIndex = i
			lastWidth = currentRow.Rectangle.Width
		} else if char == '\n' {
			currentRow.Length = length
			length = 0
			e.Rows = append(e.Rows, currentRow)
			currentRow = &Row{
				i + 1,
				0,
				rl.NewRectangle(e.Rectangle.X, currentRow.Rectangle.Y+currentRow.Rectangle.Height, 0, 0),
				false,
			}
		} else if charOutOfEditorBounds {
			newRowStart := -1
			length = 0
			width := float32(0)
			if lastSpaceIndex == -1 || lastSpaceIndex < currentRow.Start {
				// if a space isn't found within a row then we will wrap at the character
				currentRow.Length = i - currentRow.Start
				currentRow.Rectangle.Width -= charSize.X
				currentRow.AutoNewLine = true
				e.Rows = append(e.Rows, currentRow)
				newRowStart = i // might be wrong, perhaps newRowStart = i+1
				length = 1
				width = charSize.X
				logger.Println("Width when splitting", currentRow.Rectangle.Width, charPosition, editorXBoundary)
			} else {
				// if a space is found within a row then we will wrap the whole word
				currentRow.Length = lastSpaceIndex - currentRow.Start + 1
				currentRow.Rectangle.Width = lastWidth
				currentRow.AutoNewLine = true
				e.Rows = append(e.Rows, currentRow)
				// spacePosition := lastSpaceIndex
				charAfterSpace := lastSpaceIndex + 1
				newRowStart = charAfterSpace
				i = lastSpaceIndex
			}
			currentRow = &Row{
				newRowStart,
				length,
				rl.NewRectangle(e.Rectangle.X, currentRow.Rectangle.Y+currentRow.Rectangle.Height, width, 0),
				false,
			}
			lastSpaceIndex = -1
		}
		i++
	}

	if length > 0 {
		currentRow.Length = length
		e.Rows = append(e.Rows, currentRow)
	}

	// ------------ Debugging ------------

	logger.Println("------------------------------------------------")
	for i, row := range e.Rows {
		logger.Println("Row: ", i, "| Row Start: ", row.Start, "| Row Length: ", row.Length, "| Row Start+Length: ", row.Start+row.Length, "| Row X: ", row.Rectangle.X, "| Row Y: ", row.Rectangle.Y, "| Row Width: ", row.Rectangle.Width, "| Row Height: ", row.Rectangle.Height)
	}
	logger.Println("Piece table PiecesAmount: ", e.PieceTable.PiecesAmount())

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
		charSize := e.CharRectangle(char)
		rl.DrawTextEx(*e.Font, stringChar, rl.NewVector2(charXPosition, currentRow.Rectangle.Y), float32(e.FontSize), 0, e.FontColor)
		charXPosition += charSize.X
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
	// if e.InFocus {
	e.Cursor.Draw()
	// }
}

func (e *Editor) SetFontSize(fontSize int) {
	e.FontSize = fontSize
	e.Cursor.Rectangle.Height = float32(fontSize)
}

func (e *Editor) FindPositionByRowColumn(row int, column int) float32 {
	rowToSearch := e.Rows[row]
	sequence, _, err := e.PieceTable.GetSequence(uint(rowToSearch.Start), uint(rowToSearch.Length))

	if err != nil {
		return -1
	}
	var width float32 = 0
	for i, char := range sequence {
		if i == column {
			break
		}
		charSize := e.CharRectangle(char)
		width += charSize.X
	}
	return width
}

// returns: rowIndex, row, inXBounds, column, index, columnXPosition, previousChar, error
func (e *Editor) FindRowClickMetadata(mouseClick rl.Vector2) (int, *Row, bool, int, int, float32, rune, error) {
	for i, row := range e.Rows {
		inRowXBoundaries := mouseClick.X >= row.Rectangle.X && mouseClick.X <= row.Rectangle.X+row.Rectangle.Width
		inRowYBoundaries := mouseClick.Y >= row.Rectangle.Y && mouseClick.Y <= row.Rectangle.Y+row.Rectangle.Height
		if inRowXBoundaries && inRowYBoundaries {
			sequence, _, err := e.PieceTable.GetSequence(uint(row.Start), uint(row.Length))
			if err != nil {
				return -1, nil, false, -1, -1, -1, -1, err
			}
			var previousCharacter rune
			var previousCharacterX float32
			column := 0
			currentIndex := row.Start
			charXPosition := row.Rectangle.X
			for _, char := range sequence {
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
			return i, row, true, column, currentIndex, charXPosition, previousCharacter, nil
		}
		// if it isn't on row X boundaries, it makes no sense to search the metadata
		if inRowYBoundaries {
			index := row.Start + row.Length
			column := row.Length
			if !row.AutoNewLine {
				index--
				column--
			}
			previousChar, _ := e.PieceTable.GetAt(uint(index - 1))
			return i, row, false, column, index, row.Rectangle.X + row.Rectangle.Width, previousChar, nil
		}
	}
	return -1, nil, false, -1, -1, -1, -1, nil
}

func (e *Editor) MoveCursorForward() {
	if e.Cursor.CurrentIndex == int(e.PieceTable.Length)-1 {
		return
	}
	clear(e.LastCursorPositions)
	currentChar, _ := e.CurrentChar()
	currentRow := e.CurrentRow()
	charSize := e.CharRectangle(currentChar)
	isEndOfRow := e.Cursor.Column == currentRow.Length
	isNewLine := currentChar == '\n'
	isCharacter := !isEndOfRow && !isNewLine
	if isCharacter {
		e.PreviousCharacter = currentChar
		e.Cursor.SetPosition(
			e.Cursor.CurrentIndex+1,
			e.Cursor.Rectangle.X+charSize.X,
			currentRow.Rectangle.Y,
			e.Cursor.Row,
			e.Cursor.Column+1,
		)
	} else {
		e.LastRowVisited = e.Cursor.Row
		e.Cursor.Column = 0
		e.Cursor.Rectangle.X = e.Rectangle.X
		if isNewLine {
			e.Cursor.CurrentIndex++
		}
		nextRow, _ := e.NextRow()
		e.Cursor.Rectangle.Y = nextRow.Rectangle.Y
		e.Cursor.Row++
	}
}

func (e *Editor) MoveCursorBackward() {
	if e.Cursor.CurrentIndex == 0 {
		return
	}
	clear(e.LastCursorPositions)
	currentChar, _ := e.CurrentChar()
	shouldGoToPreviousRow := e.Cursor.Column == 0 && e.Cursor.Row > 0
	if shouldGoToPreviousRow {
		e.LastRowVisited = e.Cursor.Row
		previousRow, _ := e.PreviousRow()
		newColumn := previousRow.Length
		newCurrentIndex := e.Cursor.CurrentIndex
		if !previousRow.AutoNewLine {
			newColumn--
			newCurrentIndex--
		}
		e.Cursor.SetPosition(
			newCurrentIndex,
			e.Rectangle.X+previousRow.Rectangle.Width,
			previousRow.Rectangle.Y,
			e.Cursor.Row-1,
			newColumn,
		)
	} else {
		previousCharacter, _ := e.PreviousChar()
		fmt.Println(string(previousCharacter))
		charSize := e.CharRectangle(previousCharacter)
		e.Cursor.Rectangle.X -= charSize.X
		e.Cursor.CurrentIndex--
		e.Cursor.Column--
	}
	e.PreviousCharacter = currentChar
}

func (e *Editor) _internalMoveCursorBackwardOrDownward(direction int) {
	e.LastRowVisited = e.Cursor.Row
	var row *Row
	var lastCursorPosition CursorPosition
	var ok bool
	var newCurrentIndex int
	var newRow int
	var shouldDecreaseColumnAndIndex bool = false

	if direction == UPWARD {
		row, _ = e.PreviousRow()
		lastCursorPosition, ok = e.LastCursorPositions[e.Cursor.Row-1]
		newRow = e.Cursor.Row - 1
		newCurrentIndex = e.Cursor.CurrentIndex - (e.Cursor.Column + row.Length - e.Cursor.Column)
		shouldDecreaseColumnAndIndex = e.Cursor.Row-1 == 0 || !row.AutoNewLine
	}
	if direction == DOWNWARD {
		row, _ = e.NextRow()
		lastCursorPosition, ok = e.LastCursorPositions[e.Cursor.Row+1]
		newRow = e.Cursor.Row + 1
		newCurrentIndex = e.Cursor.CurrentIndex + (e.CurrentRow().Length - e.Cursor.Column + e.Cursor.Column)
		shouldDecreaseColumnAndIndex = !row.AutoNewLine
	}
	if ok {
		e.Cursor.SetPosition(
			lastCursorPosition.CurrentIndex,
			lastCursorPosition.Position.X,
			lastCursorPosition.Position.Y,
			newRow,
			lastCursorPosition.Column,
		)
	} else if e.Cursor.Column >= row.Length {
		e.LastCursorPositions[e.Cursor.Row] = CursorPosition{
			Position:     rl.Vector2{X: e.Cursor.Rectangle.X, Y: e.Cursor.Rectangle.Y},
			Row:          e.Cursor.Row,
			Column:       e.Cursor.Column,
			CurrentIndex: e.Cursor.CurrentIndex,
		}
		newColumn := row.Length
		newCurrentIndex := row.Start + row.Length
		if shouldDecreaseColumnAndIndex {
			newColumn--
			newCurrentIndex -= 1
		}
		e.Cursor.SetPosition(
			newCurrentIndex,
			row.Rectangle.X+row.Rectangle.Width,
			row.Rectangle.Y,
			newRow,
			newColumn,
		)
	} else {
		newX := e.FindPositionByRowColumn(newRow, e.Cursor.Column)
		e.Cursor.SetPosition(
			newCurrentIndex,
			newX,
			row.Rectangle.Y,
			newRow,
			e.Cursor.Column,
		)
	}
}

func (e *Editor) MoveCursorUpward() {
	if e.Cursor.Row == 0 {
		return
	}
	e._internalMoveCursorBackwardOrDownward(UPWARD)
}

func (e *Editor) MoveCursorDownward() {
	if e.Cursor.Row >= len(e.Rows)-1 {
		return
	}
	e._internalMoveCursorBackwardOrDownward(DOWNWARD)
}

func (e *Editor) SetCursorPositionByClick(mouseClick rl.Vector2) error {
	rowIndex, row, _, column, index, xPosition, previousChar, err := e.FindRowClickMetadata(mouseClick)
	if err != nil {
		return err
	}
	e.LastRowVisited = e.Cursor.Row
	clear(e.LastCursorPositions)
	if row != nil {
		e.Cursor.SetPosition(
			index,
			xPosition,
			row.Rectangle.Y,
			rowIndex,
			column,
		)
		e.PreviousCharacter = previousChar
	}
	if row == nil {
		// TODO: it needs to be within editor boundaries, if not, return
		lastRow := e.LastRow()
		index := lastRow.Start + lastRow.Length
		column := lastRow.Length
		if !lastRow.AutoNewLine {
			index--
			column--
		}
		previousChar, _ := e.PieceTable.GetAt(uint(index - 1))
		e.PreviousCharacter = previousChar
		xPosition := lastRow.Rectangle.X + lastRow.Rectangle.Width
		e.Cursor.SetPosition(
			index,
			xPosition,
			lastRow.Rectangle.Y,
			len(e.Rows)-1,
			column,
		)
	}
	return nil
}

// func (e *Editor) FindRowFromPosition(mouseClickPosition rl.Vector2) (*Row, int) {
// 	for index, row := range e.Rows {
// 		if mouseClickPosition.Y > row.Rectangle.Y && mouseClickPosition.Y < row.Rectangle.Y+row.Rectangle.Height {
// 			return row, index
// 		}
// 	}
// 	return nil, 1
// }

// func (e *Editor) MoveCursorForward() {
// 	currentChar, _ := e.CurrentChar()
// 	currentRow := e.Rows[e.Cursor.Row]
// 	isEndOfRow := e.Cursor.Column >= currentRow.Length || e.Cursor.Column >= currentRow.Length && e.LastAction() == TYPING
// 	isNewLine := currentChar == '\n'
// 	charSize := e.CharRectangle(currentChar)
// 	isCharacter := !isNewLine && !isEndOfRow
// 	if isCharacter {
// 		e.PreviousCharacter = currentChar
// 		e.Cursor.SetPosition(
// 			e.Cursor.CurrentIndex+1,
// 			e.Cursor.Rectangle.X+charSize.X,
// 			currentRow.Rectangle.Y,
// 			e.Cursor.Row,
// 			e.Cursor.Column+1,
// 		)
// 		// e.Cursor.Column++
// 		// e.Cursor.Rectangle.X += charSize.X
// 		// e.Cursor.CurrentIndex++
// 	} else {
// 		e.Cursor.Column = 0
// 		e.Cursor.Rectangle.X = e.Rectangle.X
// 		if isNewLine {
// 			e.Cursor.CurrentIndex++
// 		}
// 		if e.Cursor.Row < len(e.Rows) {
// 			nextRow, _ := e.NextRow()
// 			e.Cursor.Rectangle.Y = nextRow.Rectangle.Y
// 			e.LastRowVisited = e.Cursor.Row
// 			e.Cursor.Row++
// 		}
// 		// if e.LastAction() == TYPING {
// 		// 	e.MoveCursorForward()
// 		// }
// 	}
// }

// func (e *Editor) MoveCursorBackward() {
// 	previousCharacter, _ := e.PreviousChar()
// 	previousRow := &Row{Start: -1}
// 	if e.Cursor.Row > 0 {
// 		previousRow, _ = e.PreviousRow()
// 	}
// 	e.Cursor.Column--
// 	shouldGoToPreviousRow := previousRow.Start != -1 && e.Cursor.CurrentIndex == previousRow.Start+previousRow.Length
// 	// shouldGoToPreviousRow := previousRow.Start != -1 && e.Cursor.Column == previousRow.Start+previousRow.Length
// 	if shouldGoToPreviousRow && e.Cursor.Row > 0 {
// 		previousRow, _ := e.PreviousRow()
// 		e.LastRowVisited = e.Cursor.Row

// 		newColumn := previousRow.Length
// 		if e.Cursor.Row-1 > 0 {
// 			newColumn -= 1
// 		}
// 		e.Cursor.SetPosition(
// 			e.Cursor.CurrentIndex,
// 			e.Rectangle.X + previousRow.Rectangle.Width,
// 			previousRow.Rectangle.Y,
// 			e.Cursor.Row-1,
// 			newColumn,
// 		)

// 		isPreviousIndexFromPreviousRow := e.Cursor.CurrentIndex-1 >= previousRow.Start && e.Cursor.CurrentIndex-1 < previousRow.Start+previousRow.Length
// 		isNewLineFromPreviousRow := previousCharacter == '\n'
// 		if isPreviousIndexFromPreviousRow && isNewLineFromPreviousRow {
// 			e.Cursor.CurrentIndex--
// 		}
// 	} else {
// 		charSize := e.CharRectangle(e.PreviousCharacter)
// 		e.Cursor.Rectangle.X -= charSize.X
// 		e.Cursor.CurrentIndex--
// 	}
// }

// func (e *Editor) SetCursorPositionFromIndex(newPosition int) {
// 	currentRow := e.Rows[e.Cursor.Row]
// 	inCurrentRowBoundaries := newPosition >= currentRow.Start && newPosition <= currentRow.Start+currentRow.Length-1
// 	if inCurrentRowBoundaries {
// 		if newPosition > e.Cursor.CurrentIndex {
// 			iterations := newPosition - e.Cursor.CurrentIndex
// 			for range iterations {
// 				e.MoveCursorForward()
// 			}
// 		} else if newPosition < e.Cursor.CurrentIndex {
// 			iterations := e.Cursor.CurrentIndex - newPosition
// 			for range iterations {
// 				e.MoveCursorBackward()
// 			}
// 		}
// 	} else {
// 		for _, row := range e.Rows {
// 			foundRow := newPosition >= row.Start && newPosition <= row.Start+row.Length-1 || e.Cursor.Column == 1
// 			if foundRow {
// 				if newPosition > e.Cursor.CurrentIndex {
// 					iterations := newPosition - e.Cursor.CurrentIndex
// 					for range iterations {
// 						e.MoveCursorForward()
// 					}
// 				}
// 				if newPosition < e.Cursor.CurrentIndex {
// 					iterations := e.Cursor.CurrentIndex - newPosition
// 					for range iterations {
// 						e.MoveCursorBackward()
// 					}
// 				}
// 				return
// 			}
// 		}
// 	}
// }

// func (e *Editor) FindRowFromClick(mouseClickPosition rl.Vector2) (*Row, int) {
// 	for i, row := range e.Rows {
// 		if mouseClickPosition.Y > row.Rectangle.Y && mouseClickPosition.Y < row.Rectangle.Y+row.Rectangle.Height {
// 			return row, i
// 		}
// 	}
// 	return nil, -1
// }

// func (e *Editor) SetCursorPositionFromClickRow(rowIndex int, mouseClickPosition rl.Vector2) {
// 	row := e.Rows[rowIndex]
// 	if row != nil {
// 		sequence, _, err := e.PieceTable.GetSequence(uint(row.Start), uint(row.Length))
// 		if err != nil {
// 			return
// 		}
// 		var previousCharacter rune
// 		var previousCharacterX float32
// 		column := 0
// 		currentIndex := row.Start
// 		charXPosition := row.Rectangle.X
// 		iterableText := string(sequence)
// 		logger.Println("Row text:", iterableText, "Len:", len(iterableText))
// 		for i, char := range sequence {
// 			if char == '\n' {
// 				continue
// 			}
// 			charSize := e.CharRectangle(char)
// 			betweenPostPreviousCharHalfAndPreCharHalf := mouseClickPosition.X > previousCharacterX && mouseClickPosition.X < charXPosition+(charSize.X/2)
// 			if betweenPostPreviousCharHalfAndPreCharHalf || char == '\n' && i == len(sequence) {
// 				break
// 			}
// 			previousCharacter = char
// 			previousCharacterX = charXPosition
// 			charXPosition += charSize.X
// 			currentIndex++
// 			column++
// 		}
// 		e.Cursor.SetPosition(currentIndex, charXPosition, row.Rectangle.Y, rowIndex, column)
// 		e.PreviousCharacter = previousCharacter
// 	}
// }

// func (e *Editor) SetCursorPositionFromClick(mouseClickPosition rl.Vector2) {
// 	_, rowIndex := e.FindRowFromClick(mouseClickPosition)
// 	if rowIndex >= 0 {
// 		e.SetCursorPositionFromClickRow(rowIndex, mouseClickPosition)
// 	}
// }

// func (e *Editor) SetCursorToPreviousRow(column int) {
// 	if e.Cursor.Row > 0 {

// 	}
// }

// func (e *Editor) Insert(index int, text Sequence) {
// 	e.AddAction(TYPING)
// 	if len(text) == 0 {
// 		return
// 	}
// 	sequenceFirstChar := text[0]
// 	if sequenceFirstChar == '\n' {
// 		e.PieceTable.Insert(uint(index), []rune{text[0]})
// 		text = text[1:]
// 	}
// 	if len(text) > 0 {
// 		e.PieceTable.Insert(uint(index), text)
// 	}
// 	// before calc
// 	// rowsBeforeCalc := len(e.Rows)
// 	currentRowBeforeCalc := e.Rows[e.Cursor.Row]
// 	var nextRowBeforeCalc *Row = nil
// 	if e.Cursor.Row < len(e.Rows)-1 {
// 		nextRowBeforeCalc, _ = e.NextRow()
// 	}
// 	e.CalculateRows()
// 	// if len(text) < 1 {
// 	// 	return
// 	// }
// 	var nextRowAfterCalc *Row = nil
// 	if e.Cursor.Row < len(e.Rows)-1 {
// 		nextRowAfterCalc, _ = e.NextRow()
// 	}
// 	currentRowAfterCalc := e.Rows[e.Cursor.Row]
// 	sequenceRectangle := e.SequenceRectangle(text)

// 	if nextRowBeforeCalc != nil && nextRowAfterCalc != nil && nextRowBeforeCalc.Length < nextRowAfterCalc.Length && e.Cursor.Column >= currentRowAfterCalc.Length {
// 		// when you are inserting at the end of a row and the character is being added to the next row, that means
// 		// the word is being wrapped, so every character will end up on the next row
// 		e.Cursor.Row++
// 		e.Cursor.SetPosition(
// 			nextRowAfterCalc.Start+1,
// 			nextRowAfterCalc.Rectangle.X+sequenceRectangle.X,
// 			nextRowAfterCalc.Rectangle.Y,
// 			e.Cursor.Row,
// 			1,
// 		)
// 	} else if (currentRowAfterCalc.Length < currentRowBeforeCalc.Length && e.Cursor.Column > currentRowAfterCalc.Start+currentRowAfterCalc.Length) || sequenceFirstChar == '\n' {
// 		// to me, it seems reasonable to think that:
// 		// if the current row decreased it's Length, then a word has been wrapped
// 		e.Cursor.Row++
// 		currentRow := e.Rows[e.Cursor.Row]
// 		e.Cursor.SetPosition(
// 			currentRow.Start+currentRow.Length-1,
// 			currentRow.Rectangle.Width,
// 			currentRow.Rectangle.Y,
// 			e.Cursor.Row,
// 			currentRow.Length-1,
// 		)
// 	} else {
// 		e.Cursor.SetPosition(
// 			e.Cursor.CurrentIndex+1,
// 			e.Cursor.Rectangle.X+sequenceRectangle.X,
// 			e.Cursor.Rectangle.Y,
// 			e.Cursor.Row,
// 			e.Cursor.Column+1,
// 		)
// 	}
// }

// func (e *Editor) Delete(index int, length int) error {
// 	e.AddAction(DELETE)
// 	// TODO: Add suport to multichar deletion
// 	// things to keep in mind:
// 	// - Multichar deletion may affect another row, maybe more than 2
// 	// - Multichar deletion may only occur when there's text selected
// 	// 	 So MAYBE where the cursor stop is where it will end up
// 	sequence, _, err := e.PieceTable.GetSequence(uint(e.Cursor.CurrentIndex)-1, uint(length))
// 	if err != nil {
// 		return err
// 	}
// 	sequenceRectangle := e.SequenceRectangle(sequence)
// 	e.PieceTable.Delete(uint(index), uint(length))
// 	rowBeforeCalc := e.Rows[e.Cursor.Row]
// 	var columnBeforeCalc int
// 	if e.Cursor.Row > 0 {
// 		columnBeforeCalc = rowBeforeCalc.Length
// 		if e.Cursor.Row-1 > 0 {
// 			columnBeforeCalc -= 1
// 		}
// 	}
// 	e.CalculateRows()
// 	if e.Cursor.Row > 0 && e.Cursor.Column == 0 {
// 		// send cursor to last char of the previous row
// 		previousRow, _ := e.PreviousRow()
// 		// newCurrentIndex := previousRow.Start+previousRow.Length
// 		// newColumn := previousRow.Length
// 		// if e.Cursor.Row-1 > 0 {
// 		// 	newColumn -= 1
// 		// 	newCurrentIndex -= 1
// 		// }
// 		e.LastRowVisited = e.Cursor.Row
// 		e.Cursor.SetPosition(
// 			e.Cursor.CurrentIndex-1,
// 			rowBeforeCalc.Rectangle.Width,
// 			previousRow.Rectangle.Y,
// 			e.Cursor.Row-1,
// 			columnBeforeCalc-1,
// 		)
// 	} else {
// 		e.Cursor.SetPosition(
// 			e.Cursor.CurrentIndex-1,
// 			e.Cursor.Rectangle.X-sequenceRectangle.X,
// 			e.Cursor.Rectangle.Y,
// 			e.Cursor.Row,
// 			e.Cursor.Column-1,
// 		)
// 	}
// 	return nil
// }
