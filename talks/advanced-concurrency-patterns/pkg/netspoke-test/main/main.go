package main

import (
	"fmt"
)

// Direction is an enumeration of possible movement directions.
type Direction int

const (
	UP Direction = iota
	DOWN
	LEFT
	RIGHT
)

// The2048Bonacci represents the game state and logic.
type The2048Bonacci struct {
	gameArea [][]int
	width    int
	height   int
}

// NewThe2048Bonacci initializes a new The2048Bonacci object.
func NewThe2048Bonacci(gameArea [][]int) *The2048Bonacci {
	return &The2048Bonacci{
		gameArea: gameArea,
		width:    len(gameArea[0]),
		height:   len(gameArea),
	}
}

// GetTile returns the tile value at the specified coordinates.
func (t *The2048Bonacci) GetTile(x, y int) int {
	return t.gameArea[y][x]
}

// SetTile sets the tile value at the specified coordinates.
func (t *The2048Bonacci) SetTile(x, y int, fibValue int) {
	t.gameArea[y][x] = fibValue
}

// GetDescription returns a string representation of the game area.
func (t *The2048Bonacci) GetDescription() string {
	strLines := make([]string, len(t.gameArea))
	for i, line := range t.gameArea {
		strLine := ""
		for _, fibVal := range line {
			strLine += fmt.Sprintf("%2d ", fibVal)
		}
		strLines[i] = strLine
	}
	return fmt.Sprintf("%s\n", strLines)
}

func (t *The2048Bonacci) moveRight() {
	rowMap := t.makeRowMap()
	for rowIdx, rowNums := range rowMap {
		newRowNums := make([]int, len(rowNums))
		for i := len(rowNums) - 1; i > 0; i-- {
			if rowNums[i] == 0 {
				continue
			}
			temp := rowNums[i-1]
			rowNums[i-1] = 0
			newRowNums[i] = rowNums[i] + temp
		}
		for i := len(newRowNums) - 1; i > 0; i-- {
			if newRowNums[i] == 0 && i != 0 {
				newRowNums[i] = newRowNums[i-1]
				newRowNums[i-1] = 0
			}
		}
		rowMap[rowIdx] = newRowNums
	}
	t.gameArea = getArrayFromRowMap(rowMap)
}

func (t *The2048Bonacci) moveLeft() {
	rowMap := t.makeRowMap()
	for rowIdx, rowNums := range rowMap {
		newRowNums := make([]int, len(rowNums))
		for i := 0; i < len(rowNums)-1; i++ {
			if rowNums[i] == 0 {
				continue
			}
			temp := rowNums[i+1]
			rowNums[i+1] = 0
			newRowNums[i] = rowNums[i] + temp
		}
		for i := 0; i < len(rowNums)-1; i++ {
			if newRowNums[i] == 0 {
				newRowNums[i] = newRowNums[i+1]
				newRowNums[i+1] = 0
			}
		}
		rowMap[rowIdx] = newRowNums
	}
	t.gameArea = getArrayFromRowMap(rowMap)
}

func getArrayFromRowMap(rowMap map[int][]int) [][]int {
	result := make([][]int, len(rowMap))
	for rowIdx, rowVals := range rowMap {
		result[rowIdx] = rowVals
	}
	return result
}

func getArrayFromColMap(colMap map[int][]int) [][]int {
	numRows := len(colMap)
	numCols := len(colMap[0])

	// Initialize transposed 2D slice
	transposed := make([][]int, numCols)
	for i := range transposed {
		transposed[i] = make([]int, numRows)
	}

	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			transposed[j][i] = colMap[i][j]
		}
	}

	return transposed
}

func (t *The2048Bonacci) moveUpward() {
	colMap := t.makeColMap()
	for colIdx, colNums := range colMap {
		newColNums := make([]int, len(colNums))
		for i := 0; i < len(newColNums); i++ {
			if colNums[i] == 0 {
				continue
			}
			if i == len(colNums)-1 {
				newColNums[i] = colNums[i]
				continue
			}
			temp := colNums[i+1]
			colNums[i+1] = 0
			newColNums[i] = colNums[i] + temp
		}
		for i := 0; i < len(newColNums)-1; i++ {
			if newColNums[i] == 0 && i != 0 {
				newColNums[i] = newColNums[i+1]
				newColNums[i+1] = 0
			}
		}
		colMap[colIdx] = newColNums
	}
	t.gameArea = getArrayFromColMap(colMap)
}

func (t *The2048Bonacci) moveDownward() {
	colMap := t.makeColMap()
	for colIdx, colNums := range colMap {
		newColNums := make([]int, len(colNums))
		for i := len(colNums) - 1; i >= 0; i-- {
			if colNums[i] == 0 {
				continue
			}
			if i == 0 && colNums[i] != 0 {
				newColNums[i] = colNums[i]
				continue
			}
			temp := colNums[i-1]
			colNums[i-1] = 0
			newColNums[i] = colNums[i] + temp
		}
		for i := len(newColNums) - 1; i > 0; i-- {
			if newColNums[i] == 0 && i != 0 {
				newColNums[i] = newColNums[i-1]
				newColNums[i-1] = 0
			}
		}
		colMap[colIdx] = newColNums
	}
	t.gameArea = getArrayFromColMap(colMap)
}

func (t *The2048Bonacci) makeRowMap() map[int][]int {
	rowMap := make(map[int][]int)
	for i, numX := range t.gameArea {
		rowNums := []int{}
		rowNums = append(rowNums, numX...)
		rowMap[i] = rowNums
	}
	return rowMap
}

func (t *The2048Bonacci) makeColMap() map[int][]int {
	rowMap := make(map[int][]int)
	for _, numX := range t.gameArea {
		for j, numY := range numX {
			rowMap[j] = append(rowMap[j], numY)
		}
	}
	return rowMap
}

func main() {
	// Example game area.
	gameArea := [][]int{
		{1, 3, 5, 7},
		{2, 4, 8, 16},
		{32, 64, 128, 256},
		{512, 1024, 2048, 4096},
	}

	t := NewThe2048Bonacci(gameArea)

	// Example usage.
	fmt.Println(t.GetDescription())

	fmt.Println("Moving downward")
	t.moveDownward()
	fmt.Println(t.GetDescription())

	fmt.Println("Moving left")
	t.moveLeft()
	fmt.Println(t.GetDescription())

	fmt.Println("Moving right")
	t.moveRight()
	fmt.Println(t.GetDescription())

	fmt.Println("Moving upwards")
	t.moveUpward()
	fmt.Println(t.GetDescription())
}
