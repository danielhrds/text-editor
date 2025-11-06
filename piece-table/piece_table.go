package piecetable

// Implement ToString cache
// - Invalidate cache if an Insert or Delete occurs

import (
	"fmt"
	"iter"
	utils "main/utils"
	"slices"
	"strconv"
	"unicode/utf8"
)

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
	ByteStart  uint
	RuneStart  uint
	ByteLength uint
	RuneLength uint
	isOriginal bool
}

func (p *Piece) Hash() uint {
	var isOriginal uint
	if p.isOriginal {
		isOriginal = 1
	}
	return Pow(p.ByteLength, p.ByteStart) + uint(isOriginal)
}

type FoundPieces struct {
	Pieces            []*Piece
	FirstPieceIndex   int
	BytePosition      uint
	ByteLength        uint 
	RuneStartPosition uint // represents first piece's start position
	ByteStartPosition uint // represents first piece's start position
	RuneEndPosition   uint
	ByteEndPosition   uint
}

type Sequence []byte

func (s Sequence) ByteLength() int {
	return len(s)
}

func (s Sequence) RuneLength() int {
	return utf8.RuneCount(s)
}

func (s Sequence) ByteForward() iter.Seq2[int, byte] {
	return func(yield func(int, byte) bool) {
		for j, _byte := range s {
			if !yield(j, _byte) {
				return
			}
		}
	}
}

func (s Sequence) RuneForward() iter.Seq2[int, rune] {
	var i int
	return func(yield func(int, rune) bool) {
		j := 0
		for j < len(s) {
			decodedRune, size := utf8.DecodeRune(s[j:])
			if !yield(i, decodedRune) {
				return
			}
			i++
			j += size
		}
	}
}

type PieceTable struct {
	OriginalBuffer      Sequence
	AddBuffer           Sequence
	AddBufferRuneLength uint
	Pieces              Collection[*Piece]
	ByteLength          uint
	RuneLength          uint
}

func NewPieceTable(content Sequence) PieceTable {
	runeLength := utf8.RuneCount(content)
	ll := NewLinkedList(&Piece{
		ByteStart:  0,
		RuneStart:  0,
		ByteLength: uint(len(content)),
		RuneLength: uint(runeLength),
		isOriginal: true,
	},
	)

	return PieceTable{
		OriginalBuffer: content,
		AddBuffer:      Sequence{},
		Pieces:         &ll,
		ByteLength:     uint(len(content)),
		RuneLength:     uint(runeLength),
	}
}

func (pt *PieceTable) PieceSequence(piece *Piece, start uint, end uint) Sequence {
	if piece.isOriginal {
		return Sequence(pt.OriginalBuffer[start:end])
	} else {
		return Sequence(pt.AddBuffer[start:end])
	}
}

func (pt *PieceTable) Runes() iter.Seq2[int, rune] {
	var i int
	return func(yield func(int, rune) bool) {
		var sequence Sequence
		for _, piece := range pt.Pieces.Forward() {
			sequence = pt.PieceSequence(piece, piece.ByteStart, piece.ByteStart+piece.ByteLength)
			for _, _rune := range sequence.RuneForward() {
				if !yield(i, _rune) {
					return
				}
				i++
			}
		}
	}
}

func (pt *PieceTable) GetBytePosition(pieces []*Piece, position uint, runeStartPosition uint, byteStartPosition uint) (uint, uint) {
	// Maybe runeStartPosition and byteStartPosition is redundant
	// because runeStartPosition is the same as the first piece's start position.
	// So I can have runeStartPosition and byteStartPosition with
	// pieces[0].RuneStart and pieces[0].ByteStart

	// TODO: Later change it to search only in one piece. runeStartPosition might be too far behind where the position is.
	j := runeStartPosition
	bytePosition := byteStartPosition
	var byteIndex uint
	for _, piece := range pieces {
		// if piece.RuneStart+piece.RuneLength < position {
		// 	j += piece.RuneLength
		//  bytePosition += piece.ByteLength
		// 	continue
		// }
		buffer := pt.PieceSequence(piece, piece.ByteStart, piece.ByteStart+piece.ByteLength)
		index := 0
		for index < len(buffer) {
			if j == position {
				return bytePosition, byteIndex
			}
			_, size := utf8.DecodeRune(buffer[index:])
			byteIndex = uint(index)
			index += size
			bytePosition += uint(size)
			j++
		}
	}
	return bytePosition, byteIndex
}

