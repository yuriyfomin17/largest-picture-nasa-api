package domain

type Picture struct {
	sol  int
	size int
	url  string
}

type NewPictureData struct {
	Size int
	Sol  int
	Url  string
}

// NewPicture Constructor for Picture struct
func NewPicture(pic NewPictureData) Picture {
	return Picture{
		sol:  pic.Sol,
		size: pic.Size,
		url:  pic.Url,
	}
}

// GetSol Getter for Sol field
func (p Picture) GetSol() int {
	return p.sol
}

// GetUrl Getter for Url field
func (p Picture) GetUrl() string {
	return p.url
}

func (p Picture) GetSize() int {
	return p.size
}
