package cmd

import "testing"

func TestParseCmd(t *testing.T) {
	t.Parallel()

	_, err := parseCmd([]string{"mfeeder"})
	if err == nil {
		t.Errorf("parse should fail without command")
	}

	_, err = parseCmd([]string{"mfeeder", "foo"})
	if err == nil {
		t.Errorf("parse should fail with invalid command")
	}
}

func TestValidateDay(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateDay([]string{"mfeeder", "day", "some", "day"}, &c)
	if err == nil {
		t.Errorf("validateDay should fail with more than 3 args")
	}

	err = validateDay([]string{"mfeeder", "day", "23-32"}, &c)
	if err == nil {
		t.Errorf("validateDay should fail with invalid date")
	}

	err = validateDay([]string{"mfeeder", "day", "01-01"}, &c)
	if err != nil {
		t.Errorf("validateDay should not fail with valid args")
	}

	if c.cmd != "day" {
		t.Errorf("validateDay should set cmd to day")
	}

	if len(c.args) != 1 || c.args[0] != "01-01" {
		t.Errorf("validateDay should set args to date")
	}
}

func TestValidateEx(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateEx([]string{"mfeeder", "ex", "something"}, &c)
	if err == nil {
		t.Errorf("validateEx should fail with less then 4 args")
	}

	err = validateEx([]string{"mfeeder", "ex", "something", "else", "in", "the", "args"}, &c)
	if err == nil {
		t.Errorf("validateEx should fail with more then 4 args")
	}

	err = validateEx([]string{"mfeeder", "ex", "opt", "something"}, &c)
	if err == nil {
		t.Errorf("validateEx should fail with invalid option")
	}

	err = validateEx([]string{"mfeeder", "ex", "add", "something"}, &c)
	if err != nil {
		t.Errorf("validateEx should not fail with valid option")
	}

	if c.cmd != "ex" {
		t.Errorf("validateEx should set cmd to ex")
	}

	if len(c.opt) != 1 || c.opt["add"] != "something" {
		t.Errorf("validateEx should set options to corresponding input")
	}
}

func TestValidateSystem(t *testing.T) {
	t.Parallel()
	c := initCmd()
	err := validateSystem([]string{"mfeeder", "sys", "try"}, &c)
	if err == nil {
		t.Errorf("validateSystem should fail with invalid args")
	}

	err = validateSystem([]string{"mfeeder", "sys"}, &c)
	if err != nil {
		t.Errorf("validateSystem should not fail with valid args")
	}

	if c.cmd != "sys" {
		t.Errorf("validateSystem should set cmd to sys")
	}
}

func TestValidateGet(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateGet([]string{"mfeeder", "get", "try", "get"}, &c)
	if err == nil {
		t.Errorf("validateGet should fail with more than 3 args")
	}

	err = validateGet([]string{"mfeeder", "get", "try"}, &c)
	if err == nil {
		t.Errorf("validateGet should fail with less than 3 args")
	}

	err = validateGet([]string{"mfeeder", "get", "try"}, &c)
	if err == nil {
		t.Errorf("validateGet should fail with invalid args")
	}

	err = validateGet([]string{"mfeeder", "get", "ex"}, &c)
	if err != nil {
		t.Errorf("validateGet should not fail with valid args")
	}

	if c.cmd != "get" {
		t.Errorf("validateGet should set cmd to get")
	}
}

func initCmd() cmd {
	return cmd{
		opt:  make(map[string]string),
		args: make([]string, 0),
	}
}
