package cmd

import (
	"context"
	"fmt"
	"strings"

	"dataset-cli/internal/database"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gookit/color"
)

type Condition struct {
	Column    string
	Operator  string
	Value     string
	Connector string
}

type ConditionBuilder struct {
	db         database.Backend
	tableName  string
	columns    []string
	conditions []Condition
}

func NewConditionBuilder(db database.Backend, tableName string) *ConditionBuilder {
	return &ConditionBuilder{
		db:         db,
		tableName:  tableName,
		conditions: make([]Condition, 0),
	}
}

func (cb *ConditionBuilder) LoadColumns() error {
	info, err := cb.db.GetTableInfo(context.Background(), cb.tableName)
	if err != nil {
		return err
	}

	for _, col := range info.Columns {
		cb.columns = append(cb.columns, col.Name)
	}
	return nil
}

func (cb *ConditionBuilder) BuildConditions() (string, error) {
	for {
		condition, addMore := cb.addCondition()
		if condition != nil {
			cb.conditions = append(cb.conditions, *condition)
		}

		if !addMore {
			break
		}
	}

	return cb.buildWhereClause(), nil
}

func (cb *ConditionBuilder) addCondition() (*Condition, bool) {
	cond := &Condition{}

	var selectedColumn string
	qsColumn := &survey.Select{
		Message: "Select column to filter:",
		Options: cb.columns,
	}
	if err := survey.AskOne(qsColumn, &selectedColumn); err != nil {
		return nil, false
	}
	cond.Column = selectedColumn

	operators := []string{
		"= (equals)",
		"!= (not equals)",
		"> (greater than)",
		"< (less than)",
		">= (greater or equal)",
		"<= (less or equal)",
		"LIKE (contains)",
		"IS NULL (is empty)",
		"IS NOT NULL (is not empty)",
	}

	var selectedOp string
	qsOp := &survey.Select{
		Message: "Select operator:",
		Options: operators,
	}
	if err := survey.AskOne(qsOp, &selectedOp); err != nil {
		return nil, false
	}

	opDisplay := strings.Split(selectedOp, " ")[0]
	cond.Operator = opDisplay

	if opDisplay != "IS" && opDisplay != "IS NOT" {
		value := cb.getColumnValues(selectedColumn)

		if len(value) <= 20 {
			var selectedValue string
			qsValue := &survey.Select{
				Message: "Select value:",
				Options: append(value, "Enter custom value..."),
			}
			if err := survey.AskOne(qsValue, &selectedValue); err != nil {
				return nil, false
			}

			if selectedValue == "Enter custom value..." {
				qsCustom := &survey.Input{
					Message: "Enter custom value:",
				}
				survey.AskOne(qsCustom, &cond.Value)
			} else {
				cond.Value = selectedValue
			}
		} else {
			qsCustom := &survey.Input{
				Message: "Enter value (or search term for LIKE):",
			}
			survey.AskOne(qsCustom, &cond.Value)
		}
	}

	var connector string
	if len(cb.conditions) > 0 {
		qsConn := &survey.Select{
			Message: "Add another condition with:",
			Options: []string{"AND", "OR"},
			Default: "AND",
		}
		survey.AskOne(qsConn, &connector)
		cond.Connector = connector
	}

	var addMore bool
	qsMore := &survey.Confirm{
		Message: "Add another condition?",
		Default: false,
	}
	survey.AskOne(qsMore, &addMore)

	return cond, addMore
}

func (cb *ConditionBuilder) getColumnValues(column string) []string {
	query := fmt.Sprintf("SELECT DISTINCT \"%s\" FROM \"%s\" WHERE \"%s\" IS NOT NULL ORDER BY \"%s\" LIMIT 50",
		column, cb.tableName, column, column)

	results, err := cb.db.Execute(context.Background(), query)
	if err != nil {
		return []string{}
	}

	var values []string
	for _, row := range results {
		val := row[column]
		if val == nil {
			continue
		}
		strVal := fmt.Sprintf("%v", val)
		if len(strVal) > 50 {
			strVal = strVal[:47] + "..."
		}
		values = append(values, strVal)
	}

	return values
}

func (cb *ConditionBuilder) buildWhereClause() string {
	if len(cb.conditions) == 0 {
		return ""
	}

	var clauses []string
	for i, cond := range cb.conditions {
		clause := cb.formatCondition(cond)

		if i > 0 {
			clause = fmt.Sprintf("%s %s", cond.Connector, clause)
		}

		clauses = append(clauses, clause)
	}

	return strings.Join(clauses, " ")
}

func (cb *ConditionBuilder) formatCondition(cond Condition) string {
	quotedColumn := fmt.Sprintf("\"%s\"", cond.Column)

	switch cond.Operator {
	case "IS":
		return fmt.Sprintf("%s IS NULL", quotedColumn)
	case "IS NOT":
		return fmt.Sprintf("%s IS NOT NULL", quotedColumn)
	case "LIKE":
		return fmt.Sprintf("%s LIKE '%%%s%%'", quotedColumn, cond.Value)
	default:
		if isNumeric(cond.Value) {
			return fmt.Sprintf("%s %s %s", quotedColumn, cond.Operator, cond.Value)
		}
		return fmt.Sprintf("%s %s '%s'", quotedColumn, cond.Operator, cond.Value)
	}
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			if c != '.' {
				return false
			}
		}
	}
	return true
}

func (cb *ConditionBuilder) PrintConditions() {
	if len(cb.conditions) == 0 {
		color.Yellow.Println("No conditions applied (showing all data)")
		return
	}

	color.Cyan.Println("\nApplied conditions:")
	for i, cond := range cb.conditions {
		if i > 0 {
			color.Magenta.Printf(" %s ", cond.Connector)
		}
		color.Green.Printf("%s %s %s", cond.Column, cond.Operator, cond.Value)
	}
	fmt.Println()
}

func PrintConditionBuilder(db database.Backend, tableName string) (string, error) {
	cb := NewConditionBuilder(db, tableName)

	if err := cb.LoadColumns(); err != nil {
		return "", err
	}

	color.Bold.Println("\n=== Condition Builder ===")
	color.Gray.Println("Build filter conditions step by step\n")

	whereClause, err := cb.BuildConditions()
	if err != nil {
		return "", err
	}

	cb.PrintConditions()

	return whereClause, nil
}
