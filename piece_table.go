package main

import (
	"fmt"
	"iter"
	"log"
	"os"
	"slices"
	"strconv"
)

// ----------- DEBUG -----------
var logger = log.New(os.Stdout, "\033[33m[DEBUG]\033[0m ", 0)
// -----------------------------

// I didn't want to put Iterators methods in the Collection interface
// but I couldn't come up with a good solution because I'm dumb
type Collection[T any] interface {
	Append(value T)
	Pop()
	InsertAt(value T, index int) error
	DeleteAt(index int) error
	GetAt(index int) (T, error)
	Size() int
	Forward() iter.Seq2[int, T]
	Backward() iter.Seq2[int, T]
}

// ---------------------------------------------------
// Golang implementation of piece table datastructure
// ---------------------------------------------------

type Piece struct {
	Start      uint
	Length     uint
	isOriginal bool
}

func (p *Piece) Hash() uint {
	var isOriginal uint
	if p.isOriginal {
		isOriginal = 1
	}
	return Pow(p.Length, p.Start) + uint(isOriginal)
}

// Since piece table is a data structure like any other,
// I could implement the methods to support the Collection interface,
// But I only realized this now and I'm very lazy ngl
type Sequence = []rune
type PieceTable struct {
	OriginalBuffer Sequence
	AddBuffer      Sequence
	Pieces         Collection[*Piece]
	Length         uint
}

func NewPieceTable(content []rune) PieceTable {
	ll := NewLinkedList(&Piece{
		Start:      0,
		Length:     uint(len(content)),
		isOriginal: true,
	},
	)

	return PieceTable{
		OriginalBuffer: content,
		AddBuffer:      []rune{},
		Pieces:         &ll,
		Length:         uint(len(content)),
	}
}

func (pt *PieceTable) ToString() string {
	sequence := Sequence{}
	for _, piece := range pt.Pieces.Forward() {
		start, length := piece.Start, piece.Start+piece.Length
		if piece.isOriginal {
			sequence = append(sequence, pt.OriginalBuffer[start:length]...)
		}
		if !piece.isOriginal {
			sequence = append(sequence, pt.AddBuffer[start:length]...)
		}
	}
	return string(sequence)
}

func (pt *PieceTable) GetSequence(position uint, length uint) (Sequence, uint, error) {
	sequence := Sequence{}
	// piece, startPosition, endPosition, _, err := pt.FindPiece(start)
	pieces, _, startPosition, err := pt.FindPieces(position, length)
	if err != nil {
		return Sequence{}, 0, err
	}
	for _, piece := range pieces {
		start, length := piece.Start, piece.Start+piece.Length
		if piece.isOriginal {
			sequence = append(sequence, pt.OriginalBuffer[start:length]...)
		}
		if !piece.isOriginal {
			sequence = append(sequence, pt.AddBuffer[start:length]...)
		}
	}

	return sequence, startPosition, nil
}

func (pt *PieceTable) GetAt(position uint) (rune, error) {
	logger.Println("Position", position)
	if position > pt.Length {
		return 0, fmt.Errorf("GetAt: error trying to get char at position > Length")
	}

	sequence, startPosition, err := pt.GetSequence(position, 1)
	if err != nil {
		return 0, err
	}
	if len(sequence) < 1 {
		return 0, fmt.Errorf("GetAt: error trying to get char. len(Sequence) < 1")
	}
	var char rune
	for _, char = range sequence {
		if position == startPosition {
			break
		}
		startPosition++
	}
	return char, nil
}


func (pt *PieceTable) PrintPosition(position uint) (string, error) {
	if position > pt.Length {
		return "", fmt.Errorf("PrintPosition: error trying to print position at position > Length")
	}

	str := "\n" + pt.ToString() + "\n"
	var i uint
	for i = range uint(len(str) - 1) {
		if i != position {
			str += " "
		} else {
			str += "^"
		}
	}
	str += "Position: " + strconv.Itoa(int(position))
	return str, nil
}

func (pt *PieceTable) PiecesAmount() uint {
	return uint(pt.Pieces.Size())
}

