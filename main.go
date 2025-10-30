package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	pt "main/piece-table"
	"main/utils"
	"os"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var keys []int32 = make([]int32, 0)

// @main
func main() {
	defer rl.CloseWindow()
	window := NewWindow(60, 1600, 900)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(window.Width, window.Height, "Text Editor")
	rl.SetWindowState(rl.FlagWindowAlwaysRun)
	rl.SetTargetFPS(window.FPS)

	original, _ := ReadFile("example.txt")
	// original, _ := ReadFile("output/output 8.txt")
	pt := pt.NewPieceTable(
		pt.Sequence(original),
	)
	// pt.Insert(20, Sequence("went to the park and\n"))
	utils.Logger.Println(pt.ToString())

	// original, _ := ReadFile("example2.txt")
	// original, _ := ReadFile("example3.txt")
	// editor := NewEditor(rl.NewRectangle(20, 0, float32(window.Width-100), float32(window.Height-100)), rl.Gray)
	// editor := NewEditor(rl.NewRectangle(0, 0, 255, float32(window.Height-100)), rl.NewColor(30, 30, 30, 255))
	editor := NewEditor(rl.NewRectangle(0, 0, float32(window.Width), float32(window.Height)), rl.NewColor(30, 30, 30, 255))
	defer func() {
		if r := recover(); r != nil {
			OutputText(*editor.PieceTable)
			OutputKeys()
			os.Exit(1)
		}
	}()

	latin1 := make([]rune, 0, 255-32+1)
	var cp rune
	for cp = 32; cp <= 255; cp++ {
		latin1 = append(latin1, cp)
	}
	font := rl.LoadFontEx("fonts/JetBrainsMono-Regular.ttf", int32(editor.FontSize), latin1, int32(len(latin1)))
	rl.GenTextureMipmaps(&font.Texture)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	defer rl.UnloadFont(font)
	editor.ChangeFont(&font)
	editor.PieceTable = &pt
	window.Editor = &editor
	window.Editor.CalculateLines()

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
	i := 0
	for scanner.Scan() {
		fmt.Println(scanner.Text())
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
	currentLine := w.Editor.Lines[w.Editor.Cursor.Line]
	rl.DrawText(mouseStr, w.Width/2, w.Height/2, 20, rl.Pink)
	rl.DrawText("Current position: "+strconv.Itoa(w.Editor.Cursor.CurrentIndex), w.Width/2, w.Height/2+30, 20, rl.Pink)
	rl.DrawText(
		"Line: "+strconv.Itoa(w.Editor.Cursor.Line)+
			" | Start: "+strconv.Itoa(w.Editor.Lines[w.Editor.Cursor.Line].Start)+
			" | Length: "+strconv.Itoa(w.Editor.Lines[w.Editor.Cursor.Line].Length)+
			" | AutoNewLine: "+strconv.FormatBool(currentLine.AutoNewLine),
		w.Width/2,
		w.Height/2+60,
		20,
		rl.Pink,
	)
	lineWidth := currentLine.Rectangle.Width
	rl.DrawText("Line Width: "+strconv.Itoa(int(lineWidth)), w.Width/2, w.Height/2+90, 20, rl.Pink)
	rl.DrawText("Column: "+strconv.Itoa(w.Editor.Cursor.Column), w.Width/2, w.Height/2+120, 20, rl.Pink)

	char, _ := w.Editor.PieceTable.GetAt(uint(w.Editor.Cursor.CurrentIndex))
	rl.DrawText("Current character: "+string(char), w.Width/2, w.Height/2+150, 20, rl.Pink)
	rl.DrawText("Previous character: "+string(w.Editor.PreviousCharacter), w.Width/2, w.Height/2+180, 20, rl.Pink)

	rl.DrawText("Cursor X: "+strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.X), 'f', 2, 32), w.Width/2, w.Height/2+210, 20, rl.Pink)
	rl.DrawText("Cursor Y: "+strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.Y), 'f', 2, 32), w.Width/2, w.Height/2+240, 20, rl.Pink)
	rl.DrawText("FPS: "+strconv.Itoa(int(rl.GetFPS())), w.Width/2, w.Height/2+270, 20, rl.Pink)

	rl.DrawText("Editor.EditorRec.X: "+strconv.Itoa(int(w.Editor.EditorRec.X)), w.Width/2, w.Height/2+300, 20, rl.Pink)
	rl.DrawText("Editor.WritableRec.X: "+strconv.Itoa(int(w.Editor.WritableRec.X)), w.Width/2, w.Height/2+330, 20, rl.Pink)

	rl.DrawRectangle(
		int32(w.Editor.WritableRec.X+lineWidth),
		currentLine.Rectangle.ToInt32().Y,
		w.Editor.Cursor.Rectangle.ToInt32().Width,
		w.Editor.Cursor.Rectangle.ToInt32().Height,
		rl.NewColor(rl.Pink.R, rl.Pink.G, rl.Pink.B, 128),
	)
}

