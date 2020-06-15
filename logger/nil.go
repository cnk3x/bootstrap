package logger

var Nil null

type null struct {
}

func (n null) Prefix(string) Logger {
	return null{}
}

func (n null) Debugf(string, ...interface{}) {

}

func (n null) Infof(string, ...interface{}) {

}

func (n null) Errorf(string, ...interface{}) {

}

func (n null) Fatalf(string, ...interface{}) {

}

func (n null) Printf(string, ...interface{}) {

}
