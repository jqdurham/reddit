package reddit

type Listing struct {
	Segment `json:"data"`
}

type Segment struct {
	After    string   `json:"after"`
	Children Children `json:"children,omitempty"`
}

type Children []struct {
	Post Post `json:"data,omitempty"`
}

type Post struct {
	Name   string `json:"name"`
	Title  string `json:"title"`
	Ups    int    `json:"ups"`
	Author string `json:"author"`
}
