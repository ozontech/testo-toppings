package parallel

// Scope defines to what extent tests should become parallel.
type Scope uint8

// Available scopes.
// Scopes can be combined using bitmasks:
//
//	const All = SuiteTests | Suites | Tests
const (
	// SuiteTests covers suite tests.
	//
	// For example:
	//
	// 	func (Suite) TestA(t T) { ... }
	// 	func (Suite) TestB(t T) { ... }
	//
	// Tests A & B will be marked as parallel.
	//
	// This is a default value.
	SuiteTests Scope = 1 << iota

	// Suites covers suites but not their tests.
	//
	// For example:
	//
	// 	func Test(t *testing.T) {
	//		testo.RunSuite(t, new(Suite))
	//		testo.RunSuite(t, new(OtherSuite))
	// 	}
	//
	// Both of these suites will be run in parallel.
	Suites

	// Tests covers native tests.
	//
	// For example:
	//
	// 	func TestA(t *testing.T) {
	//		testo.RunSuite(t, new(Suite))
	// 	}
	//
	// 	func TestA(t *testing.T) {
	//		testo.RunSuite(t, new(OtherSuite))
	// 	}
	//
	// Tests A & B will be run in parallel.
	Tests
)

func (s Scope) has(flag Scope) bool {
	return s&flag != 0
}
