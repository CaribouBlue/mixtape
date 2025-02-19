package featureflag

import (
	_ "embed"
	"encoding/json"
)

//go:embed feature-flags.json
var featureFlagsJson string

type featureFlags map[string]FeatureFlagEntity

type JsonFeatureFlagRepository struct {
	FeatureFlags featureFlags `json:"featureFlags"`
}

func NewJsonFeatureFlagRepository() *JsonFeatureFlagRepository {
	featureFlags := make(map[string]FeatureFlagEntity)
	json.Unmarshal([]byte(featureFlagsJson), &featureFlags)

	return &JsonFeatureFlagRepository{
		FeatureFlags: featureFlags,
	}
}

func (r *JsonFeatureFlagRepository) GetAllFeatureFlags() ([]*FeatureFlagEntity, error) {
	featureFlagEntities := make([]*FeatureFlagEntity, 0, len(r.FeatureFlags))
	for _, featureFlag := range r.FeatureFlags {
		featureFlagEntities = append(featureFlagEntities, &featureFlag)
	}
	return featureFlagEntities, nil
}

func (r *JsonFeatureFlagRepository) GetFeatureFlag(key string) (*FeatureFlagEntity, error) {
	featureFlag, ok := r.FeatureFlags[key]
	if !ok {
		return nil, nil
	}
	return &featureFlag, nil
}
