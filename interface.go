package gobloom

type Interface interface {
	Add([]byte)
	AddString(string)
	Test([]byte) bool
	TestString(string) bool
}
