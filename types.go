package udl

type ISite interface {
	Name() string
	Run(RunOption) error
}

type RunOption struct{}

type Descriptor struct {
	Title string
	Link  string
}
