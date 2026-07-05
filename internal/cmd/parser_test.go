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

	_, err = parseCmd([]string{"mfeeder", "--help", "extra"})
	if err == nil {
		t.Errorf("parse should fail when help receives extra arguments")
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

func TestValidateDayRejectsMultipleDates(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateDay([]string{"mfeeder", "day", "01-01", "01-02"}, &c)
	if err == nil {
		t.Fatal("validateDay should fail when more than one date is provided")
	}
}

func TestValidateDayRejectsConflictingFormats(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateDay([]string{"mfeeder", "day", "-e", "-p"}, &c)
	if err == nil {
		t.Fatal("validateDay should fail when exe and project formats are both requested")
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

func TestValidateExRejectsEmptyValue(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateEx([]string{"mfeeder", "ex", "add", ""}, &c)
	if err == nil {
		t.Fatal("validateEx should fail with empty value")
	}
}

func TestValidateGet(t *testing.T) {
	t.Parallel()
	c := initCmd()

	err := validateGet([]string{"mfeeder", "get", "try", "get"}, &c)
	if err == nil {
		t.Errorf("validateGet should fail with more than 3 args")
	}

	err = validateGet([]string{"mfeeder", "get"}, &c)
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
