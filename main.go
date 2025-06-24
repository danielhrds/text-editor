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
	// editor := NewEditor(rl.NewRectangle(20, 0, float32(window.Width-100), float32(window.Height-100)), rl.Gray)
	// editor := NewEditor(rl.NewRectangle(20, 0, 250, float32(window.Height-100)), rl.NewColor(30,30,30,255))
	editor := NewEditor(rl.NewRectangle(0, 0, float32(window.Width), float32(window.Height)), rl.NewColor(30,30,30,255))
	font := rl.LoadFontEx("fonts/JetBrainsMono-Regular.ttf", int32(editor.FontSize), nil, 0)
	defer rl.UnloadFont(font)
	editor.Font = &font
	editor.PieceTable = &pt
	window.Editor = &editor
	// window.Editor.InFocus = true
	window.Editor.CalculateRows()
	// window.Editor.SetCursorPositionFromIndex(18)

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
	rl.DrawText("Current position: "+strconv.Itoa(w.Editor.Cursor.CurrentIndex), w.Width/2, w.Height/2+30, 20, rl.Pink)
	rl.DrawText("Row: "+strconv.Itoa(w.Editor.Cursor.Row)+" Start: "+strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Start)+" Length: "+strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Length), w.Width/2, w.Height/2+60, 20, rl.Pink)
	rl.DrawText("Column: "+strconv.Itoa(w.Editor.Cursor.Column), w.Width/2, w.Height/2+90, 20, rl.Pink)

	char, _ := w.Editor.PieceTable.GetAt(uint(w.Editor.Cursor.CurrentIndex))
	rl.DrawText("Current character: "+string(char), w.Width/2, w.Height/2+120, 20, rl.Pink)
	rl.DrawText("Previous character: "+string(w.Editor.PreviousCharacter), w.Width/2, w.Height/2+150, 20, rl.Pink)

}

func (w *Window) Input() {
	if w.Editor.InFocus {
		// @arrow input
		if rl.IsKeyPressed(rl.KeyRight) {
			w.Editor.SetCursorPositionFromIndex(w.Editor.Cursor.CurrentIndex + 1)
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			w.Editor.SetCursorPositionFromIndex(w.Editor.Cursor.CurrentIndex - 1)
		}

		// @keyboard input
		if rl.IsKeyPressed(rl.KeyBackspace) && w.Editor.Cursor.CurrentIndex > 0 {
			logger.Println("Deleting")
			w.Editor.PieceTable.Delete(uint(w.Editor.Cursor.CurrentIndex)-1, 1)
			w.Editor.CalculateRows()
			w.Editor.SetCursorPositionFromIndex(w.Editor.Cursor.CurrentIndex-1)
		}
		
		char := rl.GetCharPressed()
		if char > 0 {
			logger.Println(char, string(char), w.Editor.Cursor.CurrentIndex, []rune{char})
			w.Editor.Insert(w.Editor.Cursor.CurrentIndex, []rune{char})
			logger.Println(w.Editor.PieceTable.ToString())
		}

	}

	// @mouse input
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		logger.Println("Mouse button left")
		mouseClickPosition := rl.GetMousePosition()
		horizontalBoundaries := mouseClickPosition.X >= w.Editor.Rectangle.X && mouseClickPosition.X <= w.Editor.Rectangle.X+w.Editor.Rectangle.Width
		verticalBoundaries := mouseClickPosition.Y >= w.Editor.Rectangle.Y && mouseClickPosition.Y <= w.Editor.Rectangle.Y+w.Editor.Rectangle.Height
		if horizontalBoundaries && verticalBoundaries {
			logger.Println("Click inside Editor")
			logger.Println("Cursor setted")
			w.Editor.InFocus = true
			w.Editor.SetCursorPositionFromClick(mouseClickPosition)
		}
	}
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		logger.Println("Mouse button right")
		w.Editor.SetCursorPositionFromIndex(33)
	}
}