func (w *Window) Input() {
	// if w.Editor.InFocus {
	if true { // this should be on editor struct like editor.update()
		char := rl.GetCharPressed()
		if char != 0 {
			fmt.Println(char, "string:", string(char), w.Editor.CharRectangle(char))
		}

		// @arrows input
		if rl.IsKeyPressed(rl.KeyRight) {
			w.Editor.MoveCursorForward()
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			w.Editor.MoveCursorBackward()
		}
		if rl.IsKeyPressed(rl.KeyUp) {
			w.Editor.MoveCursorUpward()
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			w.Editor.MoveCursorDownward()
		}

		if rl.IsKeyPressed(rl.KeyApostrophe) {
			OutputText(*w.Editor.PieceTable)
			OutputKeys()
			os.Exit(1)
		}

		if rl.IsKeyPressed(rl.KeyBackspace) {
			w.Editor.Delete(w.Editor.Cursor.CurrentIndex, 1)
		}

		if char != 0 {
			keys = append(keys, char)
			w.Editor.Insert(w.Editor.Cursor.CurrentIndex, pt.Sequence([]byte(string(char))))
			// w.Editor.PieceTable.Insert(uint(w.Editor.Cursor.CurrentIndex), []rune{char})
			// w.Editor.MoveCursorForward()
			// w.Editor.CalculateLines()
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

		utils.Logger.Println("")
		utils.Logger.Println("Current position: " + strconv.Itoa(w.Editor.Cursor.CurrentIndex))
		utils.Logger.Println("Line: " + strconv.Itoa(w.Editor.Cursor.Line) + " Start: " + strconv.Itoa(w.Editor.Lines[w.Editor.Cursor.Line].Start) + " Length: " + strconv.Itoa(w.Editor.Lines[w.Editor.Cursor.Line].Length))
		utils.Logger.Println("Line Width: " + strconv.Itoa(int(w.Editor.Lines[w.Editor.Cursor.Line].Rectangle.Width)))
		utils.Logger.Println("Column: " + strconv.Itoa(w.Editor.Cursor.Column))
		char, _ := w.Editor.PieceTable.GetAt(uint(w.Editor.Cursor.CurrentIndex))
		utils.Logger.Println("Current character: " + string(char))
		utils.Logger.Println("Previous character: " + string(w.Editor.PreviousCharacter))
		utils.Logger.Println("Cursor X: " + strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.X), 'f', 2, 32))
		utils.Logger.Println("Cursor Y: " + strconv.FormatFloat(float64(w.Editor.Cursor.Rectangle.Y), 'f', 2, 32))

	}
}

func OutputText(pt pt.PieceTable) {
	n := 1
	for {
		fileName := fmt.Sprintf("output/output %d.txt", n)
		_, err := os.Stat(fileName)
		if errors.Is(err, os.ErrNotExist) {
			file, _ := os.Create(fileName)
			defer file.Close()
			file.WriteString(pt.ToString())
			return
		}
		n++
	}
}

func OutputKeys() {
	n := 1
	for {
		fileName := fmt.Sprintf("output/output %d keys.txt", n)
		_, err := os.Stat(fileName)
		if errors.Is(err, os.ErrNotExist) {
			file, _ := os.Create(fileName)
			defer file.Close()
			str := string(keys)
			file.WriteString(str)
			return
		}
		n++
	}
}