// maybe return error if position is greater than the length.
// Returns Piece, startPosition, endPosition, index, and error
func (pt *PieceTable) FindPiece(position uint) (*Piece, uint, uint, uint, error) {
	// IMPORTANT TO READ:
	// The whole thing with 'startPosition' and 'endPosition' is needed because it is good to know where we are on the text index.
	// Since we will have the piece's length, return just startPosition is enough though.

	if position > pt.Length {
		return nil, 0, 0, 0, fmt.Errorf("FindPiece: error trying to find piece at position > Length")
	}

	startFromBeginning := position < pt.Length/2
	if startFromBeginning {
		var startPosition uint
		var endPosition uint
		for i, piece := range pt.Pieces.Forward() {
			endPosition += piece.Length
			if endPosition > position {
				return piece, startPosition, endPosition, uint(i), nil
			}
			startPosition = endPosition
		}
	}
	if !startFromBeginning {
		var startPosition uint = pt.Length
		var endPosition uint
		for i, piece := range pt.Pieces.Backward() {
			startPosition -= piece.Length
			endPosition = startPosition + piece.Length
			if startPosition <= position {
				return piece, startPosition, endPosition, uint(i), nil
			}
		}
	}
	return nil, 0, 0, 0, nil
}

func (pt *PieceTable) FindPieces(position uint, length uint) ([]*Piece, int, uint, error) {
	if position > pt.Length {
		return nil, 0, 0, fmt.Errorf("FindPieces: error trying to find pieces at position > Length")
	}

	// I will try to give an example to remind me later because I know I will forget
	// If text is "_______________" Length 15
	// And we are trying to delete that span:
	// ____________________
	// 			^     ^
	// And we have three pieces like that (just to clear things: each piece ends where the other starts):
	// ____________________
	// ^       ^      ^
	// p1      p2     p3
	// That means the algorithm (forward) will do something like:
	// i = 0; endPosition = 0
	// ________ That's p1 length, so you can see some part of p1 will be deleted. So it will be added to pieces array that will be returned
	// Also, that means p1End > position
	//
	// i = 1; endPosition = p1's start
	// @@@@@@@@_______@@@@@
	//         ^ That's p2 start and length, so you can see some part of p2 will be deleted too.
	// Also, that means p2End > position
	//
	// i = 2; endPosition = p2's start
	// @@@@@@@@@@@@@@@_____
	//                ^ That's p3 start, you can see that it will not be affected by the delete.
	// Also, that means p3End > position and p3End

	var pieces []*Piece
	// var pieceIndexes []uint
	var firstPieceFoundIndex int
	var startPosition uint
	var startFromBeginning bool = position < pt.Length/2
	if startFromBeginning {
		var endPosition uint
		var endPositionBeforeSwitchPiece uint
		for i, piece := range pt.Pieces.Forward() {
			if endPosition > position && endPosition > position+length {
				break
			}

			endPosition += piece.Length
			if endPosition > position {
				if len(pieces) == 0 {
					startPosition = endPositionBeforeSwitchPiece
					firstPieceFoundIndex = i
				}
				pieces = append(pieces, piece)
				// pieceIndexes = append(pieceIndexes, uint(i))
			}
			endPositionBeforeSwitchPiece = endPosition
		}
	}
	if !startFromBeginning {
		var endPosition uint = pt.Length
		// var endPositionBeforeSwitchPiece uint = pt.Length
		for i, piece := range pt.Pieces.Backward() {
			endPosition -= piece.Length
			if endPosition < position+length {
				// if len(pieces) == 0 {
				// 	// startPosition = endPositionBeforeSwitchPiece
				// 	firstPieceFoundIndex = i
				// }
				pieces = append(pieces, piece)
				// pieceIndexes = append(pieceIndexes, uint(i))
			}
			if endPosition < position && endPosition < position+length {
				startPosition = endPosition
				firstPieceFoundIndex = i
				break
			}
		}
		if len(pieces) > 1 {
			slices.Reverse(pieces)
		}
	}

	return pieces, firstPieceFoundIndex, startPosition, nil
}

