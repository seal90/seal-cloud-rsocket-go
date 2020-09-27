package support

type AsyncPredicate interface {
	Apply(inter interface{}) bool
}
