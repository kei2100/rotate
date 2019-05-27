# rotate

[![CircleCI](https://circleci.com/gh/kei2100/rotate.svg?style=svg)](https://circleci.com/gh/kei2100/rotate)
[![Build status](https://ci.appveyor.com/api/projects/status/9fax0djsm5le725j/branch/master?svg=true)](https://ci.appveyor.com/project/kei2100/rotate/branch/master)

A rotating file Writer

```go
func ExampleWriter() {
	dir, err := ioutil.TempDir("", "rotate-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	const bytes3 int64 = 3
	w, err := rotate.NewWriter(dir, "test.log", rotate.WithSizeBasedPolicy(bytes3))
	if err != nil {
		panic(err)
	}
	defer w.Close()

	fmt.Fprint(w, "1")
	fmt.Fprint(w, "2")
	fmt.Fprint(w, "3")
	time.Sleep(time.Second) // wait for rotate
	fmt.Fprint(w, "4")

	b0, _ := ioutil.ReadFile(filepath.Join(dir, "test.log"))
	b1, _ := ioutil.ReadFile(filepath.Join(dir, "test.log.1"))
	fmt.Printf("%s/%s", b0, b1)

	// Output: 4/123
}
```