// position Means the index on the whole text. start Means the start in add buffer
func (pt *PieceTable) Insert(position uint, text Sequence) error {
	length := uint(len(text))
	start := uint(len(pt.AddBuffer))
	if length == 0 {
		return fmt.Errorf("Insert: error trying to insert string of length 0")
	}
	// if position < 0 {
	// 	return fmt.Errorf("Insert: error trying to insert string at position < 0")
	// }
	if position > pt.Length {
		return fmt.Errorf("Insert: error trying to insert string at position > Length")
	}

	pt.AddBuffer = append(pt.AddBuffer, text...)

	piece := &Piece{
		Start:      start,
		Length:     length,
		isOriginal: false,
	}

	foundPiece, foundPieceStartPosition, foundPieceEndPosition, index, err := pt.FindPiece(position)
	if err != nil {
		return err
	}

	switch position {
	case 0:
		// insert at the beginning of pt.Pieces
		pt.Pieces.InsertAt(piece, 0)
	case pt.Length:
		// insert at the end of pt.Pieces
		pt.Pieces.Append(piece)
	case foundPieceStartPosition:
		// insert before foundPiece
		// pt.Pieces = slices.Insert(pt.Pieces, int(index), piece)
		pt.Pieces.InsertAt(piece, int(index))
	case foundPieceEndPosition:
		// insert after foundPiece
		// pt.Pieces = slices.Insert(pt.Pieces, int(index+1), piece)
		pt.Pieces.InsertAt(piece, int(index+1))
	default:
		// I will try to give an example to remind me later because I know I will forget
		// If text is "_______________" Length 15
		// And foundPiece starts at position 2 and has length 10
		// And we want to insert at position 7:
		// _______________
		//   ^ 							Here it's the foundPiece's start, position 2
		// _______________
		//        ^ 			 	Here is where we will split, position 7
		// _______________
		//             ^ 		Here is foundPiece's end, position 12, so the foundPiece length is 12-2 = 10
		// So basically we need to split foundPiece into three pieces.
		// ----------------------------------------------------------------------------------
		// The first half will be a piece that starts at the same 'start' on whatever buffer it is from, but with a length of 7-2 = 5
		// The second piece, the middle one, will be the new one being added on position 7
		// The second half will be quite different. Before that process it was a single piece starting at position 2,
		// but now it needs it's start to be adjusted. So what we need to do is:
		// Get the old start and add the new length of the first half.
		// To conclude: oldStart + newFirstHalfLength = second half start
		// ----------------------------------------------------------------------------------

		foundPieceLengthBeforeChange := foundPiece.Length
		foundPiece.Length = position - foundPieceStartPosition
		pieceSplitTwoLength := &Piece{
			Start:      foundPiece.Start + foundPiece.Length,
			Length:     foundPieceLengthBeforeChange - foundPiece.Length,
			isOriginal: foundPiece.isOriginal,
		}
		// pt.Pieces = slices.Insert(pt.Pieces, int(index+1), piece)
		pt.Pieces.InsertAt(piece, int(index+1))
		pt.Pieces.InsertAt(pieceSplitTwoLength, int(index+2))
		// pt.Pieces = slices.Insert(pt.Pieces, int(index+2), pieceSplitTwoLength)
	}

	pt.Length += length
	return nil
}

func (pt *PieceTable) DeleteIfEmpty(piece *Piece, index int) {
	if piece.Length <= 0 {
		pt.Pieces.DeleteAt(index)
	}
}

