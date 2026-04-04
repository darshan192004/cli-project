package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/gookit/color"
)

type TableFormatter struct {
	results     []map[string]interface{}
	cols        []string
	colWidths   map[string]int
	terminalW   int
	maxColWidth int
	minColWidth int
}

func NewTableFormatter(results []map[string]interface{}) *TableFormatter {
	terminalW, _ := getTerminalWidth()
	if terminalW == 0 {
		terminalW = 100
	}

	tf := &TableFormatter{
		results:     results,
		terminalW:   terminalW,
		maxColWidth: 30,
		minColWidth: 8,
	}
	tf.calculateColumns()
	return tf
}

func getTerminalWidth() (int, error) {
	width := 100
	if envW := os.Getenv("COLUMNS"); envW != "" {
		fmt.Sscanf(envW, "%d", &width)
		return width, nil
	}
	return width, nil
}

func (tf *TableFormatter) calculateColumns() {
	if len(tf.results) == 0 {
		return
	}

	tf.cols = make([]string, 0)
	for key := range tf.results[0] {
		tf.cols = append(tf.cols, key)
	}

	tf.colWidths = make(map[string]int)

	for _, col := range tf.cols {
		maxLen := utf8.RuneCountInString(col)

		for _, row := range tf.results {
			if v, ok := row[col]; ok && !isNilValue(v) {
				vStr := fmt.Sprintf("%v", v)
				vLen := utf8.RuneCountInString(vStr)
				if vLen > maxLen {
					maxLen = vLen
				}
			}
		}

		if maxLen > tf.maxColWidth {
			maxLen = tf.maxColWidth
		}
		if maxLen < tf.minColWidth {
			maxLen = tf.minColWidth
		}

		tf.colWidths[col] = maxLen + 2
	}

	tf.adjustForTerminalWidth()
}

func (tf *TableFormatter) adjustForTerminalWidth() {
	sepWidth := len(tf.cols) + 1
	for _, w := range tf.colWidths {
		sepWidth += w
	}

	if sepWidth <= tf.terminalW {
		return
	}

	excess := sepWidth - tf.terminalW + 5
	numCols := len(tf.cols)

	if numCols <= 4 {
		for _, col := range tf.cols {
			if tf.colWidths[col] > tf.minColWidth+2 {
				tf.colWidths[col] -= excess / numCols
				if tf.colWidths[col] < tf.minColWidth {
					tf.colWidths[col] = tf.minColWidth
				}
			}
		}
	} else {
		sortCols := make([]string, len(tf.cols))
		copy(sortCols, tf.cols)

		for i := 0; i < len(sortCols)-1; i++ {
			for j := i + 1; j < len(sortCols); j++ {
				if tf.colWidths[sortCols[i]] < tf.colWidths[sortCols[j]] {
					sortCols[i], sortCols[j] = sortCols[j], sortCols[i]
				}
			}
		}

		reduceCols := 3
		if reduceCols > len(sortCols) {
			reduceCols = len(sortCols)
		}

		for i := 0; i < reduceCols; i++ {
			col := sortCols[i]
			if tf.colWidths[col] > tf.minColWidth+2 {
				tf.colWidths[col] = tf.minColWidth
			}
		}
	}
}

func (tf *TableFormatter) Print() {
	if len(tf.results) == 0 {
		color.Yellow.Println("No results to display")
		return
	}

	tf.printSeparator('+', '-')
	tf.printHeader()
	tf.printSeparator('+', '-')
	tf.printRows()
	tf.printSeparator('+', '-')
}

func (tf *TableFormatter) printSeparator(left, mid rune) {
	fmt.Print(string(left))
	for _, col := range tf.cols {
		fmt.Print(strings.Repeat(string(mid), tf.colWidths[col]))
		fmt.Print(string(mid))
	}
	fmt.Println()
}

func (tf *TableFormatter) printHeader() {
	fmt.Print("|")
	for _, col := range tf.cols {
		displayCol := col
		maxDisplay := tf.colWidths[col] - 2

		if utf8.RuneCountInString(col) > maxDisplay {
			runes := []rune(col)
			if maxDisplay > 3 {
				displayCol = string(runes[:maxDisplay-3]) + "..."
			} else {
				displayCol = string(runes[:maxDisplay])
			}
		}

		fmt.Printf("%s", color.Bold.Sprintf(" %-*s ", tf.colWidths[col]-2, displayCol))
		fmt.Print("|")
	}
	fmt.Println()
}

func (tf *TableFormatter) printRows() {
	nullColor := color.New(color.FgDarkGray)

	for _, row := range tf.results {
		fmt.Print("|")
		for _, col := range tf.cols {
			val := ""
			displayVal := ""

			if v, ok := row[col]; ok && !isNilValue(v) {
				val = fmt.Sprintf("%v", v)
				maxDisplay := tf.colWidths[col] - 2

				if utf8.RuneCountInString(val) > maxDisplay {
					runes := []rune(val)
					if maxDisplay > 3 {
						displayVal = string(runes[:maxDisplay-3]) + "..."
					} else {
						displayVal = string(runes[:maxDisplay])
					}
				} else {
					displayVal = val
				}
			} else {
				val = "NULL"
				displayVal = "NULL"
			}

			if val == "NULL" {
				fmt.Printf(" %-*s ", tf.colWidths[col]-2, nullColor.Sprint(displayVal))
			} else {
				fmt.Printf(" %-*s ", tf.colWidths[col]-2, displayVal)
			}
			fmt.Print("|")
		}
		fmt.Println()
	}
}

func isNilValue(v interface{}) bool {
	if v == nil {
		return true
	}
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Interface, reflect.Func:
		return val.IsNil()
	}
	return false
}

func PrintTable(results []map[string]interface{}) {
	tf := NewTableFormatter(results)
	tf.Print()
}
