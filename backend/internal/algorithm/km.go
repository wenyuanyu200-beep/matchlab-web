package algorithm

import (
	"errors"
	"sort"
)

var ErrInvalidMatrix = errors.New("weight matrix rows must have equal length")

type Pair struct {
	Row    int `json:"row"`
	Column int `json:"column"`
	Weight int `json:"weight"`
}

func MaximumWeightMatching(weights [][]int) ([]Pair, int, error) {
	rows, columns, err := matrixSize(weights)
	if err != nil || rows == 0 || columns == 0 {
		return []Pair{}, 0, err
	}
	paddedColumns := columns
	if rows > paddedColumns {
		paddedColumns = rows
	}

	// Hungarian minimization with cost=-weight. Dummy columns have zero weight
	// and represent unmatched surplus rows in a rows>columns matrix.
	u := make([]int, rows+1)
	v := make([]int, paddedColumns+1)
	p := make([]int, paddedColumns+1)
	way := make([]int, paddedColumns+1)
	const infinity = int(^uint(0)>>1) / 4

	for i := 1; i <= rows; i++ {
		p[0] = i
		minValues := make([]int, paddedColumns+1)
		used := make([]bool, paddedColumns+1)
		for j := 1; j <= paddedColumns; j++ {
			minValues[j] = infinity
		}
		column := 0
		for {
			used[column] = true
			row := p[column]
			delta := infinity
			nextColumn := 0
			for j := 1; j <= paddedColumns; j++ {
				if used[j] {
					continue
				}
				weight := 0
				if j <= columns {
					weight = weights[row-1][j-1]
				}
				cost := -weight - u[row] - v[j]
				if cost < minValues[j] {
					minValues[j] = cost
					way[j] = column
				}
				if minValues[j] < delta {
					delta = minValues[j]
					nextColumn = j
				}
			}
			for j := 0; j <= paddedColumns; j++ {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else if j > 0 {
					minValues[j] -= delta
				}
			}
			column = nextColumn
			if p[column] == 0 {
				break
			}
		}
		for {
			previous := way[column]
			p[column] = p[previous]
			column = previous
			if column == 0 {
				break
			}
		}
	}

	pairs := make([]Pair, 0, min(rows, columns))
	total := 0
	for column := 1; column <= columns; column++ {
		if p[column] == 0 {
			continue
		}
		row := p[column] - 1
		weight := weights[row][column-1]
		pairs = append(pairs, Pair{Row: row, Column: column - 1, Weight: weight})
		total += weight
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Row < pairs[j].Row })
	return pairs, total, nil
}

func GreedyMatching(weights [][]int) ([]Pair, int, error) {
	rows, columns, err := matrixSize(weights)
	if err != nil || rows == 0 || columns == 0 {
		return []Pair{}, 0, err
	}
	used := make([]bool, columns)
	pairs := make([]Pair, 0, min(rows, columns))
	total := 0
	for row := 0; row < rows; row++ {
		bestColumn := -1
		for column := 0; column < columns; column++ {
			if used[column] {
				continue
			}
			if bestColumn == -1 || weights[row][column] > weights[row][bestColumn] {
				bestColumn = column
			}
		}
		if bestColumn == -1 {
			break
		}
		used[bestColumn] = true
		weight := weights[row][bestColumn]
		pairs = append(pairs, Pair{Row: row, Column: bestColumn, Weight: weight})
		total += weight
	}
	return pairs, total, nil
}

func matrixSize(weights [][]int) (int, int, error) {
	if len(weights) == 0 {
		return 0, 0, nil
	}
	columns := len(weights[0])
	for _, row := range weights[1:] {
		if len(row) != columns {
			return 0, 0, ErrInvalidMatrix
		}
	}
	return len(weights), columns, nil
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