func (pt *PieceTable) Delete(position uint, length uint) error {
	// OBS: Since deletion shifts items from an array to the left, Delete would benefit from a linked list.

	if position == 0 {
		return fmt.Errorf("Delete: error trying to delete string of length 0")
	}
	// if position < 0 {
	// 	return fmt.Errorf("Delete: error trying to delete string at position < 0")
	// }
	if position > pt.Length {
		return fmt.Errorf("Delete: error trying to delete string at position > Length")
	}
	if position+length > pt.Length {
		posLengthErr := strconv.Itoa(int(position + length))
		lengthErr := strconv.Itoa(int(pt.Length))
		errStr := fmt.Sprintf("Delete: error trying to delete string at position+length >= Length Position+Length: %s, Length %s", posLengthErr, lengthErr)
		return fmt.Errorf("%s", errStr)
	}

	// fast path beginning
	// fath path end
	// middle:
	// 	- search source
	//  - to know from where to start searching (forward or backwards): greaterThanMiddle = position > (len(OriginalBuffer) + len(AddBuffer))/2
	//  - search:
	//  - - iterate through Pieces and find which ones has
	// I will try to give an example to remind me later because I know I will forget
	// If text is "_______________" Length 15
	// Three pieces composes that text:
	// _______________
	// ^    ^    ^
	// p1   p2   p3
	// And we are trying to delete:
	// _______________
	//  ^           ^
	//	| that span |
	// That means all three pieces are involved, so:
	// 		p1 would have it's length decreased
	// 		p2 would be deleted completely
	//		p3 would have it's start shifted to the right

	//
	// ad = add buffer
	// og = original
	// @ = will remain

	// |    @   |            |  @ |-
	//          |   deleted  |
	//  ttttttttttttt tttttttttttt tttttttttt ttttt
	// | og          | ad         | og       | ad

	// | @      |                        | @
	//          |   deleted              |
	//  tttttttt tttttttttttttttttttttttt tttttt
	// | og     | ad         | og | ad
	// in that case, pieces inside deletion will be removed
	// - first ad will be removed
	// - second og will be removed
	// - second ad will have it's start adjusted

	// | @      |                      | @
	//          |   deleted            |
	//  tttttttt tttttttttttt tttt tttttttttt ttttt
	// | og     | ad         | og | ad       | og
	// in that case, pieces inside deletion will be removed
	// - first ad will be removed
	// - second og will be removed
	// - second ad will have it's start adjusted

	// This risky I and don't know if it's right:
	// I'm assuming that when I delete a piece from the linked list,
	// the next one will be on the same index, so if I keep deleting
	// pieces on the same index when I need to delete it, it'll be safe...
	pieces, index, startPosition, err := pt.FindPieces(position, length)
	if err != nil {
		return err
	}
	if len(pieces) == 1 {
		piece := pieces[0]
		if position == startPosition {
			piece.Start += piece.Length - length
		}
		piece.Length -= piece.Length - length
		pt.DeleteIfEmpty(piece, index)
	} else if len(pieces) == 2 {
		firstPiece := pieces[0]
		lastPiece := pieces[len(pieces)-1]
		if position == startPosition {
			// IMPORTANT: I don't know if it's a valid thought:
			// If position == startPosition, we are at the beginning of a piece,
			// But if the len(pieces) == 2 we know that the delete affected other piece,
			// So how wrong for me to assume that the entire piece need to be deleted?
			// Like:
			// 						__________________________
			//                ^   deletion   ^
			//								^            ^
			//								p1           p2
			//
			// This is the only case where two pieces are affected and the position == startPosition, no?
			pt.Pieces.DeleteAt(index)
			startPosition += firstPiece.Length
		} else {
			// IMPORTANT: Same thing as above, an assumption:
			// The only case where two pieces are affected and now position != startPosition
			// is when the position is at the middle of the first piece at it's length enters
			// another piece, no?
			// In that case, the firstPiece should have it's length decreased.
			lengthBeforeAdd := firstPiece.Length
			firstPiece.Length = position - startPosition
			startPosition += lengthBeforeAdd // here we are at the second piece startPosition now
			pt.DeleteIfEmpty(firstPiece, index)
			index++
		}
		lastPiece.Start += (position + length) - startPosition
		lastPiece.Length -= (position + length) - startPosition
		pt.DeleteIfEmpty(lastPiece, index)
	} else if len(pieces) > 2 {
		firstPiece := pieces[0]
		lengthBeforeAdd := firstPiece.Length
		firstPiece.Length = position - startPosition // position >= startPosition always
		startPosition += lengthBeforeAdd             // here we are at the second piece startPosition now
		pt.DeleteIfEmpty(firstPiece, index)
		index++
		for i := 1; i < len(pieces)-1; i++ {
			piece := pieces[i]
			startPosition += piece.Length
			pt.DeleteIfEmpty(piece, index)
		}
		// as the loop ends we are at the last piece startPosition
		lastPiece := pieces[len(pieces)-1]
		lastPiece.Start += (position + length) - startPosition
		lastPiece.Length -= (position + length) - startPosition
		pt.DeleteIfEmpty(lastPiece, index)
	}

	pt.Length -= length
	return nil
}

// utils

func Pow(x uint, y uint) uint {
	if y == 0 {
		return 1
	}

	if y == 1 {
		return x
	}

	result := x
	var i uint
	for i = 2; i <= y; i++ {
		result *= x
	}
	return result
}
