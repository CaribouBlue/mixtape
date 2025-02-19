package featureflag

import "errors"

type FeatureFlagMode string

const (
	FeatureFlagModeOn    FeatureFlagMode = "on"
	FeatureFlagModeOff   FeatureFlagMode = "off"
	FeatureFlagModeAllow FeatureFlagMode = "allow"
	FeatureFlagModeDeny  FeatureFlagMode = "deny"
)

var (
	ErrUnsupportedFeatureFlagMode = errors.New("unsupported feature flag mode")
)

type FeatureFlagEntity struct {
	Key      string          `json:"key"`
	Mode     FeatureFlagMode `json:"mode"`
	UserList []int64         `json:"userList"`
}

type FeatureFlagRepository interface {
	GetAllFeatureFlags() ([]*FeatureFlagEntity, error)
	GetFeatureFlag(key string) (*FeatureFlagEntity, error)
}

type FeatureFlagService struct {
	featureFlagRepository FeatureFlagRepository
}

func NewFeatureFlagService(featureFlagRepository FeatureFlagRepository) *FeatureFlagService {
	return &FeatureFlagService{
		featureFlagRepository: featureFlagRepository,
	}
}

func (f *FeatureFlagService) IsFeatureEnabled(key string, userId int64) (bool, error) {
	flag, err := f.featureFlagRepository.GetFeatureFlag(key)
	if err != nil {
		return false, err
	}

	if flag.Mode == FeatureFlagModeOn {
		return true, nil
	}

	switch flag.Mode {
	case FeatureFlagModeOn:
		return true, nil
	case FeatureFlagModeOff:
		return false, nil
	case FeatureFlagModeAllow:
	case FeatureFlagModeDeny:
		for _, listedUserId := range flag.UserList {
			if listedUserId == userId {
				return flag.Mode == FeatureFlagModeAllow, nil
			}
		}
		return false, nil
	}

	return false, ErrUnsupportedFeatureFlagMode
}
