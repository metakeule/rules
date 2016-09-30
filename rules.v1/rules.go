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
	output, err = strconv.Atoi(input)
	if err != nil {
		err = fmt.Errorf("expected an integer")
	}
	return
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
		output, err = strconv.ParseBool(input)
		if err != nil {
			err = fmt.Errorf("expected 'yes' or 'no'")
		}
		return
	}

}

// Float64 validates an float64
func Float64(input string) (output interface{}, err error) {
	output, err = strconv.ParseFloat(input, 64)
	if err != nil {
		err = fmt.Errorf("expected an float64")
	}
	return
}

// Float32 validates an float32
func Float32(input string) (output interface{}, err error) {
	output, err = strconv.ParseFloat(input, 32)
	if err != nil {
		err = fmt.Errorf("expected an float32")
	}
	return
}

// Date validates a date
func Date(input string) (output interface{}, err error) {
	output, err = fmtdate.ParseDate(input)
	if err != nil {
		err = fmt.Errorf("expected a date")
	}
	return
}

// DateTime validates a datetime
func DateTime(input string) (output interface{}, err error) {
	output, err = fmtdate.ParseDateTime(input)
	if err != nil {
		err = fmt.Errorf("expected a date-time")
	}
	return
}

// Time validates a time
func Time(input string) (output interface{}, err error) {
	output, err = fmtdate.ParseTime(input)
	if err != nil {
		err = fmt.Errorf("expected a time")
	}
	return
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

func Base(options ...[2]string) interface {
	CLI(question string, val Validation, options ...string) Valuer
} {
	var b baseInteractor
	b.baseOptions = options
	return b
}

type baseInteractor struct {
	baseOptions [][2]string
	*Interactor
}

func (b baseInteractor) CLI(question string, val Validation, options ...string) Valuer {
	b.Interactor = CLI(question, val, options...).(*Interactor)
	b.Interactor.customQuestion = b.question
	return &b
}

func (b *baseInteractor) question(i *Interactor) string {
	q := i.question()
	q += "\n---------\n"
	for _, bo := range b.baseOptions {
		q += fmt.Sprintf("%s - %s\n", bo[0], bo[1])
	}

	return q
}

func (b *baseInteractor) findBaseOption(val interface{}) string {
	fmt.Printf("checking for base option: %#v\n", val)
	if s, ok := val.(string); ok {

		s = strings.ToLower(strings.TrimSpace(s))

		for _, bo := range b.baseOptions {
			if bo[0] == s {
				fmt.Printf("found: %#v\n", bo[0])
				return bo[0]
			}
		}

	}
	fmt.Println("found nothing\n")
	return ""
}

func (b *baseInteractor) optValue() (opt string, val interface{}, err error) {
	val, err = b.Interactor.value()
	opt = b.findBaseOption(val)
	return
}

func (b *baseInteractor) Value() (interface{}, error) {

	o, r, err := b.optValue()

	for ; o == "" && err != nil; o, r, err = b.optValue() {
		b.Interactor.printErr(err)
		switch err.(type) {
		case ValidationError:
		default:
			return r, err
		}
	}

	if o != "" {
		return o, nil
	}

	return r, nil
}

type Interactor struct {
	Question       string
	Options        []string
	Validation     Validation
	Stdout         io.Writer
	Stderr         io.Writer
	Stdin          io.Reader
	customQuestion func(*Interactor) string
}

func (c *Interactor) reportError(action string, err error) {
	fmt.Fprintf(c.Stderr, "error while %s: %s\n", action, err.Error())
}

type ValidationError string

func (v ValidationError) Error() string {
	return string(v)
}

func (c *Interactor) printErr(err error) {
	switch err.(type) {
	case ValidationError:
		c.reportError("validating", err)
	default:
		c.reportError("reading", err)
	}
}

func (c *Interactor) Value() (interface{}, error) {
	r, err := c.value()

	for ; err != nil; r, err = c.value() {
		c.printErr(err)
		switch err.(type) {
		case ValidationError:
		default:
			return r, err
		}
	}

	return r, nil
}

func (c *Interactor) question() string {
	var q = fmt.Sprintf("\n%s\n", strings.ToUpper(c.Question))

	if len(c.Options) > 0 {
		for i, o := range c.Options {
			q += fmt.Sprintf("%v - %s\n", i+1, o)
		}
	}
	return q
}

func (c *Interactor) value() (interface{}, error) {
	var q string
	if c.customQuestion != nil {
		q = c.customQuestion(c)
	} else {
		q = c.question()
	}

	q += "\n=> "
	_, err := fmt.Fprint(c.Stdout, q)
	if err != nil {
		c.reportError("writing", err)
		return nil, err
	}

	var originalResponse string
	var s string

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		originalResponse = scanner.Text()
		s = originalResponse
	}

	if len(c.Options) > 0 {
		i, errI := strconv.Atoi(s)
		if errI != nil {
			// c.reportError("reading", errI)
			return originalResponse, ValidationError(errI.Error())
		}
		if i <= 0 || i > len(c.Options) {
			errOpt := fmt.Errorf("invalid option: %v", i)
			// c.reportError("validating", errOpt)
			return nil, ValidationError(errOpt.Error())
		}

		s = c.Options[i-1]
	}

	val, errVal := c.Validation(s)

	if errVal != nil {
		// c.reportError("validating", errVal)
		return originalResponse, ValidationError(errVal.Error())
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
