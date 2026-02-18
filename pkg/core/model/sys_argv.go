package model

import (
	"fmt"
	"strconv"
	"time"
)

type SysArgVals map[string]string

func SysArgRealnameAndNormalizedValue(
	name string,
	sysArgPrefix string,
	value string) (realname string, normalizedValue string, err error) {

	if len(name) < len(sysArgPrefix) || name[0:len(sysArgPrefix)] != sysArgPrefix {
		return
	}
	raw := name[len(sysArgPrefix):]
	if raw == SysArgNameDelay {
		_, parseErr := strconv.ParseFloat(value, 64)
		if parseErr == nil {
			value += "s"
		}
		return name, value, nil
	} else if raw == SysArgNameDelayEnvApplyPolicy {
		if value != SysArgValueDelayEnvApplyPolicyApply {
			return "", "", fmt.Errorf("[Args.SysArgRealname] %s: the value of sys arg '%s' could only be '%s'",
				name, SysArgNameDelayEnvApplyPolicy, SysArgValueDelayEnvApplyPolicyApply)
		}
		return name, value, nil
	} else if raw == SysArgNameError {
		if value != SysArgValueOK {
			return "", "", fmt.Errorf("[Args.SysArgRealname] %s: the value of sys arg '%s' could only be '%s'",
				name, SysArgNameError, SysArgValueOK)
		}
		return name, value, nil
	} else {
		return "", "", fmt.Errorf("[Args.SysArgRealname] %s: only sys args could have '%s' prefix, '%s' is not sys arg",
			name, sysArgPrefix, raw)
	}
}

func (self SysArgVals) GetDelayStr() string {
	return self[SysArgNameDelay]
}

func (self SysArgVals) IsDelay() bool {
	return len(self[SysArgNameDelay]) != 0
}

func (self SysArgVals) GetDelayDuration() (time.Duration, error) {
	delayDur := self[SysArgNameDelay]
	dur, err := time.ParseDuration(delayDur)
	if err != nil {
		return 0, &ArgValErrWrongType{
			fmt.Sprintf("[Cmd.AsyncExecute] sys arg '%s = %s' is not valid golang duration format", SysArgNameDelay, delayDur),
			SysArgNameDelay, delayDur, "golang duration format", err,
		}
	}
	return dur, nil
}

func (self SysArgVals) IsDelayEnvEarlyApply() bool {
	return self[SysArgNameDelayEnvApplyPolicy] == SysArgValueDelayEnvApplyPolicyApply
}

func (self SysArgVals) AllowError() bool {
	return self[SysArgNameError] == SysArgValueOK
}

// TODO: put sys arg names into env strs
const (
	SysArgNameDelay               string = "delay"
	SysArgNameDelayEnvApplyPolicy string = "env"
	SysArgNameError               string = "err"

	SysArgValueDelayEnvApplyPolicyApply string = "apply"
	SysArgValueOK                       string = "ok"
)
