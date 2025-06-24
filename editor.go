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
	// X and Y should be based on row and column
	if time.Since(c.LastTick).Milliseconds() >= 0 && time.Since(c.LastTick).Milliseconds() <= 500 {
		rl.DrawRectangle(
			c.Rectangle.ToInt32().X,
			c.Rectangle.ToInt32().Y,
			c.Rectangle.ToInt32().Width,
			c.Rectangle.ToInt32().Height,
			c.Color,
		)
	}
}

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
	FontColor					rl.Color
	PreviousCharacter rune
}

func NewEditor(rectangle rl.Rectangle, backgroundColor rl.Color) Editor {
	pieceTable := NewPieceTable(Sequence{})
	fontSize := 30
	return Editor{
		Rectangle:       rectangle,
		BackgroundColor: backgroundColor,
		InFocus:         false,
		Cursor:          NewCursor(rl.NewRectangle(rectangle.X, rectangle.Y, 3, float32(fontSize)), 0, 0),
		PieceTable:      &pieceTable,
		FontSize:        fontSize,
		FontColor: 			 rl.White,
		Rows:            make([]*Row, 0),
		CharRecCache:    make(map[rune]rl.Vector2),
	}
}

func (e *Editor) GetCharRectangle(char rune) rl.Vector2 {
	fromCache, ok := e.CharRecCache[char]
	if ok {
		return fromCache
	} else {
		charRec := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0) 
		e.CharRecCache[char] = charRec
		return charRec
	}
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

		char := text[i]
		length++
		textSizeVector2 := e.GetCharRectangle(rune(char))
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
			charOutOfEditorBounds := currentRow.Rectangle.X+currentRow.Rectangle.Width+textSizeVector2.X >= e.Rectangle.X+e.Rectangle.Width
			if charOutOfEditorBounds {
				length = 0
				newRowStart := -1
				if lastSpaceIndex == -1 || lastSpaceIndex < currentRow.Start {
					// if a space isn't found within a row then we will wrap at the character
					currentRow.Length = i - currentRow.Start
					e.Rows = append(e.Rows, currentRow)
					newRowStart = i
					length = 1
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
					rl.NewRectangle(e.Rectangle.X, currentRow.Rectangle.Y+currentRow.Rectangle.Height, 0, 0),
					rl.NewVector2(0, 0),
				}
				lastSpaceIndex = -1
			}

			if currentRow.Rectangle.Height < textSizeVector2.Y {
				// this serves to adjust the row height to the higher character found
				currentRow.Rectangle.Height = textSizeVector2.Y
			}
		}
	}

	// append the final row if it needs to, otherwise it might not be added
	if length > 0 {
		currentRow.Length = length
		e.Rows = append(e.Rows, currentRow)
	}
}

func (e *Editor) DrawText() {
	totalRowChars := 0
	text := e.PieceTable.ToString()
	for _, row := range e.Rows {
		totalRowChars += row.Length
		// logger.Println("Row", row, len(text))
		// logger.Println(text[row.Start:row.Length])
	}
	if totalRowChars < len(text) {
		logger.Println("Warning: more characters than space in e.Rows", totalRowChars, len(text))
	}

	currentRowIndex := 0
	currentRow := e.Rows[currentRowIndex]
	charXPosition := currentRow.Rectangle.X
	length := 0
	for _, char := range e.PieceTable.ToString() {
		if length == currentRow.Length {
			currentRowIndex++
			if currentRowIndex < len(e.Rows) {
				currentRow = e.Rows[currentRowIndex]
			}
			length = 0
			charXPosition = currentRow.Rectangle.X
		}
		length++
		stringChar := string(char)
		textSizeVector2 := e.GetCharRectangle(char)
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
	char, err := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex + 1))
	logger.Println("Current char", string(currentChar), "Next char: ", string(char))
	if err != nil {
		return
	}
	currentRow := e.Rows[e.Cursor.Row]
	isEndOfRow := e.Cursor.CurrentIndex == currentRow.Start+currentRow.Length
	isNewLine := currentChar == '\n'
	logger.Println(isEndOfRow, isNewLine)

	charToMeasure := currentChar
	// if isNewLine {
	// 	charToMeasure = char
	// }
	charRec := e.GetCharRectangle(charToMeasure)
	if isNewLine && e.Cursor.Column > 0 {
		logger.Println("isNewLine")
		e.Cursor.CurrentIndex++
		e.Cursor.Column = 0
		e.Cursor.Rectangle.X = e.Rectangle.X
		if e.Cursor.Row != len(e.Rows)-1 {
			nextRow := e.Rows[e.Cursor.Row+1]
			e.Cursor.Rectangle.Y = nextRow.Rectangle.Y
			e.Cursor.Row++
		}
	} else if !isEndOfRow {
		e.Cursor.Column++
		e.Cursor.Rectangle.X += charRec.X
		e.PreviousCharacter = currentChar
		e.Cursor.CurrentIndex++
	} else {
		e.Cursor.Column = 0
		e.Cursor.Rectangle.X = e.Rectangle.X
		// Send Cursor to the next Row
		if e.Cursor.Row != len(e.Rows)-1 {
			nextRow := e.Rows[e.Cursor.Row+1]
			e.Cursor.Rectangle.Y = nextRow.Rectangle.Y
			e.Cursor.Row++
		}
	}
}

