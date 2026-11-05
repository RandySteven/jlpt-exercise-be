package configs

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestCacheConfigLoadsFromNestedConfigs(t *testing.T) {
	raw := []byte(`
configs:
  cacheConfig:
    exercise:
      ttl_duration: 600
      ttl_unit: seconds
    subscription_plan_detail:
      ttl_duration: 3600
      ttl_unit: seconds
`)

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}

	exercise := cfg.Configs.CacheConfig["exercise"]
	if exercise.TTLDuration != 600 || exercise.TTLUnit != "seconds" {
		t.Fatalf("exercise config = %+v", exercise)
	}
	detail := cfg.Configs.CacheConfig["subscription_plan_detail"]
	if detail.TTLDuration != 3600 || detail.TTLUnit != "seconds" {
		t.Fatalf("subscription plan detail config = %+v", detail)
	}
}
