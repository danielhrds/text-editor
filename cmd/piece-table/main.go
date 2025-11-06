package main

import (
	"fmt"
	ptm "main/piece-table"
	utils "main/utils"
)

// TODO: add assert

func test1() {
	utils.Logger.Println("TEST 1: WITHOUT MULTIBYTE CHARACTERS")
	original, _ := utils.ReadFile("../../example.txt")
	pt := ptm.NewPieceTable(
		ptm.Sequence(original),
	)
	fmt.Println(pt.ToString())
	pt.Insert(1, ptm.Sequence("@----#"))
	fmt.Println(pt.ToString())
	
	pt.Insert(10, ptm.Sequence("@----#"))
	fmt.Println(pt.ToString())
	// uncomment this to test if it is creating another piece 
	// when inserting at the end of the last piece. it should not.
	// it should increase the last piece's length
	// pt.Insert(16, ptm.Sequence("TESTE")) 
	// fmt.Println(pt.ToString())
	
	pt.Delete(8, 1)
	fmt.Println(pt.ToString())
	
	pt.Delete(12, 4) // should find 2 pieces
	fmt.Println(pt.ToString())

	pt.Delete(9, 4) // should find 2 pieces
	fmt.Println(pt.ToString())

	pt.Delete(6, 2) // should find 2 pieces
	fmt.Println(pt.ToString())
	
	// pt.Insert(12, ptm.Sequence("TESTE"))
	// fmt.Println(pt.ToString())

	utils.Logger.Println("TEST 1 ENDED")
}

func test2() {
	utils.Logger.Println("TEST 2: WITH MULTIBYTE CHARACTERS")
	original, _ := utils.ReadFile("../../example.txt")
	pt := ptm.NewPieceTable(
		ptm.Sequence(original),
	)
	fmt.Println(pt.ToString())

	pt.Insert(0, ptm.Sequence("ç"))
	fmt.Println(pt.ToString())

	pt.Insert(1, ptm.Sequence("ç"))
	fmt.Println(pt.ToString())

	pt.Insert(2, ptm.Sequence("@"))
	fmt.Println(pt.ToString())

	pt.Insert(2, ptm.Sequence("#"))
	fmt.Println(pt.ToString())

	pt.Delete(3, 5)
	fmt.Println(pt.ToString())
	
	utils.Logger.Println("TEST 2 ENDED")
}

func main() {
	// original, _ := utils.ReadFile("../../example.txt")
	test1()
	test2()

}
