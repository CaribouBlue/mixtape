package model

type Model interface {
	SetId(int64)
}

type OptsFn func(model Model) Model

func WithId(id int64) OptsFn {
	return func(model Model) Model {
		model.SetId(id)
		return model
	}
}
