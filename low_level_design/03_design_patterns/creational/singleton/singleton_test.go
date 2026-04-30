package singleton

import (
	"sync"
	"testing"
)

func TestGetConfig_ReturnsSameInstance(t *testing.T) {
	ResetConfig()

	c1 := GetConfig()
	c2 := GetConfig()

	if c1 != c2 {
		t.Error("GetConfig() returned different instances")
	}
}

func TestGetConfig_Defaults(t *testing.T) {
	ResetConfig()

	c := GetConfig()
	if c.AppName != "LLD-App" {
		t.Errorf("AppName = %q", c.AppName)
	}
	if c.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d", c.MaxRetries)
	}
}

func TestConfig_SharedState(t *testing.T) {
	ResetConfig()

	c1 := GetConfig()
	c1.Set("db_host", "localhost")

	c2 := GetConfig()
	if got := c2.Get("db_host"); got != "localhost" {
		t.Errorf("shared state broken: got %q", got)
	}
}

func TestGetConfig_ThreadSafe(t *testing.T) {
	ResetConfig()

	var wg sync.WaitGroup
	instances := make([]*Config, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			instances[idx] = GetConfig()
		}(i)
	}
	wg.Wait()

	first := instances[0]
	for i, c := range instances {
		if c != first {
			t.Errorf("instance[%d] differs from instance[0]", i)
		}
	}
}
