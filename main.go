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
	// editor := NewEditor(rl.NewRectangle(20, 0, float32(window.Width-100), float32(window.Height-100)), rl.Gray)
	// editor := NewEditor(rl.NewRectangle(0, 0, 250, float32(window.Height-100)), rl.NewColor(30, 30, 30, 255))
	editor := NewEditor(rl.NewRectangle(0, 0, float32(window.Width), float32(window.Height)), rl.NewColor(30, 30, 30, 255))
	font := rl.LoadFontEx("fonts/JetBrainsMono-Regular.ttf", int32(editor.FontSize), nil, 0)
	defer rl.UnloadFont(font)
	editor.Font = &font
	editor.PieceTable = &pt
	window.Editor = &editor
	window.Editor.CalculateRows()

	for !rl.WindowShouldClose() {
		rl.ClearBackground(rl.White)
		window.Input()
		rl.BeginDrawing()
		window.Draw()
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
	currentRow := w.Editor.Rows[w.Editor.Cursor.Row]
	rl.DrawText(mouseStr, w.Width/2, w.Height/2, 20, rl.Pink)
	rl.DrawText("Current position: "+strconv.Itoa(w.Editor.Cursor.CurrentIndex), w.Width/2, w.Height/2+30, 20, rl.Pink)
	rl.DrawText(
		"Row: "+strconv.Itoa(w.Editor.Cursor.Row)+
			" | Start: "+strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Start)+
			" | Length: "+strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Length)+
			" | AutoNewLine: "+strconv.FormatBool(currentRow.AutoNewLine),
		w.Width/2,
		w.Height/2+60,
		20,
		rl.Pink,
	)
	rowWidth := currentRow.Rectangle.Width
	rl.DrawText("Row Width: "+strconv.Itoa(int(rowWidth)), w.Width/2, w.Height/2+90, 20, rl.Pink)
	rl.DrawText("Column: "+strconv.Itoa(w.Editor.Cursor.Column), w.Width/2, w.Height/2+120, 20, rl.Pink)

	char, _ := w.Editor.PieceTable.GetAt(uint(w.Editor.Cursor.CurrentIndex))
	rl.DrawText("Current character: "+string(char), w.Width/2, w.Height/2+150, 20, rl.Pink)
	rl.DrawText("Previous character: "+string(w.Editor.PreviousCharacter), w.Width/2, w.Height/2+180, 20, rl.Pink)

	rl.DrawText("Cursor X: "+strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.X), 'f', 2, 32), w.Width/2, w.Height/2+210, 20, rl.Pink)
	rl.DrawText("Cursor Y: "+strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.Y), 'f', 2, 32), w.Width/2, w.Height/2+240, 20, rl.Pink)
	rl.DrawText("FPS: "+strconv.Itoa(int(rl.GetFPS())), w.Width/2, w.Height/2+270, 20, rl.Pink)

	rl.DrawRectangle(
		int32(rowWidth),
		currentRow.Rectangle.ToInt32().Y,
		w.Editor.Cursor.Rectangle.ToInt32().Width,
		w.Editor.Cursor.Rectangle.ToInt32().Height,
		rl.NewColor(rl.Pink.R, rl.Pink.G, rl.Pink.B, 128),
	)
}

func (w *Window) Input() {
	// if w.Editor.InFocus {
	if true { // this should be on editor struct like editor.update()
		// char := rl.GetCharPressed()

		// @arrow input
		if rl.IsKeyDown(rl.KeyRight) {
			w.Editor.MoveCursorForward()
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			w.Editor.MoveCursorBackward()
		}
		if rl.IsKeyDown(rl.KeyUp) {
			w.Editor.MoveCursorUpward()
		}
		if rl.IsKeyDown(rl.KeyDown) {
			w.Editor.MoveCursorDownward()
		}
	}

	// @mouse input
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		err := w.Editor.SetCursorPositionByClick(rl.GetMousePosition())
		if err != nil {
			log.Fatal("Mouse right click: ", err)
		}
	}
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		// ------------ Debugging ------------

		logger.Println("")
		logger.Println("Current position: " + strconv.Itoa(w.Editor.Cursor.CurrentIndex))
		logger.Println("Row: " + strconv.Itoa(w.Editor.Cursor.Row) + " Start: " + strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Start) + " Length: " + strconv.Itoa(w.Editor.Rows[w.Editor.Cursor.Row].Length))
		logger.Println("Row Width: " + strconv.Itoa(int(w.Editor.Rows[w.Editor.Cursor.Row].Rectangle.Width)))
		logger.Println("Column: " + strconv.Itoa(w.Editor.Cursor.Column))
		char, _ := w.Editor.PieceTable.GetAt(uint(w.Editor.Cursor.CurrentIndex))
		logger.Println("Current character: " + string(char))
		logger.Println("Previous character: " + string(w.Editor.PreviousCharacter))
		logger.Println("Cursor X: " + strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.X), 'f', 2, 32))
		logger.Println("Cursor Y: " + strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.Y), 'f', 2, 32))

		// -----------------------------------

	}
}
