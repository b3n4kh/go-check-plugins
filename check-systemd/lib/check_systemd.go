package checksystemd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
)

// https://github.com/sensu-plugins/sensu-plugins-process-checks
var opts struct {
	WarningOver   *int64   `short:"w" long:"warning-over" value-name:"N" description:"Trigger a warning if over a number"`
	CritOver      *int64   `short:"c" long:"critical-over" value-name:"N" description:"Trigger a critical if over a number"`
	CmdPatterns   []string `short:"p" long:"pattern" value-name:"PATTERN" description:"Match a command against these patterns"`
	CmdExcludePat string   `short:"x" long:"exclude-pattern" value-name:"PATTERN" description:"Don't match against a pattern to prevent false positives"`
}

type sysState struct {
	unit        string
	load        string
	active      string
	sub         string
	description string
}

type unitState struct {
	failed     []string
	active     []string
	activating []string
	inactive   []string
}

// Do the plugin
func Do() {
	ckr := run(os.Args[1:])
	ckr.Name = "Systemd"
	ckr.Exit()
}

func run(args []string) *checkers.Checker {
	_, err := flags.ParseArgs(&opts, args)
	if err != nil {
		os.Exit(1)
	}

	sysstate, err := getSystemd()
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, err.Error())
	}
	var cmdPatRegexp []*regexp.Regexp
	for _, ptn := range opts.CmdPatterns {
		r, err := regexp.Compile(ptn)
		if err != nil {
			return checkers.NewChecker(checkers.UNKNOWN, err.Error())
		}
		cmdPatRegexp = append(cmdPatRegexp, r)
	}
	if len(cmdPatRegexp) == 0 {
		cmdPatRegexp = append(cmdPatRegexp, regexp.MustCompile(".*"))
	}
	cmdExcludePatRegexp := regexp.MustCompile(".*")
	if opts.CmdExcludePat != "" {
		r, err := regexp.Compile(opts.CmdExcludePat)
		if err != nil {
			return checkers.NewChecker(checkers.UNKNOWN, err.Error())
		}
		cmdExcludePatRegexp = r
	}
	result := checkers.OK
	var msg string

	var resultStates []sysState
	for _, reg := range cmdPatRegexp {
		for _, state := range sysstate {
			if matchState(state, reg, cmdExcludePatRegexp) {
				resultStates = append(resultStates, state)
			}
		}
	}
	count := int64(len(resultStates))
	result = getStatus(count)
	msg += fmt.Sprintf("\n%q", resultStates)
	return checkers.NewChecker(result, msg)
}

func matchState(state sysState, cmdPatRegexp *regexp.Regexp, cmdExcludePatRegexp *regexp.Regexp) bool {
	return (len(opts.CmdPatterns) == 0 || cmdPatRegexp.MatchString(state.unit)) &&
		(opts.CmdExcludePat == "" || !cmdExcludePatRegexp.MatchString(state.unit))
}

func getStatus(count int64) checkers.Status {
	if opts.CritOver != nil && count > *opts.CritOver {
		return checkers.CRITICAL
	}
	if opts.WarningOver != nil && count > *opts.WarningOver {
		return checkers.WARNING
	}
	return checkers.OK
}

func getSystemd() (sysstate []sysState, err error) {
	var states []sysState

	output, err := exec.Command("systemctl", "--failed", "--no-legend").Output()
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(output), "\n") {
		state, err := parseSysState(line)
		if err != nil {
			continue
		}
		states = append(states, state)
	}
	return states, nil
}

/*
	# Output of `systemctl list-units --all --no-legend`:
	# UNIT           LOAD   ACTIVE SUB        DESCRIPTION
	# foobar.service loaded active waiting    Description text
*/
func parseSysState(line string) (state sysState, err error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return sysState{}, errors.New("parseSysState: couldnt parse 'systemctl list-units'")
	}
	unit := fields[0]
	load := fields[1]
	active := fields[2]
	sub := fields[3]
	description := strings.Join(fields[4:], " ")

	return sysState{unit, load, active, sub, description}, nil
}
