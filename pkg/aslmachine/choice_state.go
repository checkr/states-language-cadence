package aslmachine

import (
	"fmt"
	"time"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence/workflow"
)

//
// ChoiceState allows you to specify which state to go to next based on rules related to
// this states inputs.
//
//	"Choice": {
//		"Type": "Choice",
//		"Choices": [
//			{
//			  "Variable": "$.value",
//			  "NumericEquals": 0,
//			  "Next": "ValueIsZero"
//			},
//			{
//			  "And": [
//				{
//				  "Variable": "$.value",
//				  "NumericGreaterThanEquals": 20.5
//				},
//				{
//				  "Variable": "$.value",
//				  "NumericLessThan": 30
//				}
//			  ],
//			  "Next": "ValueInTwenties"
//			}
//		],
//		"Default": "DefaultState"
//	}

type ChoiceState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`

	Default *string `json:",omitempty"` // Default State if no choices match

	Choices []*Choice `json:",omitempty"`
}

type Choice struct {
	ChoiceRule

	Next *string `json:",omitempty"`
}

type ChoiceRule struct {
	Variable *jsonpath.Path `json:",omitempty"`

	StringEquals            *string `json:",omitempty"`
	StringLessThan          *string `json:",omitempty"`
	StringGreaterThan       *string `json:",omitempty"`
	StringLessThanEquals    *string `json:",omitempty"`
	StringGreaterThanEquals *string `json:",omitempty"`

	NumericEquals            *float64 `json:",omitempty"`
	NumericLessThan          *float64 `json:",omitempty"`
	NumericGreaterThan       *float64 `json:",omitempty"`
	NumericLessThanEquals    *float64 `json:",omitempty"`
	NumericGreaterThanEquals *float64 `json:",omitempty"`

	BooleanEquals *bool `json:",omitempty"`

	TimestampEquals            *time.Time `json:",omitempty"`
	TimestampLessThan          *time.Time `json:",omitempty"`
	TimestampGreaterThan       *time.Time `json:",omitempty"`
	TimestampLessThanEquals    *time.Time `json:",omitempty"`
	TimestampGreaterThanEquals *time.Time `json:",omitempty"`

	And []*ChoiceRule `json:",omitempty"`
	Or  []*ChoiceRule `json:",omitempty"`
	Not *ChoiceRule   `json:",omitempty"`
}

func (s *ChoiceState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	next := chooseNextState(input, s.Default, s.Choices)
	if next == nil {
		return nil, nil, fmt.Errorf("choice state error: no choice found")
	}
	return input, next, nil
}

func (s *ChoiceState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(
		s,
		processInputOutput(
			s.InputPath,
			s.OutputPath,
			s.process,
		),
	)(ctx, input)
}

func chooseNextState(input interface{}, defaultChoice *string, choices []*Choice) *string {
	for _, choice := range choices {
		if choiceRulePositive(input, &choice.ChoiceRule) {
			return choice.Next
		}
	}
	return defaultChoice
}

func choiceRulePositive(input interface{}, cr *ChoiceRule) bool {
	if cr.And != nil {
		for _, a := range cr.And {
			// if any choices have false then return false
			if !choiceRulePositive(input, a) {
				return false
			}
		}
		return true
	}

	if cr.Or != nil {
		for _, a := range cr.Or {
			// if any choices have true then return true
			if choiceRulePositive(input, a) {
				return true
			}
		}
		return false
	}

	if cr.Not != nil {
		return !choiceRulePositive(input, cr.Not)
	}

	if cr.StringEquals != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr == *cr.StringEquals
	}

	if cr.StringLessThan != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr < *cr.StringLessThan
	}

	if cr.StringGreaterThan != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr > *cr.StringGreaterThan
	}

	if cr.StringLessThanEquals != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr <= *cr.StringLessThanEquals
	}

	if cr.StringGreaterThanEquals != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr >= *cr.StringGreaterThanEquals
	}

	// NUMBERs
	if cr.NumericEquals != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum == *cr.NumericEquals
	}

	if cr.NumericLessThan != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum < *cr.NumericLessThan
	}

	if cr.NumericGreaterThan != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum > *cr.NumericGreaterThan
	}

	if cr.NumericLessThanEquals != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum <= *cr.NumericLessThanEquals
	}

	if cr.NumericGreaterThanEquals != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum >= *cr.NumericGreaterThanEquals
	}

	if cr.BooleanEquals != nil {
		vbool, err := cr.Variable.GetBool(input)
		if err != nil {
			return false
		}
		return *vbool == *cr.BooleanEquals
	}

	if cr.TimestampEquals != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return *vtime == *cr.TimestampEquals
	}

	if cr.TimestampLessThan != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return vtime.Before(*cr.TimestampLessThan)
	}

	if cr.TimestampGreaterThan != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return vtime.After(*cr.TimestampGreaterThan)
	}

	if cr.TimestampLessThanEquals != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return *vtime == *cr.TimestampLessThanEquals || vtime.Before(*cr.TimestampLessThanEquals)
	}

	if cr.TimestampGreaterThanEquals != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return *vtime == *cr.TimestampGreaterThanEquals || vtime.After(*cr.TimestampGreaterThanEquals)
	}

	return false
}

// VALIDATION LOGIC

func (s *ChoiceState) Validate() error {
	s.SetType(to.Strp("Choice"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %w", errorPrefix(s), err)
	}

	if len(s.Choices) == 0 {
		return fmt.Errorf("%v must have Choices", errorPrefix(s))
	}

	for _, c := range s.Choices {
		err := validateChoice(c)
		if err != nil {
			return fmt.Errorf("%v %w", errorPrefix(s), err)
		}
	}

	return nil
}

func validateChoice(c *Choice) error {
	if c.Next == nil {
		return fmt.Errorf("Choice must have Next")
	}

	allChoiceRules := recursiveAllChoiceRule(&c.ChoiceRule)

	for _, cr := range allChoiceRules {
		if err := validateChoiceRule(cr); err != nil {
			return err
		}
	}

	return nil
}

func recursiveAllChoiceRule(c *ChoiceRule) []*ChoiceRule {
	if c == nil {
		return []*ChoiceRule{}
	}

	crs := []*ChoiceRule{c}

	if c.Not != nil {
		crs = append(crs, c.Not)
	}

	if c.And != nil {
		for _, cr := range c.And {
			crs = append(crs, recursiveAllChoiceRule(cr)...)
		}
	}

	if c.Or != nil {
		for _, cr := range c.Or {
			crs = append(crs, recursiveAllChoiceRule(cr)...)
		}
	}

	return crs
}

func validateChoiceRule(c *ChoiceRule) error {
	// Exactly One Comparison Operator
	allComparisonOperators := []bool{
		c.Not != nil,
		c.And != nil,
		c.Or != nil,
		c.StringEquals != nil,
		c.StringLessThan != nil,
		c.StringGreaterThan != nil,
		c.StringLessThanEquals != nil,
		c.StringGreaterThanEquals != nil,
		c.NumericEquals != nil,
		c.NumericLessThan != nil,
		c.NumericGreaterThan != nil,
		c.NumericLessThanEquals != nil,
		c.NumericGreaterThanEquals != nil,
		c.BooleanEquals != nil,
		c.TimestampEquals != nil,
		c.TimestampLessThan != nil,
		c.TimestampGreaterThan != nil,
		c.TimestampLessThanEquals != nil,
		c.TimestampGreaterThanEquals != nil,
	}

	count := 0
	for _, co := range allComparisonOperators {
		if co {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("Not Exactly One comparison Operator")
	}

	// Variable must be defined, UNLESS AND NOT OR, in which case error if defined
	notAndOr := c.Not != nil || c.And != nil || c.Or != nil

	if notAndOr {
		if c.Variable != nil {
			return fmt.Errorf("Variable defined with Not And Or defined")
		}
	} else {
		if c.Variable == nil {
			return fmt.Errorf("Variable Not defined")
		}
	}

	if c.And != nil && len(c.And) == 0 {
		return fmt.Errorf("And Must have elements")
	}

	if c.Or != nil && len(c.Or) == 0 {
		return fmt.Errorf("Or Must have elements")
	}

	return nil
}

func (s *ChoiceState) SetType(t *string) {
	s.Type = t
}

func (s *ChoiceState) GetType() *string {
	return s.Type
}
