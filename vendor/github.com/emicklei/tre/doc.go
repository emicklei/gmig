/*
Package tre has the TracingError type to collect stack information when an error is caught.

It is inspired the minio probe package ; this one is leaner and has no external dependencies.

	func main() {
		err := doSomething("demo")
		println(tre.New(err, "doSomething failed").Error())
	}

	func doSomething(with string) error {
		err := doAnotherThingThatCanFail(with)
		return tre.New(err, "doAnotherThingThatCanFail failed", "with", with) // pass error, message and context
	}

	func doAnotherThingThatCanFail(with string) error {
		return errors.New("something bad happened")
	}

The TracingError Error() function returns a verbose output of stack information including file,line,function,message and custom key,values.

	main.go:11 main.main:doSomething failed
	main.go:16 main.doSomething:doAnotherThingThatCanFail failed with=demo
	something bad happened

(c) 2016, http://ernestmicklei.com. MIT License
*/
package tre
