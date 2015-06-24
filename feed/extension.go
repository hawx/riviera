package feed

type Extension struct {
	Name      string
	Value     string
	Attrs     map[string]string
	Childrens map[string][]Extension
}
