package testext

// Sequence is a convenient way to capture a series of values in a specified order that you can use
// to determine if code was fired in a specific sequence/order.
type Sequence struct {
	values []string
}

// Append writes the next value for the piece of code that executed.
func (seq *Sequence) Append(value string) {
	seq.values = append(seq.values, value)
}

// Value returns the value at the specific index. If you haven't appended that much yet, then this
// will return "".
func (seq *Sequence) Value(index int) string {
	if index >= len(seq.values) {
		return ""
	}
	return seq.values[index]
}

// Values returns all the values that you collected during the test case.
func (seq *Sequence) Values() []string {
	return seq.values
}

// Reset erases all current values in the sequence, allowing you to re-use this sequence multiple
// times within the same test case.
func (seq *Sequence) Reset() {
	seq.values = nil
}
