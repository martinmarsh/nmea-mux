package test_helpers

import (
	"fmt"
	"time"
)

func UnexpectedErrorMessage(message string, err error) bool {
	unexpected := true
	if err != nil {
		err_str := err.Error()
		if err_str[:min(len(message), len(err_str))] == message {
			unexpected = false
		}
	}
	return unexpected
}

func GetMessages(m chan string) []string {
	run_for := time.NewTicker(50 * time.Millisecond)
	ret := make([]string, 0)

	for run := 4; run > 0; {
		select {
		case str := <-m:
			ret = append(ret, str)
			run = 4
		case <-run_for.C:
			run--
		}
	}

	return ret
}

func MessagesIn(expected []string, messages []string) (int, int, bool, error) {
	not_found := false
	match := -1
	match_expected := -1
	var err error = nil

	for e_index, e := range expected {
		if im, _, i_err := MessageIn(e, messages); i_err != nil {
			match_expected = e_index
			not_found = true
			match = im
			err = i_err
			break
		}

	}

	return match_expected, match, not_found, err

}

func MessageIn(expected string, messages []string) (int, bool, error) {
	not_found := true
	match := -1
	var err error = nil

	for i, m := range messages {
		len := min(len(expected), len(m))
		if m[:len] == expected[:len] {
			not_found = false
			match = i
			break
		}

	}
	if match < 0 {
		err = fmt.Errorf("expected <%s> not found in <%s>", expected, messages)

	}
	return match, not_found, err
}