// treat as byte
func (pt *PieceTable) ToString() string {
	sequence := Sequence{}
	for _, piece := range pt.Pieces.Forward() {
		start, length := piece.ByteStart, piece.ByteStart+piece.ByteLength
		if piece.isOriginal {
			sequence = append(sequence, pt.OriginalBuffer[start:length]...)
		}
		if !piece.isOriginal {
			sequence = append(sequence, pt.AddBuffer[start:length]...)
		}
	}
	return string(sequence)
}

func (pt *PieceTable) PiecesAmount() uint {
	return uint(pt.Pieces.Size())
}

// position and length are from rune contexts
//
// sequence, startPosition, error
func (pt *PieceTable) GetSequence(position uint, length uint) (Sequence, uint, error) {
	sequence := Sequence{}
	foundPiecesMetadata, err := pt.FindPieces(position, length)
	if err != nil {
		return Sequence{}, 0, err
	}

	trackLength := 0
	for i, piece := range foundPiecesMetadata.Pieces {
		start, end := piece.ByteStart, piece.ByteStart+piece.ByteLength
		if i == 0 {
			start += foundPiecesMetadata.BytePosition - foundPiecesMetadata.ByteStartPosition
		}
		if len(foundPiecesMetadata.Pieces) > 1 && i == len(foundPiecesMetadata.Pieces)-1 {
			end = piece.ByteStart + (foundPiecesMetadata.ByteLength - uint(trackLength))
		}

		isGreaterThanLength := end-start > foundPiecesMetadata.ByteLength
		if isGreaterThanLength {
			end = start + (foundPiecesMetadata.ByteLength - uint(trackLength))
		}
		pieceSequence := pt.PieceSequence(piece, start, end)
		sequence = append(sequence, pieceSequence...)
		trackLength += int(end - start)
		if trackLength == int(foundPiecesMetadata.ByteLength) {
			return sequence, foundPiecesMetadata.RuneStartPosition, nil
		}
		foundPiecesMetadata.RuneStartPosition += piece.RuneLength
	}

	if trackLength > int(foundPiecesMetadata.ByteLength) {
		utils.Logger.Println("GetSequence: trackLength > length. Should be trackLength == length")
	}

	return Sequence{}, 0, fmt.Errorf("GetSequence: error trying to find sequence")
}

func (pt *PieceTable) GetAt(position uint) (rune, error) {
	if position > pt.RuneLength {
		return 0, fmt.Errorf("GetAt: error trying to get char at position > RuneLength")
	}

	sequence, _, err := pt.GetSequence(position, 1)
	if err != nil {
		return 0, err
	}
	if len(sequence) < 1 {
		return 0, fmt.Errorf("GetAt: error trying to get char. len(Sequence) < 1")
	}
	char, _ := utf8.DecodeRune(sequence)
	return char, nil
	// for _, char := range sequence.RuneForward() {
	// 	if position == runeStartPosition {
	// 		return char, nil
	// 	}
	// 	runeStartPosition++
	// }
	// return 0, nil
}

// position on the whole text, coming from editor, so it's in rune index.
//
// Returns foundPiece, bytePosition, runeStartPosition, runeEndPosition, byteStartPosition, byteEndPosition, index, error
func (pt *PieceTable) FindPiece(position uint) (FoundPieces, error) {
	if position > pt.RuneLength {
		return FoundPieces{}, fmt.Errorf("FindPiece: error trying to find piece at position > Length")
	}

	fp := FoundPieces{}
	startFromBeginning := position < pt.RuneLength/2
	if startFromBeginning {
		for i, piece := range pt.Pieces.Forward() {
			fp.RuneEndPosition += piece.RuneLength
			fp.ByteEndPosition += piece.ByteLength
			if fp.RuneEndPosition >= position {
				var byteIndex uint
				fp.BytePosition, byteIndex = pt.GetBytePosition(
					[]*Piece{piece},
					position,
					fp.RuneStartPosition,
					fp.ByteStartPosition,
				)
				sequence := pt.PieceSequence(piece, piece.ByteStart, piece.ByteStart+piece.ByteLength)
				_, size := utf8.DecodeRune(sequence[byteIndex:])
				fp.ByteLength = uint(size)
				fp.Pieces = []*Piece{piece}
				fp.FirstPieceIndex = i
				return fp, nil
			}
			fp.RuneStartPosition = fp.RuneEndPosition
			fp.ByteStartPosition = fp.ByteEndPosition
		}
	} else {
		fp.RuneStartPosition = pt.RuneLength
		fp.ByteStartPosition = pt.ByteLength
		for i, piece := range pt.Pieces.Backward() {
			fp.RuneStartPosition -= piece.RuneLength
			fp.RuneEndPosition = fp.RuneStartPosition + piece.RuneLength
			fp.ByteStartPosition -= piece.ByteLength
			fp.ByteEndPosition = fp.ByteStartPosition + piece.ByteLength
			if fp.RuneStartPosition < position {
				var byteIndex uint
				fp.BytePosition, byteIndex = pt.GetBytePosition(
					[]*Piece{piece},
					position,
					fp.RuneStartPosition,
					fp.ByteStartPosition,
				)
				sequence := pt.PieceSequence(piece, piece.ByteStart, piece.ByteStart+piece.ByteLength)
				_, size := utf8.DecodeRune(sequence[byteIndex:])
				fp.ByteLength = uint(size)
				fp.Pieces = []*Piece{piece}
				fp.FirstPieceIndex = i
				// i dont know if im right but the fp.RuneEndPosition should be fp.RuneEndPosition-1 i think
				return fp, nil
			}
		}
	}
	return FoundPieces{}, nil
}

