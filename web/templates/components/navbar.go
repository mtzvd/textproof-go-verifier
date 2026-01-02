package components

type NavBarItem struct {
	Label  string
	Href   string
	Icon   string
	Active bool
	Align  string
}

type NavBar struct {
	Brand string
	Icon  string
	Items []NavBarItem
}
