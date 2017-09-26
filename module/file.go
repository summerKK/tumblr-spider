package module

import "path"

var (
	gfyRequest = "https://gfycat.com/cajax/get/%s"
)

type File struct {
	User          *User
	URL           string
	UnixTimestamp int64
	Filename      string
}

func NewFile(URL string) File {
	return File{
		URL:      URL,
		Filename: path.Base(URL),
	}
}
