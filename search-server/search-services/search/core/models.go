package core

type ImageInformation struct {
	ID  int
	Url string
}

type WordInformation struct {
	Word string
	ID   int
	Url  string
}

type QuantityComics struct {
	ImageInfo ImageInformation
	Total     int
}
