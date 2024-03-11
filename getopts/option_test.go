package getopts

import "testing"

//Basic recognition of short options
func TestParseCase01(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "-v" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if !verbose.Passed {
		t.Fatalf("-v passed, should be true")
	}
}

//Ensure that -- stops processing flags/options
func TestParseCase02(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "--", "-v" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if verbose.Passed {
		t.Fatalf("-v passed after --, should not be true")
	}
}

//Basic recognition of long options
func TestParseCase03(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "--verbose" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if !verbose.Passed {
		t.Fatalf("--verbose passed, should be true")
	}
}

//Testing -- with long flag
func TestParseCase04(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "--", "--verbose"})
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if verbose.Passed {
		t.Fatalf("--verbose passed after --, should not be true")
	}
}

//Ensure that --flag=bool works
func TestParseCase05(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "--verbose=true" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if !verbose.Passed {
		t.Fatalf("--verbose=true should result in true for flag")
	}
}

//Ensure that short negation of flag works
func TestParseCase06(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "-v", "+v" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if verbose.Passed {
		t.Fatalf("+v after -v, should negate flag")
	}
}

//Basic processing of clump of options
func TestParseCase07(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	all := NewFlag('a', "all", "All things")
	_, err := ArgParse([]string{ "test", "-av" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if !verbose.Passed {
		t.Log("-av should set verbose")
		t.Fail()
	}

	if !all.Passed {
		t.Log("-av should set all")
		t.Fail()
	}
}

//Test --flag=bool for negation
func TestParseCase08(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	_, err := ArgParse([]string{ "test", "-v", "--verbose=false" })
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if verbose.Passed {
		t.Log("--verbose=false should negate verbose set by -v")
		t.Fail()
	}
}

//Test grouping of rest arguments
func TestParseCase09(t *testing.T) {
	resetParams()
	verbose := NewFlag('v', "verbose", "Increase verbosity")
	rest, err := ArgParse([]string{ "test", "hello", "-", "-v", "--", "-" })
	exp_rest := []Rest{
		{
			Argument:		"hello",
			AfterDashes:		false,
		},
		{
			Argument:		"-",
			AfterDashes:		false,
		},
		{
			Argument:		"-",
			AfterDashes:		true,
		},
	}
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	if !verbose.Passed {
		t.Log("-v should make verbose true")
		t.Fail()
	}
	if len(exp_rest) != len(rest) {
		t.Logf("Got %d in rest, expected %d", len(rest), len(exp_rest))
		t.Fatal()
	}

	for i, r := range exp_rest {
		if r.Argument != rest[i].Argument {
			t.Fatalf("Got %s expected %s",
				rest[i].Argument,
				r.Argument)
		}

		if r.AfterDashes != rest[i].AfterDashes {
			t.Fatalf("For %s, got dashed %v expected %v",
				r.Argument,
				rest[i].AfterDashes,
				r.AfterDashes)
		}
	}
}
