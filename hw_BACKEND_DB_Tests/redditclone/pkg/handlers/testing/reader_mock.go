package handlerstestsutils

import "fmt"

type MockReader struct {
	A int
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("something gone wrong")
}
