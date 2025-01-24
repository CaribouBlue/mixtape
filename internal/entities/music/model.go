package music

type Artist struct {
	Id   string
	Name string
}

type Track struct {
	Id      string
	Name    string
	Artists []Artist
	Url     string
}

type Playlist struct {
	Id   string
	Name string
	Url  string
}
