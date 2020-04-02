package verify

// FloatVerify are float64 verifiers.
type FloatVerify func(float64) error

// Float verifies float64s.
func Float(i float64, verifiers ...FloatVerify) error {
	for _, v := range verifiers {
		if err := v(i); err != nil {
			return err
		}
	}
	return nil
}

// FloatPositive verifies that an float64 is positive.
func FloatPositive(i float64) error {
	return Float(i, FloatMin(1))
}

// FloatRange verifies that an float64 is in range.
func FloatRange(low, high float64) FloatVerify {
	return func(i float64) error {
		return Float(i, FloatMin(low), FloatMax(high))
	}
}

// FloatMin verifies that the float64 is at least l.
func FloatMin(l float64) FloatVerify {
	return func(i float64) error {
		if i < l {
			return Errorf("must be at least %v", l)
		}
		return nil
	}
}

// FloatMax verifies that the float64 is at most l.
func FloatMax(l float64) FloatVerify {
	return func(i float64) error {
		if i > l {
			return Errorf("must be at most %v", l)
		}
		return nil
	}
}
