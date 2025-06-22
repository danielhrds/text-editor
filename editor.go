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
	Rectangle   rl.Rectangle
	Row, Column int
	LastTick    time.Time
}

func NewCursor(rectangle rl.Rectangle, row int, column int) Cursor {
	return Cursor{
		Rectangle: rectangle,
		Row:       row,
		Column:    column,
		LastTick:  time.Now(),
	}
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
			rl.Black,
		)
	}
}

// @editor
type Editor struct {
	Rectangle         rl.Rectangle
	BackgroundColor   rl.Color
	InFocus           bool
	Cursor            Cursor
	Rows              []*Row
	CurrentPosition   int
	PieceTable        *PieceTable
	Font              *rl.Font
	FontSize          int
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
		Rows:            make([]*Row, 0),
		CurrentPosition: -1,
	}
}

func (e *Editor) DrawText() {
	currentRow := &Row{
		0,
		0,
		rl.NewRectangle(e.Rectangle.X, e.Rectangle.Y, 0, 0),
		rl.NewVector2(0, 0),
	}

	length := 0
	newRow := false

	e.Rows = e.Rows[:0]
	for i, char := range e.PieceTable.ToString() {
		strChar := string(char)
		if newRow {
			previousRow := e.Rows[len(e.Rows)-1]
			currentRow = &Row{
				i,
				0,
				rl.NewRectangle(e.Rectangle.X, previousRow.Rectangle.Y+previousRow.Rectangle.Height, 0, 0),
				rl.NewVector2(0, 0),
			}
			newRow = false
		}

		length++
		if char == '\n' {
			currentRow.Length = length
			newRow = true
			length = 0
			// logger.Printf("Row address %p %d %d", currentRow, currentRow.Start, currentRow.Length)
			e.Rows = append(e.Rows, currentRow)
		}
		// if char != '\n' && i == int(e.PieceTable.PiecesAmount()-1) {
		// 	currentRow.Length = length
		// 	e.Rows = append(e.Rows, currentRow)
		// }

		textRec := rl.MeasureTextEx(*e.Font, strChar, float32(e.FontSize), 0)
		if char != '\n' && currentRow.Rectangle.Height < textRec.Y {
			currentRow.Rectangle.Height = textRec.Y
		}
		rl.DrawTextEx(*e.Font, strChar, rl.NewVector2(currentRow.Rectangle.X+currentRow.Rectangle.Width, currentRow.Rectangle.Y), float32(e.FontSize), 0, rl.Black)
		currentRow.Rectangle.Width += textRec.X
	}
	// logger.Println(e.Rows)
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

func (e *Editor) MoveCursorForward() {
	char, err := e.PieceTable.GetAt(uint(e.CurrentPosition+1))
	logger.Println(string(char))
	if err != nil {
		return
	}
	isNewLine := char == '\n'
	// if e.Cursor.Row != 0 || e.Cursor.Column != 0 {
	// }
	e.CurrentPosition++
	if !isNewLine {
		charRec := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0)
		e.Cursor.Column++
		e.Cursor.Rectangle.X += charRec.X
		e.PreviousCharacter = char
	} else {
		e.Cursor.Column = 0
		e.Cursor.Rectangle.X = e.Rectangle.X

		currentRowIndex := e.Cursor.Row
		// Send Cursor to the next Row
		if currentRowIndex != len(e.Rows)-1 {
			nextRow := e.Rows[currentRowIndex+1]
			e.Cursor.Rectangle.Y = nextRow.Rectangle.Y
			e.Cursor.Row++
		}
	}
}

func (e *Editor) MoveCursorBackward() {
	char, err := e.PieceTable.GetAt(uint(e.CurrentPosition))
	logger.Println(string(char))
	if err != nil {
		return
	}
	e.Cursor.Column--
	// if e.Cursor.Row > 0 && e.Cursor.Column == 0 {
	// }
	// e.CurrentPosition--
	isNewLine := char == '\n'
	if !isNewLine {
		charRec := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0)
		e.Cursor.Rectangle.X -= charRec.X
		e.CurrentPosition--
	} else {
		currentRowIndex := e.Cursor.Row
		currentRow := e.Rows[currentRowIndex]
		isTheSameRow := e.CurrentPosition >= currentRow.Start && e.CurrentPosition <= currentRow.Start+currentRow.Length-1
		// Send Cursor to the previous Row
		if currentRowIndex != 0 && !isTheSameRow {
			previousRow := e.Rows[currentRowIndex-1]
			e.Cursor.Rectangle.Y = previousRow.Rectangle.Y
			e.Cursor.Rectangle.X = e.Rectangle.X + previousRow.Rectangle.Width
			e.Cursor.Column = previousRow.Length-1
			e.CurrentPosition--
			e.Cursor.Row--
		}
		if isTheSameRow {
			charRec := rl.MeasureTextEx(*e.Font, string(e.PreviousCharacter), float32(e.FontSize), 0)
			e.Cursor.Rectangle.X -= charRec.X
		}
	}
}

func (e *Editor) SetCursorPosition(newPosition int) {
	currentRow := e.Rows[e.Cursor.Row]
	inRowBoundaries := newPosition >= currentRow.Start && newPosition <= currentRow.Start+currentRow.Length-1
	if inRowBoundaries {
		if newPosition > e.CurrentPosition {
			iterations := newPosition - e.CurrentPosition
			for range iterations {
				e.MoveCursorForward()
			}
		}
		if newPosition < e.CurrentPosition {
			iterations := e.CurrentPosition - newPosition
			for range iterations {
				e.MoveCursorBackward()
			}
		}
	} else {
		for _, row := range e.Rows {
			foundRow := newPosition >= row.Start && newPosition <= row.Start+row.Length-1 || e.Cursor.Column == 1
			if foundRow {
				// e.Cursor.Column = 0
				// e.Cursor.Row = rowIndex
				// e.Cursor.Rectangle.X = e.Rectangle.X
				// e.Cursor.Rectangle.Y = row.Rectangle.Y
				if newPosition > e.CurrentPosition {
					iterations := newPosition - e.CurrentPosition
					for range iterations {
						e.MoveCursorForward()
					}
				}
				if newPosition < e.CurrentPosition {
					iterations := e.CurrentPosition - newPosition
					for range iterations {
						e.MoveCursorBackward()
					}
				}
				// e.Cursor.Row = rowIndex
				break
			}
		}
	}

	// currentPosition := e.CurrentPosition
	// char, err := e.PieceTable.GetAt(uint(currentPosition))
	// charRec := rl.MeasureTextEx(*e.Font, string(char), float32(e.FontSize), 0)
	// if char != '\n' {
	// 	if err == nil {
	// 		e.Cursor.Column++
	// 		e.CurrentPosition++
	// 		e.Cursor.Rectangle.X += charRec.X
	// 	}
	// } else {
	// 	e.Cursor.Column = 0
	// 	e.Cursor.Rectangle.X = e.Rectangle.X

	// 	currentRowIndex := e.Cursor.Row
	// 	currentRow := e.Rows[currentRowIndex]
	// 	e.Cursor.Rectangle.Y = currentRow.Rectangle.Y + currentRow.Rectangle.Height
	// 	e.Cursor.Row++
	// }
}