// position on the whole text, coming from editor, so it's in rune index.
//
// returns foundPieces, pieceFoundIndex, bytePosition, runeStartPosition, byteStartPosition, error
func (pt *PieceTable) FindPieces(position uint, length uint) (FoundPieces, error) {
	if position > pt.RuneLength {
		return FoundPieces{}, fmt.Errorf("FindPieces: error trying to find pieces at position > RuneLength")
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

	fp := FoundPieces{}
	var startFromBeginning bool = position < pt.RuneLength/2
	if startFromBeginning {
		var runeEndPositionBeforeSwitchPiece uint
		var byteEndPositionBeforeSwitchPiece uint
		for i, piece := range pt.Pieces.Forward() {
			fp.RuneEndPosition += piece.RuneLength
			fp.ByteEndPosition += piece.ByteLength
			if fp.RuneEndPosition > position {
				if len(fp.Pieces) == 0 {
					fp.RuneStartPosition = runeEndPositionBeforeSwitchPiece
					fp.ByteStartPosition = byteEndPositionBeforeSwitchPiece
					fp.FirstPieceIndex = i
				}
				fp.Pieces = append(fp.Pieces, piece)
			}
			runeEndPositionBeforeSwitchPiece = fp.RuneEndPosition
			byteEndPositionBeforeSwitchPiece = fp.ByteEndPosition
			if fp.RuneEndPosition > position && fp.RuneEndPosition >= position+length {
				fp.BytePosition, _ = pt.GetBytePosition(
					fp.Pieces,
					position,
					fp.RuneStartPosition,
					fp.ByteStartPosition,
				)
				byteEnd, _ := pt.GetBytePosition(
					fp.Pieces,
					position+length,
					fp.RuneStartPosition,
					fp.ByteStartPosition,
				)
				fp.ByteLength = byteEnd - fp.BytePosition
				break
			}
		}
	} else {
		fp.RuneEndPosition = pt.RuneLength
		fp.ByteEndPosition = pt.ByteLength
		runeEndPosition := pt.RuneLength
		byteEndPosition := pt.ByteLength
		for i, piece := range pt.Pieces.Backward() {
			runeEndPosition -= piece.RuneLength
			byteEndPosition -= piece.ByteLength
			if runeEndPosition < position+length {
				fp.Pieces = append(fp.Pieces, piece)
			}
			if runeEndPosition <= position && runeEndPosition < position+length {
				fp.RuneStartPosition = runeEndPosition
				fp.ByteStartPosition = byteEndPosition
				fp.FirstPieceIndex = i
				if len(fp.Pieces) > 1 {
					slices.Reverse(fp.Pieces)
				}
				fp.BytePosition, _ = pt.GetBytePosition(
					fp.Pieces,
					position,
					fp.RuneStartPosition,
					fp.ByteStartPosition,
				)
				byteEnd, _ := pt.GetBytePosition(
					fp.Pieces,
					position+length,
					fp.RuneStartPosition,
					fp.ByteStartPosition,
				)
				fp.ByteLength = byteEnd - fp.BytePosition
				break
			}
			fp.RuneEndPosition -= piece.RuneLength
			fp.ByteEndPosition -= piece.ByteLength
		}
	}

	return fp, nil
}

// position on the whole text, coming from editor, so it's in rune index.
func (pt *PieceTable) Insert(position uint, text Sequence) (uint, error) {
	byteLength := uint(len(text))
	byteStart := uint(len(pt.AddBuffer))
	runeLength := uint(utf8.RuneCount(text))
	runeStart := pt.AddBufferRuneLength
	if byteLength == 0 {
		return 0, fmt.Errorf("Insert: error trying to insert string of length 0")
	}
	if position > pt.RuneLength {
		return 0, fmt.Errorf("Insert: error trying to insert string at position > RuneLength")
	}

	foundPiecesMetadata, err := pt.FindPiece(position)
	foundPiece := foundPiecesMetadata.Pieces[0]
	if err != nil {
		return 0, err
	}
	pt.AddBuffer = append(pt.AddBuffer, text...)

	piece := &Piece{
		ByteStart:  byteStart,
		ByteLength: byteLength,
		RuneStart:  runeStart,
		RuneLength: runeLength,
		isOriginal: false,
	}

	switch position {
	case 0:
		// insert at the beginning of pt.Pieces
		pt.Pieces.InsertAt(piece, 0)
	case pt.RuneLength:
		// insert at the end of pt.Pieces
		pt.Pieces.Append(piece)
	case foundPiecesMetadata.RuneStartPosition:
		// insert before foundPiece
		pt.Pieces.InsertAt(piece, int(foundPiecesMetadata.FirstPieceIndex))
	case foundPiecesMetadata.RuneEndPosition:
		// insert after foundPiece or increases piece's Length
		// Why? otherwise it would add a new piece every time.
		// Since the new sequence is added to the AddBuffer,
		// I'm assuming it's safe to just increase the piece's length
		// when it's not original
		// UPDATE: I WAS WRONG. I'M LEAVING THE COMMENT THERE AS A REMINDER
		isLastPiece := int(foundPiecesMetadata.FirstPieceIndex) == pt.Pieces.Size()-2
		isContinuous := foundPiece.RuneStart+foundPiece.RuneLength == piece.RuneStart
		if isLastPiece && !foundPiece.isOriginal && isContinuous {
			foundPiece.ByteLength += byteLength
			foundPiece.RuneLength += runeLength
		} else {
			pt.Pieces.InsertAt(piece, int(foundPiecesMetadata.FirstPieceIndex)+1)
		}
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

		foundPieceByteLengthBeforeChange := foundPiece.ByteLength
		foundPieceRuneLengthBeforeChange := foundPiece.RuneLength
		foundPiece.ByteLength = foundPiecesMetadata.BytePosition - foundPiecesMetadata.ByteStartPosition
		foundPiece.RuneLength = position - foundPiecesMetadata.RuneStartPosition
		thirdPiece := &Piece{
			ByteStart:  foundPiece.ByteStart + foundPiece.ByteLength,
			ByteLength: foundPieceByteLengthBeforeChange - foundPiece.ByteLength,
			RuneLength: foundPieceRuneLengthBeforeChange - foundPiece.RuneLength,
			RuneStart:  foundPiece.RuneStart + foundPiece.RuneLength,
			isOriginal: foundPiece.isOriginal,
		}
		pt.Pieces.InsertAt(piece, int(foundPiecesMetadata.FirstPieceIndex+1))
		pt.Pieces.InsertAt(thirdPiece, int(foundPiecesMetadata.FirstPieceIndex+2))
	}

	pt.ByteLength += byteLength
	pt.RuneLength += runeLength
	pt.AddBufferRuneLength += runeLength
	return runeLength, nil
}

func (pt *PieceTable) DeleteIfEmpty(piece *Piece, index int) bool {
	if piece.ByteLength <= 0 {
		pt.Pieces.DeleteAt(index)
		return true
	}
	return false
}

func (pt *PieceTable) Delete(position uint, length uint) error {
	if length == 0 {
		return fmt.Errorf("Delete: error trying to delete string of length 0")
	}
	if position > pt.RuneLength {
		return fmt.Errorf("Delete: error trying to delete string at position > RuneLength")
	}
	if position+length > pt.RuneLength {
		posLengthErr := strconv.Itoa(int(position + length))
		lengthErr := strconv.Itoa(int(pt.RuneLength))
		errStr := fmt.Sprintf("Delete: error trying to delete string at position+length >= RuneLength Position+RuneLength: %s, RuneLength %s", posLengthErr, lengthErr)
		return fmt.Errorf("%s", errStr)
	}

	foundPiecesMetadata, err := pt.FindPieces(position, length)
	if err != nil {
		return err
	}

	if len(foundPiecesMetadata.Pieces) == 1 {
		piece := foundPiecesMetadata.Pieces[0]
		deletionInTheMiddle := position != foundPiecesMetadata.RuneStartPosition && position+length != foundPiecesMetadata.RuneEndPosition
		if deletionInTheMiddle {
			piece.RuneLength = position - foundPiecesMetadata.RuneStartPosition
			piece.ByteLength = foundPiecesMetadata.BytePosition - foundPiecesMetadata.ByteStartPosition
			runeNewPieceStart := piece.RuneStart + length + (position - foundPiecesMetadata.RuneStartPosition)
			byteNewPieceStart := piece.ByteStart + foundPiecesMetadata.ByteLength + (position - foundPiecesMetadata.ByteStartPosition)
			newPiece := &Piece{
				RuneStart:  runeNewPieceStart,
				ByteStart:  byteNewPieceStart,
				RuneLength: foundPiecesMetadata.RuneEndPosition - (position + 1),
				ByteLength: foundPiecesMetadata.ByteEndPosition - (position + foundPiecesMetadata.ByteLength),
				isOriginal: piece.isOriginal,
			}
			pt.Pieces.InsertAt(newPiece, foundPiecesMetadata.FirstPieceIndex+1)
		} else {
			if position == foundPiecesMetadata.RuneStartPosition {
				piece.RuneStart += length
				piece.ByteStart += foundPiecesMetadata.ByteLength
			}
			piece.RuneLength -= length
			piece.ByteLength -= foundPiecesMetadata.ByteLength
		}
		pt.DeleteIfEmpty(piece, foundPiecesMetadata.FirstPieceIndex)
	} else {
		firstPiece := foundPiecesMetadata.Pieces[0]
		lastPiece := foundPiecesMetadata.Pieces[len(foundPiecesMetadata.Pieces)-1]
		piecesToDeleteIndex := 0
		deletionInTheMiddle := position != foundPiecesMetadata.RuneStartPosition
		if deletionInTheMiddle {
			runeLengthBeforeAdd := firstPiece.RuneLength
			byteLengthBeforeAdd := firstPiece.ByteLength
			firstPiece.RuneLength = position - foundPiecesMetadata.RuneStartPosition                         // position >= runeStartPosition always
			firstPiece.ByteLength = foundPiecesMetadata.BytePosition - foundPiecesMetadata.ByteStartPosition // position >= runeStartPosition always
			foundPiecesMetadata.RuneStartPosition += runeLengthBeforeAdd                                     // here we are at the second piece runeStartPosition now
			foundPiecesMetadata.ByteStartPosition += byteLengthBeforeAdd
			// deleted := pt.DeleteIfEmpty(firstPiece, foundPiecesMetadata.FirstPieceIndex)
			// if !deleted {
			// 	foundPiecesMetadata.FirstPieceIndex++
			// }
			piecesToDeleteIndex = 1
			foundPiecesMetadata.FirstPieceIndex++
		}
		for i := piecesToDeleteIndex; i < len(foundPiecesMetadata.Pieces)-1; i++ {
			piece := foundPiecesMetadata.Pieces[i]
			foundPiecesMetadata.RuneStartPosition += piece.RuneLength
			foundPiecesMetadata.ByteStartPosition += piece.ByteLength
			pt.Pieces.DeleteAt(foundPiecesMetadata.FirstPieceIndex)
		}
		// as the loop ends we are at the last piece runeStartPosition
		lastPiece.RuneStart += (position + length) - foundPiecesMetadata.RuneStartPosition
		lastPiece.ByteStart += (foundPiecesMetadata.BytePosition + foundPiecesMetadata.ByteLength) - foundPiecesMetadata.ByteStartPosition
		lastPiece.RuneLength -= (position + length) - foundPiecesMetadata.RuneStartPosition
		lastPiece.ByteLength -= (foundPiecesMetadata.BytePosition + foundPiecesMetadata.ByteLength) - foundPiecesMetadata.ByteStartPosition
		pt.DeleteIfEmpty(lastPiece, foundPiecesMetadata.FirstPieceIndex)
	}

	pt.RuneLength -= length
	pt.ByteLength -= foundPiecesMetadata.ByteLength
	return nil
}

func (pt *PieceTable) Empty() bool {
	return pt.RuneLength == 0 && pt.ByteLength == 0
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
