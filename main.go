// TODO:
// - Cache
// - Refactor piece table delete function

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// @main
func main() {
	defer rl.CloseWindow()
	window := NewWindow(60, 1600, 900)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(window.Width, window.Height, "Text Editor")
	rl.SetWindowState(rl.FlagWindowAlwaysRun)
	rl.SetTargetFPS(window.FPS)

	original, _ := ReadFile("example.txt")
	pt := NewPieceTable(
		Sequence(original),
	)
	pt.Insert(20, Sequence("went to the park and\n"))
	logger.Println(pt.ToString())

	// original, _ := ReadFile("example2.txt")
	// original, _ := ReadFile("example3.txt")
	// pt := NewPieceTable(
	// 	Sequence(original),
	// )
	// editor := NewEditor(rl.NewRectangle(20, 0, pfloat32(window.Width-100), float32(window.Height-100)), rl.Gray)
	editor := NewEditor(rl.NewRectangle(20, 0, 250, float32(window.Height-100)), rl.Gray)
	font := rl.LoadFontEx("fonts/JetBrainsMono-Regular.ttf", int32(editor.FontSize), nil, 0)
	defer rl.UnloadFont(font)
	editor.Font = &font
	editor.PieceTable = &pt
	window.Editor = &editor
	window.Editor.InFocus = true
	window.Editor.CalculateRows()
	// window.Editor.SetCursorPosition(18)

	for !rl.WindowShouldClose() {
		rl.ClearBackground(rl.White)
		rl.BeginDrawing()
		window.Draw()
		window.Input()
		rl.EndDrawing()
	}
}

func ReadFile(path string) (string, int) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	str := ""
	i := 1
	for scanner.Scan() {
		str += scanner.Text() + "\n"
		i++
	}
	return str, i
}

// @window
type Component interface {
	Draw()
}

type Window struct {
	FPS           int32
	Width, Height int32
	Editor        *Editor
	// Events        []Event
}

func NewWindow(FPS int32, Width int32, Height int32) Window {
	return Window{
		FPS,
		Width,
		Height,
		&Editor{},
	}
}

func (w *Window) Draw() {
	w.Editor.Draw()
	mouse := rl.GetMousePosition()
	mouseStr := fmt.Sprintf("Mouse X: %f Mouse Y: %f", mouse.X, mouse.Y)
	rl.DrawText(mouseStr, w.Width/2, w.Height/2, 20, rl.Pink)
	rl.DrawText("Current position: "+strconv.Itoa(w.Editor.CurrentPosition), w.Width/2, w.Height/2+30, 20, rl.Pink)
	rl.DrawText("Row: "+strconv.Itoa(w.Editor.Cursor.Row)+" Start: "+strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Start)+" Length: "+strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Length), w.Width/2, w.Height/2+60, 20, rl.Pink)
	rl.DrawText("Column: "+strconv.Itoa(w.Editor.Cursor.Column), w.Width/2, w.Height/2+90, 20, rl.Pink)
	
	char, _ := w.Editor.PieceTable.GetAt(uint(w.Editor.CurrentPosition))
	rl.DrawText("Current character: "+string(char), w.Width/2, w.Height/2+120, 20, rl.Pink)
	rl.DrawText("Previous character: "+string(w.Editor.PreviousCharacter), w.Width/2, w.Height/2+150, 20, rl.Pink)

}

func (w *Window) Input() {
	// key := rl.GetKeyPressed()
	if w.Editor.InFocus {
		if rl.IsKeyPressed(rl.KeyRight) {
			w.Editor.SetCursorPosition(w.Editor.CurrentPosition + 1)
			// position := w.Editor.CurrentPosition
			// char, err := w.Editor.PieceTable.GetAt(uint(position))
			// charRec := rl.MeasureTextEx(*w.Editor.Font, string(char), float32(w.Editor.FontSize), 0)
			// if char != '\n' {
			// 	if err == nil {
			// 		w.Editor.Cursor.Column++
			// 		w.Editor.CurrentPosition++
			// 		w.Editor.Cursor.Rectangle.X += charRec.X
			// 	}
			// } else {
			// 	w.Editor.Cursor.Column = 0
			// 	w.Editor.Cursor.Rectangle.X = w.Editor.Rectangle.X

			// 	currentRowIndex := w.Editor.Cursor.Row
			// 	currentRow := w.Editor.Rows[currentRowIndex]
			// 	w.Editor.Cursor.Rectangle.Y = currentRow.Rectangle.Y + currentRow.Rectangle.Height
			// 	w.Editor.Cursor.Row++
			// }
		}

		if rl.IsKeyPressed(rl.KeyLeft) {
			w.Editor.SetCursorPosition(w.Editor.CurrentPosition - 1)
			// position := w.Editor.CurrentPosition
			// char, err := w.Editor.PieceTable.GetAt(uint(position))
			// isNewLine := char == '\n'
			// w.Editor.Cursor.Column--
			// w.Editor.CurrentPosition--
			// if err == nil {
			// 	if !isNewLine {
			// 		charRec := rl.MeasureTextEx(*w.Editor.Font, string(char), float32(w.Editor.FontSize), 0)
			// 		w.Editor.Cursor.Rectangle.X -= charRec.X
			// 	} else {
			// 		currentRowIndex := w.Editor.Cursor.Row
			// 		if currentRowIndex != 0 {
			// 			previousLine := w.Editor.Rows[currentRowIndex-1]
			// 			w.Editor.Cursor.Rectangle.Y = previousLine.Rectangle.Y
			// 			w.Editor.Cursor.Rectangle.X = w.Editor.Rectangle.X + previousLine.Rectangle.Width
			// 		}
			// 	}
			// }
		}
	}

	// @arrow input
	// if rl.IsKeyPressed(rl.KeyLeft) {
	// 	logger.Println("Left Arrow", key)
	// }
	// if rl.IsKeyPressed(rl.KeyRight) {
	// 	logger.Println("Right Arrow", key)
	// }
	// if rl.IsKeyPressed(rl.KeyUp) {
	// 	logger.Println("Up Arrow ", key)
	// }
	// if rl.IsKeyPressed(rl.KeyDown) {
	// 	logger.Println("Down Arrow ", key)
	// }

	// @mouse input
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		logger.Println("Mouse button left")
		mousePosition := rl.GetMousePosition()
		horizontalBoundaries := mousePosition.X >= w.Editor.Rectangle.X && mousePosition.X <= w.Editor.Rectangle.X+w.Editor.Rectangle.Width
		verticalBoundaries := mousePosition.Y >= w.Editor.Rectangle.Y && mousePosition.Y <= w.Editor.Rectangle.Y+w.Editor.Rectangle.Height
		if horizontalBoundaries && verticalBoundaries {
			logger.Println("Click inside Editor")
			w.Editor.InFocus = true
			// Right now I will put the cursor at the beginning but it should be where the mouse clicked
			// X and Y should be based on row and column
			logger.Println("Cursor setted")
			w.Editor.Cursor.Rectangle.X = w.Editor.Rectangle.X
			w.Editor.Cursor.Rectangle.Y = w.Editor.Rectangle.Y
			w.Editor.Cursor.Row = 0
			w.Editor.Cursor.Column = 0
			w.Editor.CurrentPosition = 0
		}
	}
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		logger.Println("Mouse button right")
		w.Editor.SetCursorPosition(33)
	}
}
