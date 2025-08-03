package udl

type Descriptor struct {
	Title string
	Link  string
}

type RunOption struct {
	Verbose bool
}

type ISite interface {
	Name() string
	Run(RunOption) error
	ListSeasons(Descriptor) ([]Descriptor, error)
	ListEpisodes(Descriptor) ([]Descriptor, error)
	ListQuality(Descriptor) ([]Descriptor, error)
	Download(Descriptor) error
	BulkDownload([]Descriptor) error
}

type BaseSite struct{}

func (BaseSite) Name() string {
	return "BaseSite"
}

func (BaseSite) Run(RunOption) error {
	return nil
}

func (BaseSite) ListSeasons(Descriptor) ([]Descriptor, error) {
	return []Descriptor{}, nil
}

func (BaseSite) ListEpisodes(Descriptor) ([]Descriptor, error) {
	return []Descriptor{}, nil
}

func (BaseSite) ListQuality(Descriptor) ([]Descriptor, error) {
	return []Descriptor{}, nil
}

func (BaseSite) Download(Descriptor) error {
	return nil
}

func (BaseSite) BulkDownload([]Descriptor) error {
	return nil
}