func (e *Editor) MoveCursorBackward() {
	previousCharacter, _ := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex - 1))
	// char, err := e.PieceTable.GetAt(uint(e.Cursor.CurrentIndex))

	// if err != nil {
	// 	return
	// }1
	e.Cursor.Column--

	previousRow := &Row{
		Start: -1,
	}
	if e.Cursor.Row > 0 {
		previousRow = e.Rows[e.Cursor.Row-1]
	}

	shouldGoBackToPreviousRow := previousRow.Start != -1 && e.Cursor.CurrentIndex == previousRow.Start+previousRow.Length
	// charToMeasure := char
	// if !shouldGoBackToPreviousRow {
	// }
	charToMeasure := e.PreviousCharacter
	if shouldGoBackToPreviousRow && e.Cursor.Row > 0 {
		previousRow := e.Rows[e.Cursor.Row-1]
		e.Cursor.Rectangle.Y = previousRow.Rectangle.Y
		e.Cursor.Rectangle.X = e.Rectangle.X + previousRow.Rectangle.Width
		e.Cursor.Column = previousRow.Length
		e.Cursor.Row--
		isFromPreviousRow := e.Cursor.CurrentIndex-1 > previousRow.Start && e.Cursor.CurrentIndex-1 < previousRow.Start+previousRow.Length
		isNewLineFromPreviousRow := previousCharacter == '\n'
		if isFromPreviousRow && isNewLineFromPreviousRow {
			e.Cursor.CurrentIndex--
		}
	} else {
		charRec := e.GetCharRectangle(charToMeasure)
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
		}
		if newPosition < e.Cursor.CurrentIndex {
			iterations := e.Cursor.CurrentIndex - newPosition
			for range iterations {
				e.MoveCursorBackward()
			}
		}
	} else {
		for _, row := range e.Rows {
			foundRow := newPosition >= row.Start && newPosition <= row.Start+row.Length-1 || e.Cursor.Column == 1
			logger.Println(row.Start, row.Start+row.Length-1, foundRow)
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

func (e *Editor) SetCursorPositionFromClick(mouseClickPosition rl.Vector2) {
	var foundRow *Row = nil
	rowIndex := 0
	for i, row := range e.Rows {
		if mouseClickPosition.Y > row.Rectangle.Y && mouseClickPosition.Y < row.Rectangle.Y+row.Rectangle.Height {
			foundRow = row
			rowIndex = i
		}
	}
	if foundRow != nil {
		text := e.PieceTable.ToString()
		iterableText := text[foundRow.Start : foundRow.Start+foundRow.Length]
		charXPosition := foundRow.Rectangle.X
		currentIndex := foundRow.Start
		// wasCharXFound := false
		column := 0
		var previousCharacter rune
		var previousCharacterX float32
		for _, char := range iterableText {
			if char == '\n' {
				continue
			}
			column++
			charRec := e.GetCharRectangle(char)
			betweenPostPreviousCharHalfAndPreHalfChar := mouseClickPosition.X > previousCharacterX && mouseClickPosition.X < charXPosition+(charRec.X/2)
			if betweenPostPreviousCharHalfAndPreHalfChar {
				break
			}
			previousCharacter = char
			previousCharacterX = charXPosition
			charXPosition += charRec.X
			currentIndex++
		}
		e.Cursor.SetPosition(currentIndex, charXPosition, foundRow.Rectangle.Y, rowIndex, column)
		e.PreviousCharacter = previousCharacter
	}
}

func (e *Editor) Insert(index int, text Sequence) {
	e.PieceTable.Insert(uint(index), text)
	e.CalculateRows()
	e.SetCursorPositionFromIndex(index+1)
}

func (e *Editor) Delete(index int, text Sequence) {
	e.PieceTable.Delete(uint(index), 1)
	e.CalculateRows()
	e.SetCursorPositionFromIndex(index-1)
}