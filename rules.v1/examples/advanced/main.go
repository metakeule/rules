package main

import (
	"fmt"
	"github.com/metakeule/rules/rules.v1"
	// "os"
)

var (
	base    = rules.Base([2]string{"q", "quit"})
	qAdd    = base.CLI("Do you want to add a person?", rules.Bool)
	qName   = base.CLI("What is your name?", rules.String)
	qGender = base.CLI("What is your gender?", rules.String, "male", "female")
	qAge    = base.CLI("What is your age?", rules.Int)
	qProf   = base.CLI("What is your profession?", rules.String)
)

type persons struct {
	p []*person
}

func (p *persons) addQuitHandling(r rules.Rule) rules.Rule {
	return func(val interface{}) (rules.Valuer, rules.Rule) {
		// fmt.Printf("----quit-handling for: %#v----\n", val)
		if p.isQuit(val) {
			return p.quit(val)
		}
		return r(val)
	}
}

func (p *persons) quit(val interface{}) (rules.Valuer, rules.Rule) {
	fmt.Println("aborting")
	// os.Exit(1)
	return nil, nil
}

func (p *persons) isQuit(val interface{}) bool {
	s, ok := val.(string)
	return ok && s == "q"
}

func (p *persons) print(val interface{}) (rules.Valuer, rules.Rule) {
	for _, pers := range p.p {
		fmt.Println("---------------")
		pers.print()
	}
	fmt.Println("---------------")
	return nil, nil
}

func (p *persons) handleAdd(val interface{}) (rules.Valuer, rules.Rule) {
	d := &person{all: p}
	p.p = append(p.p, d)
	return qName, p.addQuitHandling(d.handleName)
}

type person struct {
	all        *persons
	name       string
	age        int
	profession string
	gender     string
}

func (d *person) isAduld() bool {
	return d.age > 19
}

func (d *person) print() {
	if d.isAduld() {
		d.printAduld()
	} else {
		d.printChild()
	}
}

func (d *person) printAduld() {
	fmt.Printf("Name: %v\nGender: %v\nAge: %v\nProfession: %s\n", d.name, d.gender, d.age, d.profession)
}

func (d *person) printChild() {
	fmt.Printf("Name: %v\nGender: %v\nAge: %v\n", d.name, d.gender, d.age)
}

func (d *person) handleExit(val interface{}) (rules.Valuer, rules.Rule) {
	var add = val.(bool)
	if add {
		return d.all.handleAdd(nil)
	}
	return d.all.print(nil)
}

func (d *person) handleAddPerson(val interface{}) (rules.Valuer, rules.Rule) {
	return qAdd, d.all.addQuitHandling(d.handleExit)
}

func (d *person) handleProfession(val interface{}) (rules.Valuer, rules.Rule) {
	d.profession = val.(string)
	return nil, d.all.addQuitHandling(d.handleAddPerson)
}

func (d *person) handleAge(val interface{}) (rules.Valuer, rules.Rule) {
	d.age = val.(int)
	if d.isAduld() {
		return qProf, d.all.addQuitHandling(d.handleProfession)
	}
	return nil, d.all.addQuitHandling(d.handleAddPerson)
}

func (d *person) handleGender(val interface{}) (rules.Valuer, rules.Rule) {
	d.gender = val.(string)
	return qAge, d.all.addQuitHandling(d.handleAge)
}

func (d *person) handleName(val interface{}) (rules.Valuer, rules.Rule) {
	d.name = val.(string)
	return qGender, d.all.addQuitHandling(d.handleGender)
}

func main() {
	p := &persons{}
	rules.Run(nil, p.handleAdd)
}
