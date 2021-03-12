package rotate_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/kei2100/rotate"
)

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

func ExampleWriter_timeBased() {
	dir, err := ioutil.TempDir("", "rotate-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	w, err := rotate.NewWriter(dir, "test.log", rotate.WithTimeBasedPolicy(func(openedAtUnix int64) bool {
		opendeAt := time.Unix(openedAtUnix, 0)
		now := time.Now()
		return now.Second()-opendeAt.Second() > 0 // rotate if 1 sec has passed
	}))
	if err != nil {
		panic(err)
	}
	defer w.Close()

	fmt.Fprint(w, "1")
	time.Sleep(1 * time.Second)
	fmt.Fprint(w, "2") // rotate will start after this writing
	time.Sleep(1 * time.Second)
	fmt.Fprint(w, "3") // rotate will start after this writing
	time.Sleep(1 * time.Second)
	fmt.Fprint(w, "4")          // rotate will start after this writing
	time.Sleep(1 * time.Second) // wait for rotate

	b0, _ := ioutil.ReadFile(filepath.Join(dir, "test.log"))
	b1, _ := ioutil.ReadFile(filepath.Join(dir, "test.log.1"))
	b2, _ := ioutil.ReadFile(filepath.Join(dir, "test.log.2"))
	b3, _ := ioutil.ReadFile(filepath.Join(dir, "test.log.3"))
	fmt.Printf("%s/%s/%s/%s", b0, b1, b2, b3)

	// Output: /4/3/12
}
