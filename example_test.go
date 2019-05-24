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
	w, err := rotate.NewWriter(dir, "test.log", rotate.WithSizeBasedConfig(bytes3))
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
