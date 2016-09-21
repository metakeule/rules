package rules

import (
	"bufio"
	"strings"
	//"lib"
	"fmt"
	"github.com/metakeule/fmtdate"
	"io"
	"os"
	"strconv"
)

// Valuer gets a value from somewhere and returns it
type Valuer interface {
	Value() (interface{}, error)
}

// Rule performs some action and returns the next Valuer and Ruler
type Rule func(val interface{}) (Valuer, Rule)

// Validation validates and converts the input and returns an error if data is invalid
type Validation func(input string) (output interface{}, err error)

// String validates a string
func String(input string) (output interface{}, err error) {
	return input, nil
}

// Int validates an int
func Int(input string) (output interface{}, err error) {
	return strconv.Atoi(input)
}

// Bool validates a bool
func Bool(input string) (output interface{}, err error) {
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "yes", "y":
		return true, nil
	case "no", "n":
		return false, nil
	default:
		return strconv.ParseBool(input)
	}

}

// Float64 validates an float64
func Float64(input string) (output interface{}, err error) {
	return strconv.ParseFloat(input, 64)
}

// Float32 validates an float32
func Float32(input string) (output interface{}, err error) {
	return strconv.ParseFloat(input, 32)
}

// Date validates a date
func Date(input string) (output interface{}, err error) {
	return fmtdate.ParseDate(input)
}

// DateTime validates a datetime
func DateTime(input string) (output interface{}, err error) {
	return fmtdate.ParseDateTime(input)
}

// Time validates a time
func Time(input string) (output interface{}, err error) {
	return fmtdate.ParseTime(input)
}

func CLI(question string, val Validation, options ...string) Valuer {
	return &Interactor{
		Question:   question,
		Options:    options,
		Validation: val,
		Stdout:     os.Stdout,
		Stdin:      os.Stdin,
		Stderr:     os.Stderr,
	}
}

type Interactor struct {
	Question   string
	Options    []string
	Validation Validation
	Stdout     io.Writer
	Stderr     io.Writer
	Stdin      io.Reader
}

func (c *Interactor) reportError(action string, err error) {
	fmt.Fprintf(c.Stderr, "error while %s: %s\n", action, err.Error())
}

type ValidationError string

func (v ValidationError) Error() string {
	return string(v)
}

func (c *Interactor) Value() (interface{}, error) {
	r, err := c.value()

	for ; err != nil; r, err = c.value() {
		switch err.(type) {
		case ValidationError:
		default:
			return nil, err
		}
	}

	return r, nil
}

func (c *Interactor) value() (interface{}, error) {
	var q = fmt.Sprintf("=> %s\n", c.Question)

	if len(c.Options) > 0 {
		for i, o := range c.Options {
			q += fmt.Sprintf("(%v) %s\n", i+1, o)
		}
	}
	_, err := fmt.Fprint(c.Stdout, q)

	if err != nil {
		c.reportError("writing", err)
		return nil, err
	}

	var s string

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		s = scanner.Text()
	}

	if len(c.Options) > 0 {
		i, errI := strconv.Atoi(s)
		if errI != nil {
			c.reportError("reading", errI)
			return nil, ValidationError(errI.Error())
		}
		if i <= 0 || i > len(c.Options) {
			errOpt := fmt.Errorf("invalid option: %v", i)
			c.reportError("validating", errOpt)
			return nil, ValidationError(errOpt.Error())
		}

		s = c.Options[i-1]
	}

	val, errVal := c.Validation(s)

	if errVal != nil {
		c.reportError("validating", errVal)
		return nil, ValidationError(errVal.Error())
	}

	return val, nil
}

// run always returns the last valuer and ruler that have been successful
func run(valuer Valuer, rule Rule) (vl Valuer, r Rule, err error) {
	var v interface{}
	if valuer != nil {
		v, err = valuer.Value()
		if err != nil {
			return valuer, rule, err
		}
	}
	valuer, rule = rule(v)
	return valuer, rule, nil
}

func Run(valuer Valuer, rule Rule) (err error) {
	for ; err == nil && rule != nil; valuer, rule, err = run(valuer, rule) {
		// just a loop
	}

	return err
}
