package toggle

import (
	"fmt"
	"os"
	"testing"

	"github.com/gomatbase/go-env"
	err "github.com/gomatbase/go-error"
)

func TestAdd(t *testing.T) {
	if e := Add("TestAdd", func() error {
		return nil
	}); e == nil {
		t.Error("Allowed to add toggleable with less than 2 options")
	}

	if e := Add("TestAdd", func() error {
		return nil
	}, func() error {
		return nil
	}); e != nil {
		t.Error("Failed to add toggleable ", e)
	}

	if e := Add("TestAdd", func() error {
		return nil
	}, func() error {
		return nil
	}); e == nil {
		t.Error("Allowed overwriting a toggleable ", e)
	}
}

func TestExecuteAndToggle(t *testing.T) {
	if e := Execute("TestExecute"); e != ToggleableNotFoundError {
		t.Error("Execution of non-existing toggleable should have failed")
	}
	if e := Toggle("TestExecute", 4); e != nil {
		t.Error("Failed to set preemptive toggle value")
	}
	result1 := err.Error("1")
	result2 := err.Error("2")
	result3 := err.Error("3")
	_ = Add("TestExecute", func() error {
		return result1
	}, func() error {
		return result2
	}, func() error {
		return result3
	})

	if Execute("TestExecute") != result1 {
		t.Error("Unexpected value for toggle 1")
	}
	if e := Toggle("TestExecute", 0); e != nil {
		t.Error("Failed to set toggle value 0")
	}
	if Execute("TestExecute") != result1 {
		t.Error("Unexpected value for toggle 1")
	}
	if e := Toggle("TestExecute", 1); e != nil {
		t.Error("Failed to set toggle value 1")
	}
	if Execute("TestExecute") != result2 {
		t.Error("Unexpected value for toggle 2")
	}
	if e := Toggle("TestExecute", 2); e != nil {
		t.Error("Failed to set toggle value 2")
	}
	if Execute("TestExecute") != result3 {
		t.Error("Unexpected value for toggle 3")
	}
	if e := Toggle("TestExecute", 3); e != InvalidToggleError {
		t.Error("Failed to raise out of bounds error for invalid toggle value")
	}
	if Execute("TestExecute") != result3 {
		t.Error("Expected toggleable to run unaltered after an invalid toggle call")
	}

	if e := Toggle("TestExecute2", 1); e != nil {
		t.Error("Failed to set preemptive toggle value")
	}
	_ = Add("TestExecute2", func() error {
		return result1
	}, func() error {
		return result2
	})
	if Execute("TestExecute2") != result2 {
		t.Error("Unexpected value for toggle 2")
	}

}

func ExampleRun() {
	os.Args = []string{"app", "-TExampleRun", "1"}
	env.Load() // For testing purposes the environment has to be reloaded after updating the cml arguments
	_ = Add("ExampleRun", func() error {
		fmt.Println("1")
		return nil
	}, func() error {
		fmt.Println("2")
		return nil
	})
	_ = Execute("ExampleRun")
	_ = Toggle("ExampleRun", 0)
	_ = Execute("ExampleRun")
	_ = Run("ExampleRun", func() error {
		fmt.Println("2")
		return nil
	}, func() error {
		fmt.Println("1")
		return nil
	})

	// Output:
	// 2
	// 1
	// 2
}
