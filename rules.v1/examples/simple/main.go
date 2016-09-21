package main

import (
	"fmt"
	"github.com/metakeule/rules/rules.v1"
)

var (
	qAdd    = rules.CLI("Do you want to add a person", rules.Bool)
	qName   = rules.CLI("What is your name", rules.String)
	qGender = rules.CLI("What is your gender", rules.String, "male", "female")
	qAge    = rules.CLI("What is your age", rules.Int)
	qProf   = rules.CLI("What is your profession", rules.String)
)

type persons struct {
	p []*person
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
	return qName, (d.handleName)
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
	return qAdd, (d.handleExit)
}

func (d *person) handleProfession(val interface{}) (rules.Valuer, rules.Rule) {
	d.profession = val.(string)
	return nil, d.handleAddPerson
}

func (d *person) handleAge(val interface{}) (rules.Valuer, rules.Rule) {
	d.age = val.(int)
	if d.isAduld() {
		return qProf, (d.handleProfession)
	}
	return nil, d.handleAddPerson
}

func (d *person) handleGender(val interface{}) (rules.Valuer, rules.Rule) {
	d.gender = val.(string)
	return qAge, (d.handleAge)
}

func (d *person) handleName(val interface{}) (rules.Valuer, rules.Rule) {
	d.name = val.(string)
	return qGender, (d.handleGender)
}

func main() {
	p := &persons{}
	rules.Run(nil, p.handleAdd)
}
