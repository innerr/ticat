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
	value string) (realname string, normalizedValue string) {

	if len(name) < len(sysArgPrefix) || name[0:len(sysArgPrefix)] != sysArgPrefix {
		return
	}
	raw := name[len(sysArgPrefix):]
	if raw == SysArgNameDelay {
		_, err := strconv.ParseFloat(value, 64)
		if err == nil {
			value += "s"
		}
		return name, value
	} else if raw == SysArgNameDelayEnvApplyPolicy {
		if value != SysArgValueDelayEnvApplyPolicyApply {
			panic(fmt.Errorf("[Args.SysArgRealname] %s: the value of sys arg '%s' could only be '%s'",
				name, SysArgNameDelayEnvApplyPolicy, SysArgValueDelayEnvApplyPolicyApply))
		}
		return name, value
	} else if raw == SysArgNameError {
		if value != SysArgValueOK {
			panic(fmt.Errorf("[Args.SysArgRealname] %s: the value of sys arg '%s' could only be '%s'",
				name, SysArgNameError, SysArgValueOK))
		}
		return name, value
	} else {
		panic(fmt.Errorf("[Args.SysArgRealname] %s: only sys args could have '%s' prefix, '%s' is not sys arg",
			name, sysArgPrefix, raw))
	}
	return //nolint:unreachable
}

func (self SysArgVals) GetDelayStr() string {
	return self[SysArgNameDelay]
}

func (self SysArgVals) IsDelay() bool {
	return len(self[SysArgNameDelay]) != 0
}

func (self SysArgVals) GetDelayDuration() time.Duration {
	delayDur := self[SysArgNameDelay]
	dur, err := time.ParseDuration(delayDur)
	if err != nil {
		panic(&ArgValErrWrongType{
			fmt.Sprintf("[Cmd.AsyncExecute] sys arg '%s = %s' is valid not golang duration format", SysArgNameDelay, delayDur),
			SysArgNameDelay, delayDur, "golan duration format", err,
		})
	}
	return dur
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
