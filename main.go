package main

import (
    "fmt"
    "runtime"
)

type Matrix struct {
    rows    int
    cols    int
    data    [][]float64
}

func (m *Matrix) multiply(b *Matrix, block_size int) *Matrix {
    c := &Matrix{rows: m.rows, cols: b.cols}
    c.data = make([][]float64, m.rows)
    for i := range c.data {
        c.data[i] = make([]float64, b.cols)
    }

    num_blocks := m.rows / block_size
    if m.rows%block_size != 0 {
        num_blocks++
    }

    num_procs := runtime.NumCPU()
    if num_procs > num_blocks {
        num_procs = num_blocks
    }

    ch := make(chan *MatrixBlock, num_blocks*num_procs)

    for i := 0; i < num_blocks; i++ {
        for j := 0; j < num_procs; j++ {
            go func(i, j int) {
                block_a := &MatrixBlock{
                    start_row: i * block_size,
                    end_row:   (i + 1) * block_size,
                    start_col: 0,
                    end_col:   m.cols,
                    data:     m.data,
                }
                block_b := &MatrixBlock{
                    start_row: 0,
                    end_row:   b.rows,
                    start_col: j * block_size,
                    end_col:   (j + 1) * block_size,
                    data:     b.data,
                }
                result := block_a.multiply(block_b)
                ch <- result
            }(i, j)
        }
    }

    for i := 0; i < num_blocks*num_procs; i++ {
        block := <-ch
        c.data[block.start_row][block.start_col] = block.data[0][0]
    }

    return c
}

type MatrixBlock struct {
    start_row int
    end_row   int
    start_col int
    end_col   int
    data    [][]float64
}

func (b *MatrixBlock) multiply(a *MatrixBlock) *MatrixBlock {
    result := &MatrixBlock{
        start_row: b.start_row,
        end_row:   b.end_row,
        start_col: b.start_col,
        end_col:   b.end_col,
        data:     make([][]float64, b.end_row-b.start_row),
    }
    for i := range result.data {
        result.data[i] = make([]float64, b.end_col-b.start_col)
    }

    for i := b.start_row; i < b.end_row; i++ {
        for j := b.start_col; j < b.end_col; j++ {
            sum := 0.0
            for k := a.start_col; k < a.end_col; k++ {
                sum += a.data[i][k] * b.data[k][j]
            }
            result.data[i-b.start_row][j-b.start_col] = sum
        }
    }

    return result
}

func main() {
    m1 := Matrix{
        rows: 1000,
        cols: 1000,
        data: make([][]float64, 1000),
    }
    for i := range m1.data {
        m1.data[i] = make([]float64, 1000)
        for j := range m1.data[i] {
            m1.data[i][j] = rand.Float64()
        }
    }

    m2 := Matrix{
        rows: 1000,
        cols: 1000,
        data: make([][]float64, 1000),
    }
    for i := range m2.data {
        m2.data[i] = make([]float64, 1000)
        for j := range m2.data[i] {
            m2.data[i][j] = rand.Float64()
        }
    }

    block_size := 100

    result := m1.multiply(&m2, block_size)

    fmt.Printf("Result matrix:\n")
    for i := 0; i < result.rows; i++ {
        for j := 0; j < result.cols; j++ {
            fmt.Printf("%f ", result.data[i][j])
        }
        fmt.Println()
    }
}
